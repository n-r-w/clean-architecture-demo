// Package usecase Модели данных, относящиеся к пользователю. Сейчас все в одном
// файле. При большом количестве моделей и операций имеет смысл разбить на
// несколько файлов или каталогов Сюда не входит операции, связанные с
// аутентификацие пользователя (логин, выход и т.п.)
package usecase

import (
	"errors"
	"strings"

	"github.com/n-r-w/log-server-v2/internal/entity"
	"github.com/n-r-w/log-server-v2/internal/usecase/repo"
)

type UserUseCase struct {
	repo         repo.User
	superAdminID uint64
}

func NewUserCase(r repo.User, superAdminID uint64) *UserUseCase {
	return &UserUseCase{
		repo:         r,
		superAdminID: superAdminID,
	}
}

// CheckPassword Проверить пароль
func (u *UserUseCase) CheckPassword(login string, password string) (ID uint64, err error) {
	// ищем в БД по логину
	user, err := u.repo.FindByLogin(login)
	if err != nil {
		return 0, err
	}
	// проверяем наличие пользователя в БД и пароль
	if u == nil || !user.ComparePassword(password) {
		return 0, errors.New("incorrect email or password")
	}

	return user.ID, nil
}

// ChangePassword Проверить пароль
func (u *UserUseCase) ChangePassword(currentUser entity.User, login string, password string) (ID uint64, err error) {
	login = strings.TrimSpace(login)
	password = strings.TrimSpace(password)
	changeSelf := currentUser.Login == login

	var id uint64

	if !changeSelf {
		if currentUser.ID != u.superAdminID {
			// если не админ, то менять можно только себе
			return 0, errNotAdmin
		}

		user, err := u.FindByLogin(login)
		if err != nil {
			return 0, err
		}

		if user.IsEmpty() {
			return 0, errUserNotFound
		}

		id = user.ID
	} else {
		id = currentUser.ID
	}

	return id, u.repo.ChangePassword(id, password)
}

func (u *UserUseCase) Insert(user entity.User) error {
	return u.repo.Insert(user) //nolint:wrapcheck
}

func (u *UserUseCase) Remove(id uint64) error {
	return u.repo.Remove(id) //nolint:wrapcheck
}

func (u *UserUseCase) Update(user entity.User) error {
	return u.repo.Update(user) //nolint:wrapcheck
}

func (u *UserUseCase) FindByID(id uint64) (entity.User, error) {
	return u.repo.FindByID(id) //nolint:wrapcheck
}

func (u *UserUseCase) FindByLogin(login string) (entity.User, error) {
	return u.repo.FindByLogin(login) //nolint:wrapcheck
}

func (u *UserUseCase) GetUsers() ([]entity.User, error) {
	return u.repo.GetUsers() //nolint:wrapcheck
}
