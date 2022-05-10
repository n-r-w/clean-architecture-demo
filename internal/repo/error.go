// Package repo ...
package repo

import "errors"

var (
	ErrLoginExist              = errors.New("login exist")
	ErrUserNotFound            = errors.New("user not found")
	ErrCantChangeAdminPassword = errors.New("can't change admin password")
	ErrCantChangeAdminUser     = errors.New("can't change admin user")
)
