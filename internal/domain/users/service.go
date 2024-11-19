package users

import (
	"context"
	"errors"

	"github.com/go-playground/validator/v10"
	"github.com/samber/lo"
	"go.uber.org/zap"
)

var ErrValidation = errors.New("validation error")

type Service interface {
	AddUser(ctx context.Context, user NewUser) (*User, error)
	UpdateUser(ctx context.Context, user UpdateUser) (*User, error)
	GetUser(ctx context.Context, id string) (*User, error)
	DeleteUser(ctx context.Context, id string) error
	GetUsers(ctx context.Context, query Query) ([]User, error)
	Watch(ctx context.Context) (<-chan UserEvent, error)
}

var validate = validator.New()

type userServiceImpl struct {
	repository Repository
	logger     *zap.Logger
}

func NewUserService(repository Repository) *userServiceImpl {
	return &userServiceImpl{
		repository: repository,
		logger:     zap.L().Named("user-service"),
	}
}

// AddUser adds a new user to the database.
func (s *userServiceImpl) AddUser(ctx context.Context, user NewUser) (*User, error) {
	s.logger.Info("Adding a new user", zap.Any("user", user))

	// Validate the user
	err := validate.Struct(user)
	if err != nil {
		return nil, errors.Join(ErrValidation, err)
	}

	newUser := toUser(&user)
	err = s.repository.AddUser(ctx, newUser)
	if err != nil {
		return nil, err
	}

	return newUser, nil
}

// UpdateUser updates a user in the database.
func (s *userServiceImpl) UpdateUser(ctx context.Context, user UpdateUser) (*User, error) {
	s.logger.Info("Updating a user", zap.Any("user", user))

	// todo check if password is updated - we don't want to overwrite the password with the empty string

	repoUser := User{
		FirstName: user.FirstName,
		LastName:  user.LastName,
		Nickname:  user.Nickname,
		Email:     user.Email,
		Country:   user.Country,
	}
	res, err := s.repository.UpdateUser(ctx, repoUser)
	if err != nil {
		return nil, err
	}

	return res, nil
}

// DeleteUser deletes a user from the database.
func (s *userServiceImpl) DeleteUser(ctx context.Context, id string) error {
	s.logger.Info("Deleting a user", zap.String("id", id))

	return s.repository.DeleteUser(ctx, id)
}

// GetUsers returns a list of users from the database.
func (s *userServiceImpl) GetUsers(ctx context.Context, query Query) ([]User, error) {
	s.logger.Info("Getting users", zap.Any("query", query))

	repoUsers, err := s.repository.GetUsers(ctx, query.FirstName, query.LastName, query.Nickname, query.Country, query.Email, query.Limit, query.Offset)
	if err != nil {
		return nil, err
	}

	resp := lo.Map(repoUsers, func(item User, index int) User {
		return item
	})

	return resp, nil
}

// GetUser returns a user from the database.
func (s *userServiceImpl) GetUser(ctx context.Context, id string) (*User, error) {
	s.logger.Info("Getting users", zap.String("id", id))

	repoUser, err := s.repository.GetUser(ctx, id)
	if err != nil {
		return nil, err
	}

	return repoUser, nil
}

func (s *userServiceImpl) Watch(ctx context.Context) (<-chan UserEvent, error) {
	return s.repository.Watch(ctx)
}

func toUser(user *NewUser) *User {
	return &User{

		FirstName: user.FirstName,
		LastName:  user.LastName,
		Nickname:  user.Nickname,
		Email:     user.Email,
		Country:   user.Country,
	}
}
