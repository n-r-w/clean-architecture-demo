package httpcontroller

import (
	"net/http"
)

type CompressionType int

const (
	CompressionGzip = CompressionType(iota)
	CompressionDeflate
)

type MiddlewareFunc func(next http.Handler) http.Handler

// Interface - интерфейс http контроллера
// Создан для исключения зависимости обработчиков запросов от используемого роутера
type Interface interface {
	/* RespondData - ответ на запрос
	data содержит []byte или указатель на объект. Во втором случае этот объект преобразуется в JSON */
	RespondData(w http.ResponseWriter, code int, data interface{})
	/* RespondCompressed - ответ на запрос
	data содержит []byte или указатель на объект. Во втором случае этот объект преобразуется в JSON.
	Дополнительно проверяет заголовок запроса на "Accept-Encoding" и решает сжимать ли ответ на самом деле,
	т.е. в итоге ответ может быть и без сжатия */
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
