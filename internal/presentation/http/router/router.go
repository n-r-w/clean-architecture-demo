// Package router ...
package router

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strings"

	"github.com/gorilla/handlers"
	"github.com/gorilla/mux"
	"github.com/gorilla/sessions"
	"github.com/n-r-w/log-server-v2/internal/presentation/http/handler"
	"github.com/n-r-w/log-server-v2/internal/presentation/http/handler/rest"
	"github.com/n-r-w/log-server-v2/pkg/logger"
	"github.com/n-r-w/log-server-v2/pkg/tools"
	"golang.org/x/exp/slices"
)

// Тип для описания ключевых значений параметров, добавляемых в контекст запроса
// в процессе его обработки через middleware
type contextKey string

const (
	// Ключ для хранения информации о сессии со стороны пользователя
	sessionName = "logserver"
	// Ключ для хранения id пользователя в сессии (в куках)
	userIDKeyName = "user_id"
	// Ключ для хранения номера сессии в контексте запроса
	ctxKeyRequestID = contextKey("request-id")
)

// Router - реализует интерфейс handler.RouterInterface
type Router struct {
	mux          *mux.Router
	sessionStore sessions.Store // Управление сессиями пользователей
	logger       logger.Interface
	user         handler.UserInterface
	log          handler.LogInterface

	subrouters map[string]*mux.Router
}

func NewRouter(logger logger.Interface, user handler.UserInterface, log handler.LogInterface, sessionEncriptionKey string, superAdminID uint64, sessionAge int, maxLogRecordsResult int) *Router {
	r := &Router{
		mux:          mux.NewRouter(),
		sessionStore: sessions.NewCookieStore([]byte(sessionEncriptionKey)),
		logger:       logger,
		user:         user,
		log:          log,
		subrouters:   make(map[string]*mux.Router),
	}

	// подмешивание номера сессии
	r.mux.Use(r.setRequestID)
	// журналирование запросов
	r.mux.Use(r.logRequest)

	// разрешаем запросы к серверу c любых доменов (cross-origin resource sharing)
	r.mux.Use(handlers.CORS(handlers.AllowedOrigins([]string{"*"})))

	// создаем маршруты для rest
	rest.InitRoutes(r, user, log, superAdminID, sessionAge, maxLogRecordsResult)

	return r
}

func (router *Router) Handler() http.Handler {
	return router.mux
}

// RespondError Ответ с ошибкой
func (router *Router) RespondError(w http.ResponseWriter, code int, err error) {
	rw, ok := w.(*responseWriterEx)
	if !ok {
		panic("internal error")
	}

	rw.err = err

	router.RespondData(rw, code, map[string]string{"error": err.Error()})
}

// RespondData Ответ на запрос без сжатия
func (router *Router) RespondData(w http.ResponseWriter, code int, data interface{}) {
	rw, ok := w.(*responseWriterEx)
	if !ok {
		panic("internal error")
	}

	if code > 0 {
		rw.WriteHeader(code)
	}
	if data != nil {
		switch d := data.(type) {
		case string:
			_, _ = rw.Write([]byte(d))

			rw.Header().Add("Content-Type", "application/octet-stream")
		default:
			if err := json.NewEncoder(rw).Encode(data); err != nil {
				rw.WriteHeader(http.StatusInternalServerError)
				_, _ = rw.Write([]byte(fmt.Sprintf(`{"error": "%v"}`, err)))

				return
			}

			rw.Header().Set("Content-Type", "application/json")
		}
	} else {
		_, _ = rw.Write([]byte("{}"))
	}
}

