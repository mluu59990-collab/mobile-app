package model

import "errors"

var (
	ErrEmailTaken      = errors.New("email already registered")
	ErrNotFound        = errors.New("not found")
	ErrRefreshInvalid  = errors.New("invalid refresh token")
	ErrExpenseNotFound = errors.New("expense not found")
)
