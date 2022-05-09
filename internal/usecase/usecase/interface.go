package usecase

import (
	"time"

	"github.com/n-r-w/log-server-v2/internal/entity"
)

type (
	User interface {
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

	Log interface {
		Insert(logs []entity.LogRecord) error

		Find(dateFrom time.Time, dateTo time.Time, limit uint) (records []entity.LogRecord, limited bool, err error)
	}
)