// RespondCompressed Ответ на запрос со сжатием если его поддерживает клиент
func (router *Router) RespondCompressed(w http.ResponseWriter, r *http.Request, code int, ctype handler.CompressionType, data interface{}) {
	if data == nil {
		router.RespondData(w, code, data)

		return
	}

	// проверяем хочет ли клиент сжатие
	compressionType := handler.CompressionNo

	accepted := strings.Split(r.Header.Get("Accept-Encoding"), ",")
	if slices.Contains(accepted, "gzip") && ctype == handler.CompressionGzip {
		compressionType = handler.CompressionGzip
	} else if slices.Contains(accepted, "deflate") && ctype == handler.CompressionDeflate {
		compressionType = handler.CompressionDeflate
	}

	if compressionType == handler.CompressionNo {
		router.RespondData(w, code, data)

		return
	}

	// заполняем буфер для сжатия
	var sourceData []byte
	switch d := data.(type) {
	case string:
	case []byte:
		sourceData = []byte(d)
	default:
		var errJSON error
		sourceData, errJSON = json.Marshal(data)

		if errJSON != nil {
			router.RespondError(w, http.StatusInternalServerError, errJSON)
		}

		w.Header().Set("Content-Type", "application/json")
	}

	if compressionType == handler.CompressionGzip {
		w.Header().Set("Content-Encoding", "gzip")
	} else {
		w.Header().Set("Content-Encoding", "deflate")
	}

	compressedData, err := tools.CompressData(compressionType == handler.CompressionDeflate, sourceData)

	if err != nil {
		router.RespondError(w, http.StatusInternalServerError, err)

		return
	}

	w.WriteHeader(code)
	_, _ = w.Write(compressedData)
}

// AddRoute ...
func (router *Router) AddRoute(subroute string, route string, handler http.HandlerFunc, methods ...string) {
	var r *mux.Router
	if len(subroute) == 0 {
		r = router.mux
	} else {
		r = router.getSubrouter(subroute)
	}

	r.HandleFunc(route, handler).Methods(methods...)
}

// AddMiddleware ...
func (router *Router) AddMiddleware(subroute string, mwf ...handler.MiddlewareFunc) {
	funcs := make([]mux.MiddlewareFunc, len(mwf))
	for i, f := range mwf {
		funcs[i] = func(h http.Handler) http.Handler { return f(h) }
	}

	if len(subroute) == 0 {
		router.mux.Use(funcs...)
	} else {
		router.getSubrouter(subroute).Use(funcs...)
	}
}

// StartSession ...
func (router *Router) StartSession(w http.ResponseWriter, r *http.Request, userID uint64, sessionAge int) error {
	// получаем сесиию
	session, err := router.sessionStore.New(r, sessionName)
	if err != nil {
		return err
	}

	// записываем информацию о том, что пользователь с таким ID залогинился
	session.Values[userIDKeyName] = userID
	session.Options = &sessions.Options{
		Path:   "/",
		Domain: "",
		MaxAge: int(sessionAge),
		Secure: false,
		// HttpOnly: true, // прячем содержимое сессии от доступа через JavaSript в браузере
		HttpOnly: false,
		SameSite: 0,
	}

	return router.sessionStore.Save(r, w, session)
}

func (router *Router) CheckSession(r *http.Request) (userID uint64, err error) {
	// извлекаем из запроса пользователя куки с информацией о сессии
	session, err := router.sessionStore.Get(r, sessionName)
	if err != nil {
		return 0, err
	}

	// ищем в информацию о пользователе в сессиях
	ID, ok := session.Values[userIDKeyName]
	if !ok || session.Options.MaxAge < 0 {
		return 0, errors.New("unauthorized")
	}

	return ID.(uint64), nil
}

func (router *Router) CloseSession(w http.ResponseWriter, r *http.Request) {
	// получаем сесиию
	session, err := router.sessionStore.Get(r, sessionName)
	if err != nil {
		router.logger.Error("session store get error %v", err)

		return
	}
	if session == nil {
		return
	}

	// удаляем из нее данные о логине
	delete(session.Values, userIDKeyName)
	// сохраняем
	if err := router.sessionStore.Save(r, w, session); err != nil {
		router.logger.Error("session save error")
	}
}

func (router *Router) getSubrouter(path string) *mux.Router {
	sr := router.subrouters[path]
	if sr == nil {
		sr = router.mux.PathPrefix(path).Subrouter()
		router.subrouters[path] = sr
	}

	return sr
}
