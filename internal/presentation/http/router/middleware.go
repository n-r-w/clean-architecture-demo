package router

import (
	"context"
	"net/http"

	"github.com/google/uuid"
)

// Реализует интерфейс http.ResponseWriter
// Подменяет собой стандартный http.ResponseWriter и позволяет дополнительно сохранить в нем ошибку
type responseWriterEx struct {
	http.ResponseWriter
	code int
	err  error
}

func (w *responseWriterEx) WriteHeader(statusCode int) {
	w.code = statusCode
	w.ResponseWriter.WriteHeader(statusCode)
}

// Добавляем к контексту уникальный ID сесии с ключом ctxKeyRequestID
func (router *Router) setRequestID(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		id := uuid.New().String()
		w.Header().Set("X-Request-ID", id)
		next.ServeHTTP(w, r.WithContext(context.WithValue(r.Context(), ctxKeyRequestID, id)))
	})
}

// Выводим все запросы в журнал
func (router *Router) logRequest(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		/*
			// пишем инфу о начале обработки запроса
			router.logger.Info("addr: %s, id: %s, started %s %s",
				r.RemoteAddr,
				r.Context().Value(ctxKeyRequestID),
				r.Method,
				r.RequestURI)

			start := time.Now() */
		rw := &responseWriterEx{
			ResponseWriter: w,
			code:           http.StatusOK,
			err:            nil,
		}

		// вызываем обработчик нижнего уровня
		next.ServeHTTP(rw, r)
		/*
			// выводим в журнал результат
			var level logger.MessageLevel
			switch {
			case rw.code >= http.StatusInternalServerError:
				level = logger.ErrorLevel
			case rw.code >= http.StatusBadRequest:
				level = logger.WarnLevel
			default:
				level = logger.InfoLevel
			}

			var errorText string
			if rw.err != nil {
				errorText = rw.err.Error()
				errorText = strings.ReplaceAll(errorText, `"`, "")
			} else {
				errorText = "-"
			}

			router.logger.Level(level, "addr: %s, id: %s, completed with %d %s in %v, info: %s",
				r.RemoteAddr,
				r.Context().Value(ctxKeyRequestID),
				rw.code,
				http.StatusText(rw.code),
				time.Since(start),
				errorText) */
	})
}
