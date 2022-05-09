// Package entity ...
package entity

import (
	"regexp"
	"strings"

	validation "github.com/go-ozzo/ozzo-validation/v4"
	"github.com/n-r-w/log-server-v2/pkg/tools"
)

// User Сущность "Пользователь"
type User struct {
	ID                uint64 `json:"id"`
	Login             string `json:"login"`
	Name              string `json:"name"`
	Password          string `json:"password,omitempty"`
	EncryptedPassword string `json:"-"`
}

// IsEmpty ...
func (u *User) IsEmpty() bool {
	return u.ID == 0
}

// Validate Валидация ...
func (u *User) Validate(passwordRegex string, passwordRegexError string) error {
	return validation.ValidateStruct(
		u,
		validation.Field(&u.Login, validation.Required),
		validation.Field(&u.Name, validation.Required),
		validation.Field(&u.Password, validation.When(len(u.EncryptedPassword) == 0, validation.Required)),
		validation.Field(&u.Password, validation.When(len(u.EncryptedPassword) == 0,
			validation.Match(regexp.MustCompile(passwordRegex)).Error(passwordRegexError))),
	)
}

// Prepare Подготовка данных после первой инициализации (инициализация хэша пароля)
func (u *User) Prepare(sanitize bool) error {
	u.Login = strings.TrimSpace(u.Login)
	u.Name = strings.TrimSpace(u.Name)
	u.Password = strings.TrimSpace(u.Password)

	if len(u.Password) > 0 {
		enc, err := tools.EncryptPassword(u.Password)
		if err != nil {

			return err
		}

		u.EncryptedPassword = enc
	}

	if sanitize {
		u.sanitize()
	}

	return nil
}

// Очистка пароля после генерации хэша
func (u *User) sanitize() {
	u.Password = ""
}

// ComparePassword Подходит ли пароль
func (u *User) ComparePassword(password string) bool {
	return tools.ComparePassword(u.EncryptedPassword, password)
}
