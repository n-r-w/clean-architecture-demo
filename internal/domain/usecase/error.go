package usecase

import "errors"

var (
	errNotAdmin     = errors.New("not admin user")
	errUserNotFound = errors.New("user not found")
)
