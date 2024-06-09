package repositories

import (
	"context"
	"errors"

	"github.com/kamva/mgm/v3"
	"golang.org/x/crypto/bcrypt"
)

var (
	ErrUserNotFound = errors.New("user not found")
)

type UserRepository interface {
	AddUser(ctx context.Context, user *User) error
	UpdateUser(ctx context.Context, user User) (*User, error)
	DeleteUser(ctx context.Context, id string) error
	GetUser(ctx context.Context, id string) (*User, error)
	GetUsers(ctx context.Context, firstName, lastName, nickname, country, email *string, limit, offset *int64) ([]User, error)
}

const schemaVersion = 1

type User struct {
	// DefaultModels contains the id, created and updated at fields for the model.
	mgm.DefaultModel `bson:",inline"`

	SchemaVersion int `json:"schema_version" bson:"schema_version"`

	// FirstName of the user.
	FirstName string `json:"first_name"`

	// LastName of the user.
	LastName string `json:"last_name"`

	// Nickname is the nickname of the user.
	Nickname string `json:"nickname"`

	// Email of the user.
	Email string `json:"email"`

	// Hashed Password of the user.
	Password string `json:"password"`

	// Country of the user.
	Country string `json:"country"`
}

// NewUser creates a new user with the given parameters.
func NewUser(firstName, lastName, nickname, email, password, country string) *User {
	return &User{
		SchemaVersion: schemaVersion,
		FirstName:     firstName,
		LastName:      lastName,
		Nickname:      nickname,
		Email:         email,
		Password:      password,
		Country:       country,
	}
}

// Creating is a hook that is called before the user is created. It will hash the password to securely store it.
func (u *User) Creating(ctx context.Context) error {
	if u.SchemaVersion != schemaVersion {
		u.SchemaVersion = schemaVersion
	}

	// Hash the password
	pass, err := hashPassword(u.Password)
	if err != nil {
		return err
	}

	u.Password = pass
	return nil
}

func hashPassword(password string) (string, error) {
	bytes, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	return string(bytes), err
}
