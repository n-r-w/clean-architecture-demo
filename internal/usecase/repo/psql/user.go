// Package psql Содержит интерфейс для работы с хранилищем пользователей в postgres
package psql

import (
	"context"
	"database/sql"
	"errors"
	"strings"

	"github.com/n-r-w/log-server-v2/internal/entity"
	"github.com/n-r-w/log-server-v2/internal/usecase/repo"
	"github.com/n-r-w/log-server-v2/pkg/logger"
	"github.com/n-r-w/log-server-v2/pkg/postgres"
	"github.com/n-r-w/log-server-v2/pkg/tools"
	"github.com/omeid/pgerror"
)

type UserRepo struct {
	*postgres.Postgres
	logger             logger.Interface
	superAdminID       uint64
	superAdminLogin    string
	superAdminPassword string

	passwordRegex      string
	passwordRegexError string
}

func NewUser(pg *postgres.Postgres, logger logger.Interface, superAdminID uint64, superAdminLogin string, superAdminPassword string,
	passwordRegex string, passwordRegexError string) *UserRepo {
	return &UserRepo{
		Postgres:           pg,
		logger:             logger,
		superAdminID:       superAdminID,
		superAdminLogin:    superAdminLogin,
		superAdminPassword: superAdminPassword,
		passwordRegex:      passwordRegex,
		passwordRegexError: passwordRegexError,
	}
}

// Insert Добавить нового пользвателя
func (r *UserRepo) Insert(user entity.User) error {
	if user.ID == r.superAdminID || strings.EqualFold(user.Login, r.superAdminLogin) {
		return repo.ErrCantChangeAdminUser
	}

	if err := user.Prepare(true); err != nil {
		return err
	}

	if err := user.Validate(r.passwordRegex, r.passwordRegexError); err != nil {
		return err
	}

	err := r.Pool.QueryRow(context.Background(),
		"INSERT INTO users (login, name, encrypted_password) VALUES ($1, $2, $3) RETURNING id",
		user.Login,
		user.Name,
		user.EncryptedPassword,
	).Scan(&user.ID)
	if err != nil {
		if e := pgerror.UniqueViolation(err); e != nil {
			return repo.ErrLoginExist
		}

		return err
	}

	return err
}

// ChangePassword Изменить пароль пользователя
func (r *UserRepo) ChangePassword(userID uint64, password string) error {
	if userID == r.superAdminID {
		return repo.ErrCantChangeAdminPassword
	}

	password = strings.TrimSpace(password)
	enc, err := tools.EncryptPassword(password)

	if err != nil {
		return err
	}

	var user entity.User
	user, err = r.FindByID(userID)

	if err != nil {
		return err
	}

	if user.IsEmpty() {
		return repo.ErrUserNotFound
	}

	user.Password = password
	if err = user.Validate(r.passwordRegex, r.passwordRegexError); err != nil {
		return err
	}

	if err = user.Prepare(true); err != nil {
		return err
	}

	_, err = r.Pool.Exec(context.Background(), "UPDATE users SET encrypted_password=$1 WHERE id=$2", enc, userID)
	if err != nil {
		if e := pgerror.UniqueViolation(err); e != nil {
			return repo.ErrLoginExist
		}

		return err
	}

	return nil
}

// FindByID Поиск пользователя по ID
func (r *UserRepo) FindByID(userID uint64) (entity.User, error) {
	// не админ ли это?
	if userID == r.superAdminID {
		return r.AdminUser(), nil
	}

	u := entity.User{
		ID:                0,
		Login:             "",
		Name:              "",
		Password:          "",
		EncryptedPassword: "",
	}
	if err := r.Pool.QueryRow(context.Background(),
		"SELECT id, login, name, encrypted_password FROM users WHERE id = $1",
		userID,
	).Scan(
		&u.ID,
		&u.Login,
		&u.Name,
		&u.EncryptedPassword,
	); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return entity.User{}, nil
		}

		return entity.User{}, err
	}

	return u, nil
}

// FindByLogin Поиск пользователя по логину
func (r *UserRepo) FindByLogin(login string) (entity.User, error) {
	u := entity.User{
		ID:                0,
		Login:             "",
		Name:              "",
		Password:          "",
		EncryptedPassword: "",
	}

	// не админ ли это?
	if strings.EqualFold(login, r.superAdminLogin) {
		u = r.AdminUser()
	} else {
		if err := r.Pool.QueryRow(context.Background(),
			"SELECT id, login, name, encrypted_password FROM users WHERE login = $1",
			login,
		).Scan(
			&u.ID,
			&u.Login,
			&u.Name,
			&u.EncryptedPassword,
		); err != nil {
			if errors.Is(err, sql.ErrNoRows) {
				return entity.User{}, nil
			}

			return entity.User{}, err
		}
	}

	return u, nil
}

// GetUsers Получить список пользователей
func (r *UserRepo) GetUsers() ([]entity.User, error) {
	rows, err := r.Pool.Query(context.Background(),
		`SELECT id, login, name, encrypted_password FROM users`)
	if err != nil {

		return nil, err
	}
	defer rows.Close() // освобождаем контекст sql запроса при выходе

	var users []entity.User

	for rows.Next() {
		var usr entity.User
		err = rows.Scan(&usr.ID, &usr.Login, &usr.Name, &usr.EncryptedPassword)

		if err != nil {
			return nil, err
		}

		users = append(users, usr)
	}

	rows.Close()

	return users, nil
}

func (r *UserRepo) Remove(_ uint64) error {

	return errors.New("not implemeted")
}

func (r *UserRepo) Update(_ entity.User) error {
	
	return errors.New("not implemeted")
}

// AdminUser - Фейковый пользователь - админ
func (r *UserRepo) AdminUser() entity.User {
	user := entity.User{
		ID:                r.superAdminID,
		Name:              "admin",
		Login:             r.superAdminLogin,
		Password:          r.superAdminPassword,
		EncryptedPassword: "",
	}

	if err := user.Prepare(true); err != nil {
		r.logger.Error("user prepare error %v", err)
	}

	return user
}
