// Package handler ...
package handler

import (
	"net/http"
	"time"

	"github.com/n-r-w/log-server-v2/internal/domain/entity"
)

type CompressionType int

const (
	CompressionNo = CompressionType(iota)
	CompressionGzip
	CompressionDeflate
)

type MiddlewareFunc func(next http.Handler) http.Handler

// RouterInterface - интерфейс http роутера
// Создан для исключения зависимости обработчиков запросов от используемого роутера
type RouterInterface interface {
	// RespondData - ответ на запрос
	// data содержит []byte или указатель на объект. Во втором случае этот объект преобразуется в JSON */
	RespondData(w http.ResponseWriter, code int, data interface{})
	// RespondCompressed - ответ на запрос
	// data содержит []byte или указатель на объект. Во втором случае этот объект преобразуется в JSON.
	// Дополнительно проверяет заголовок запроса на "Accept-Encoding" и решает сжимать ли ответ на самом деле,
	// т.е. в итоге ответ может быть и без сжатия
	RespondCompressed(w http.ResponseWriter, r *http.Request, code int, ctype CompressionType, data interface{})
	// RespondError - возврат ошибки
	RespondError(w http.ResponseWriter, code int, err error)

	// AddRoute - добавить обработчик
	AddRoute(subroute string, route string, handler http.HandlerFunc, methods ...string)
	// AddMiddleware - добавить цепочку обработчиков на промежуточном уровне
	AddMiddleware(subroute string, mwf ...MiddlewareFunc)

	// StartSession - запомнить новую сессию после логина
	StartSession(w http.ResponseWriter, r *http.Request, userID uint64, sessionAge uint) error
	// CheckSession - проверить залогинен ли пользователь
	CheckSession(*http.Request) (userID uint64, err error)
	// CloseSession - закрыть сессию
	CloseSession(w http.ResponseWriter, r *http.Request)
}

// Интерфейсы по работе с доменом (юскейсами). Реализуются в каталоге domain/usecase
type (
	// UserInterface интерфейс, реализуемый юскейсом работы с пользователями
	UserInterface interface {
		// CheckPassword Проверить пароль
		CheckPassword(login string, password string) (ID uint64, err error)
		// ChangePassword Сменить пароль
		ChangePassword(currentUser entity.User, login string, password string) (ID uint64, err error)

		Insert(user entity.User) error
		Remove(id uint64) error
		Update(user entity.User) error

		FindByID(id uint64) (entity.User, error)
		FindByLogin(login string) (entity.User, error)
		GetUsers() ([]entity.User, error)
	}

	// LogInterface интерфейс, реализуемый юскейсом работы с логами
	LogInterface interface {
		Insert(logs []entity.LogRecord) error

		Find(dateFrom time.Time, dateTo time.Time, limit uint) (records []entity.LogRecord, limited bool, err error)
	}
)
