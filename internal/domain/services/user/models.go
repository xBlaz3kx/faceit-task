package user

import (
	"context"
	"errors"
)

var ErrValidation = errors.New("validation error")

type UserService interface {
	AddUser(ctx context.Context, user NewUser) (*User, error)
	UpdateUser(ctx context.Context, user UpdateUser) (*User, error)
	GetUser(ctx context.Context, id string) (*User, error)
	DeleteUser(ctx context.Context, id string) error
	GetUsers(ctx context.Context, query Query) ([]User, error)
	GetChangeStreamChannel(clientId string) <-chan ChangeStreamData
	RemoveStream(clientId string) error
}

type ChangeStreamOperation string

const (
	ChangeStreamOperationInsert ChangeStreamOperation = "insert"
	ChangeStreamOperationUpdate ChangeStreamOperation = "update"
	ChangeStreamOperationDelete ChangeStreamOperation = "delete"
)

// ChangeStreamData is the struct that represents the data sent to the change stream.
type ChangeStreamData struct {
	// OperationType is the type of operation that was performed.
	OperationType ChangeStreamOperation `json:"operationType"`

	// User is the user that was affected by the operation.
	User User `json:"user"`
}

// NewUser is the struct used to create a new user.
type NewUser struct {
	// FirstName of the user.
	FirstName string `json:"first_name"`

	// LastName of the user.
	LastName string `json:"last_name"`

	// Nickname is the nickname of the user.
	Nickname string `json:"nickname"`

	// Email of the user.
	Email string `json:"email" validate:"required,email"`

	// Hashed Password of the user.
	Password string `json:"password" validate:"required,min=8"`

	// Country of the user.
	Country string `json:"country"`
}

// UpdateUser is the struct used to update a user.
type UpdateUser struct {
	Id string `json:"id"`

	// FirstName of the user.
	FirstName string `json:"first_name"`

	// LastName of the user.
	LastName string `json:"last_name"`

	// Nickname is the nickname of the user.
	Nickname string `json:"nickname"`

	// Email of the user.
	Email string `json:"email"`

	// Password of the user (unhashed).
	Password string `json:"password"`

	// Country of the user.
	Country string `json:"country"`
}

// User is the struct that represents a user.
type User struct {
	ID string `json:"id"`

	CreatedAt string `json:"created_at"`

	UpdatedAt string `json:"updated_at"`

	// FirstName of the user.
	FirstName string `json:"first_name"`

	// LastName of the user.
	LastName string `json:"last_name"`

	// Nickname is the nickname of the user.
	Nickname string `json:"nickname"`

	// Email of the user.
	Email string `json:"email"`

	// Country of the user.
	Country string `json:"country"`
}

// Query is a filter for the GetUsers method. Provides limit and offset for pagination.
type Query struct {
	FirstName *string `json:"first_name,omitempty"`
	LastName  *string `json:"last_name,omitempty"`
	Nickname  *string `json:"nickname,omitempty"`
	Country   *string `json:"country,omitempty"`
	Email     *string `json:"email,omitempty"`
	Limit     *int64  `json:"limit,omitempty"`
	Offset    *int64  `json:"offset,omitempty"`
}
