package rest

import (
	"context"
	"encoding/json"
	"net/http"

	"github.com/n-r-w/log-server-v2/internal/domain/entity"
)

// Логин (создание сессии)
func (info *restInfo) handleSessionsCreate() http.HandlerFunc {
	type request struct {
		Login    string `json:"login"`
		Password string `json:"password"`
	}

	return func(w http.ResponseWriter, r *http.Request) {
		loginData := &request{
			Login:    "",
			Password: "",
		}
		// парсим входящий json
		if err := json.NewDecoder(r.Body).Decode(loginData); err != nil {
			info.controller.RespondError(w, http.StatusBadRequest, err)

			return
		}
		// ищем в БД по логину
		ID, err := info.user.CheckPassword(loginData.Login, loginData.Password)
		if err != nil {
			info.controller.RespondError(w, http.StatusForbidden, err)

			return
		}
		// получаем сесиию
		if err = info.controller.StartSession(w, r, ID, info.sessionAge); err != nil {
			info.controller.RespondError(w, http.StatusForbidden, err)

			return
		}

		info.controller.RespondData(w, http.StatusOK, nil)
	}
}

// Аутентификация пользователя на основании ранее прошедшего логина
func (info *restInfo) authenticateUser(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ID, err := info.controller.CheckSession(r)
		if err != nil {
			info.controller.RespondError(w, http.StatusUnauthorized, err)

			return
		}

		// берем инфу о пользователе из БД
		var user entity.User
		user, err = info.user.FindByID(ID)
		if err != nil {
			info.controller.RespondError(w, http.StatusInternalServerError, err)

			return
		}

		// добавляем модель пользователя в контекст запроса
		next.ServeHTTP(w, r.WithContext(context.WithValue(r.Context(), ctxKeyUser, &user)))
	})
}

// Обработчик запроса с информацией о текущей сессии
func (info *restInfo) handleWhoami() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		info.controller.RespondData(w, http.StatusOK,
			// объект "пользователь" кладется в контекст при логине
			currentUser(r))
	}
}

// Обработчик запроса закрытия сессии
func (info *restInfo) closeSession() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		info.controller.CloseSession(w, r)
		info.controller.RespondData(w, http.StatusOK, nil)
	}
}
