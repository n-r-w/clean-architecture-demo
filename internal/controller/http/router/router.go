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
	httpcontroller "github.com/n-r-w/log-server-v2/internal/controller/http"
	"github.com/n-r-w/log-server-v2/internal/controller/http/rest"
	"github.com/n-r-w/log-server-v2/internal/usecase/usecase"
	"github.com/n-r-w/log-server-v2/pkg/logger"
	"github.com/n-r-w/log-server-v2/pkg/tools"
	"golang.org/x/exp/slices"
)

// Тип для описания ключевых значений параметров, добавляемых в контекст запроса
// в процессе его обработки через middleware
type contextKey string

const (
	// SessionName Ключ для хранения информации о сессии со стороны пользователя
	sessionName = "logserver"
	// UserIDKeyName Ключ для хранения id пользователя в сессии (в куках)
	userIDKeyName = "user_id"

	ctxKeyRequestID = contextKey("request-id")
)

// Router - реализует интерфейс httpcontroller.Interface
type Router struct {
	mux          *mux.Router
	sessionStore sessions.Store // Управление сессиями пользователей
	logger       logger.Interface
	user         usecase.User
	log          usecase.Log

	subrouters map[string]*mux.Router
}

func NewRouter(logger logger.Interface, user usecase.User, log usecase.Log, sessionEncriptionKey string, superAdminID uint64, sessionAge uint) *Router {
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
	rest.InitRoutes(r, superAdminID, sessionAge, user, log)

	return r
}

func (router *Router) Handler() http.Handler {
	return router.mux
}

// Ответ с ошибкой
func (router *Router) RespondError(w http.ResponseWriter, code int, err error) {
	router.RespondData(w, code, map[string]string{"error": err.Error()})
}

// Ответ на запрос без сжатия
func (router *Router) RespondData(w http.ResponseWriter, code int, data interface{}) {
	if code > 0 {
		w.WriteHeader(code)
	}
	if data != nil {
		switch d := data.(type) {
		case string:
			_, _ = w.Write([]byte(d))

			w.Header().Add("Content-Type", "application/octet-stream")
		default:
			if err := json.NewEncoder(w).Encode(data); err != nil {
				w.WriteHeader(http.StatusInternalServerError)
				_, _ = w.Write([]byte(fmt.Sprintf(`{"error": "%v"}`, err)))

				return
			}

			w.Header().Set("Content-Type", "application/json")
		}
	} else {
		_, _ = w.Write([]byte("{}"))
	}
}

// Ответ на запрос со сжатием если его поддерживает клиент
func (router *Router) RespondCompressed(w http.ResponseWriter, r *http.Request, code int, ctype httpcontroller.CompressionType, data interface{}) {
	if data == nil {
		router.RespondData(w, code, data)

		return
	}

	// проверяем хочет ли клиент сжатие
	accepted := strings.Split(r.Header.Get("Accept-Encoding"), ",")
	gzipCompression := slices.Contains(accepted, "gzip")
	deflateCompression := !gzipCompression && slices.Contains(accepted, "deflate")

	if !gzipCompression && !deflateCompression {
		router.RespondData(w, code, data)

		return
	}

	// заполняем буфер для сжатия
	var sourceData []byte
	switch d := data.(type) {
	case string:
		sourceData = []byte(d)
	default:
		var errJSON error
		sourceData, errJSON = json.Marshal(data)

		if errJSON != nil {
			router.RespondError(w, http.StatusInternalServerError, errJSON)
		}

		w.Header().Set("Content-Type", "application/json")
	}

	if deflateCompression {
		w.Header().Set("Content-Encoding", "deflate")
	} else {
		w.Header().Set("Content-Encoding", "gzip")
	}

	compressedData, err := tools.CompressData(deflateCompression, sourceData)

	if err != nil {
		router.RespondError(w, http.StatusInternalServerError, err)

		return
	}

	w.WriteHeader(code)
	_, _ = w.Write(compressedData)
}

func (router *Router) AddRoute(subroute string, route string, handler http.HandlerFunc, methods ...string) {
	var r *mux.Router
	if len(subroute) == 0 {
		r = router.mux
	} else {
		r = router.getSubrouter(subroute)
	}

	r.HandleFunc(route, handler).Methods(methods...)
}

func (router *Router) AddMiddleware(subroute string, mwf ...httpcontroller.MiddlewareFunc) {
	var funcs []mux.MiddlewareFunc
	for _, f := range mwf {
		funcs = append(funcs, func(h http.Handler) http.Handler { return f(h) })
	}

	if len(subroute) == 0 {
		router.mux.Use(funcs...)
	} else {
		router.getSubrouter(subroute).Use(funcs...)
	}
}

func (router *Router) StartSession(w http.ResponseWriter, r *http.Request, userID uint64, sessionAge uint) error {
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
	// извлекаем из запроса пользователя куки с инфорамацией о сессии
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
