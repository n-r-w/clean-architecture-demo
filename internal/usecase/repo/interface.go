// Package repo ...
package repo

import (
	"time"

	"github.com/n-r-w/log-server-v2/internal/entity"
)

type (
	// User Интерфейс работы с данными пользователей
	User interface {
		// Insert добавить нового пользователя. ID прописывается в модель
		Insert(user entity.User) error
		Remove(userID uint64) error
		Update(user entity.User) error
		ChangePassword(userID uint64, password string) error

		FindByID(userID uint64) (entity.User, error)
		FindByLogin(login string) (entity.User, error)
		GetUsers() ([]entity.User, error)
	}

	// Log Интерфейс работы с журналом
	Log interface {
		Insert(records []entity.LogRecord) error

		Find(dateFrom time.Time, dateTo time.Time, limit uint) (records []entity.LogRecord, limited bool, err error)
	}
)
