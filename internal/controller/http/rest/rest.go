// Package rest ...
package rest

import (
	"errors"
	"net/http"

	httpcontroller "github.com/n-r-w/log-server-v2/internal/controller/http"
	"github.com/n-r-w/log-server-v2/internal/entity"
	"github.com/n-r-w/log-server-v2/internal/usecase/usecase"
)

var (
	errNotAuthenticated = errors.New("not authenticated")
	errNotAdmin         = errors.New("not admin")
)

// задаем свой тип, чтобы была возможность отличить что лежит в переменной any
type ctxKey string

const (
	// Ключ для хранения модели пользователя в контексте запроса после успешной аунтетификации
	ctxKeyUser = ctxKey("rest-user")

	// Имя хедера REST запроса, в котором клиент указывает в каком виде он желает получить ответ
	binaryFormatHeaderName = "binary-format"
	// Требуется ответ в формате protobuf
	binaryFormatHeaderProtobuf = "protobuf"
)

type restInfo struct {
	controller   httpcontroller.Interface
	user         usecase.User
	log          usecase.Log
	superAdminID uint64
	sessionAge   uint
}

// InitRoutes Инициализация маршрутов
func InitRoutes(controller httpcontroller.Interface, superAdminID uint64, sessionAge uint, user usecase.User, log usecase.Log) {
	i := &restInfo{
		controller:   controller,
		user:         user,
		log:          log,
		superAdminID: superAdminID,
		sessionAge:   sessionAge,
	}

	// логин
	controller.AddRoute("/api/auth", "/login", i.handleSessionsCreate(), "POST")
	// закрытие сессии
	controller.AddRoute("/api/auth", "/close", i.closeSession(), "DELETE")

	// устанавливаем middleware для проверки валидности сессии
	controller.AddMiddleware("/api/private", i.authenticateUser)

	// запрос с информацией о текущей сессии
	controller.AddRoute("/api/private", "/whoami", i.handleWhoami(), "GET")
	// добавить пользователя
	controller.AddRoute("/api/private", "/add-user", i.addUser(), "POST")
	// сменить пароль
	controller.AddRoute("/api/private", "/change", i.changePassword(), "PUT")
	// получить список пользователей
	controller.AddRoute("/api/private", "/users", i.getUsers(), "GET")
	// добавить запись в лог
	controller.AddRoute("/api/private", "/add-log", i.addLogRecord(), "POST")
	// получить список записей из лога. Ответ в gzip формате
	controller.AddRoute("/api/private", "/records", i.getLogRecords(), "GET")
}

// Текущий пользователь. Он помещается в контекст в методе authenticateUser
func currentUser(r *http.Request) *entity.User {
	user, ok := r.Context().Value(ctxKeyUser).(*entity.User)
	if ok {
		return user
	}

	return nil
}
