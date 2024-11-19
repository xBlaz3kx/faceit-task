package users

import (
	"context"
	"errors"
)

var (
	ErrUserNotFound      = errors.New("user not found")
	ErrUserAlreadyExists = errors.New("user already exists")
)

type Repository interface {
	AddUser(ctx context.Context, user *User) error
	UpdateUser(ctx context.Context, user User) (*User, error)
	DeleteUser(ctx context.Context, id string) error
	GetUser(ctx context.Context, id string) (*User, error)
	GetUsers(ctx context.Context, firstName, lastName, nickname, country, email *string, limit, offset *int64) ([]User, error)
	Watch(ctx context.Context) (<-chan UserEvent, error)
}
