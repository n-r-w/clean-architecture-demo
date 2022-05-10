package rest

import (
	"encoding/json"
	"errors"
	"net/http"

	"github.com/n-r-w/log-server-v2/internal/domain/entity"
)

// Добавить пользователя
func (info *restInfo) addUser() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		cu := currentUser(r)
		if cu == nil {
			info.controller.RespondError(w, http.StatusInternalServerError, errors.New("no current user"))

			return
		}
		if cu.ID != info.superAdminID {
			info.controller.RespondError(w, http.StatusForbidden, errNotAdmin)

			return
		}

		u := entity.User{
			ID:                0,
			Login:             "",
			Name:              "",
			Password:          "",
			EncryptedPassword: "",
		}
		// парсим входящий json
		if err := json.NewDecoder(r.Body).Decode(&u); err != nil {
			info.controller.RespondError(w, http.StatusBadRequest, err)

			return
		}

		if err := info.user.Insert(u); err != nil {
			info.controller.RespondError(w, http.StatusForbidden, err)
		}

		info.controller.RespondData(w, http.StatusCreated, nil)
	}
}

// Список пользователей
func (info *restInfo) getUsers() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		cu := currentUser(r)
		if cu == nil {
			info.controller.RespondError(w, http.StatusInternalServerError, errNotAuthenticated)

			return
		}
		if cu.ID != info.superAdminID {
			info.controller.RespondError(w, http.StatusForbidden, errNotAdmin)

			return
		}

		users, err := info.user.GetUsers()
		if err != nil {
			info.controller.RespondError(w, http.StatusInternalServerError, err)

			return
		}

		info.controller.RespondData(w, http.StatusOK, &users)
	}
}

// Изменить пароль пользователя
func (info *restInfo) changePassword() http.HandlerFunc {
	type request struct {
		Login    string `json:"login"`
		Password string `json:"password"`
	}

	return func(w http.ResponseWriter, r *http.Request) {
		req := request{
			Login:    "",
			Password: "",
		}
		// парсим входящий json
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			info.controller.RespondError(w, http.StatusBadRequest, err)

			return
		}

		currentUser := currentUser(r)
		if currentUser == nil {
			info.controller.RespondError(w, http.StatusForbidden, errNotAuthenticated)

			return
		}

		_, err := info.user.ChangePassword(*currentUser, req.Login, req.Password)
		if err != nil {
			info.controller.RespondError(w, http.StatusForbidden, err)
		}

		info.controller.RespondData(w, http.StatusOK, nil)
	}
}
