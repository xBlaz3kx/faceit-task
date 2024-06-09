package user

import (
	"context"

	"github.com/samber/lo"
	"github.com/xBlaz3kx/faceit-task/internal/repositories"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.uber.org/zap"
)

type UserService interface {
	AddUser(ctx context.Context, user NewUser) (*User, error)
	UpdateUser(ctx context.Context, user UpdateUser) (*User, error)
	GetUser(ctx context.Context, id string) (*User, error)
	DeleteUser(ctx context.Context, id string) error
	GetUsers(ctx context.Context, query Query) ([]User, error)
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
	Email string `json:"email"`

	// Hashed Password of the user.
	Password string `json:"password"`

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

type userServiceImpl struct {
	repository repositories.UserRepository
	logger     *zap.Logger
}

func NewUserService(repository repositories.UserRepository) *userServiceImpl {
	return &userServiceImpl{
		repository: repository,
		logger:     zap.L().Named("user-service"),
	}
}

// AddUser adds a new user to the database.
func (s *userServiceImpl) AddUser(ctx context.Context, user NewUser) (*User, error) {
	s.logger.Info("Adding a new user", zap.Any("user", user))

	repoUser := repositories.NewUser(user.FirstName, user.LastName, user.Nickname, user.Email, user.Password, user.Country)
	err := s.repository.AddUser(ctx, *repoUser)
	if err != nil {
		return nil, err
	}

	return lo.ToPtr(toUser(repoUser)), nil
}

// UpdateUser updates a user in the database.
func (s *userServiceImpl) UpdateUser(ctx context.Context, user UpdateUser) (*User, error) {
	s.logger.Info("Updating a user", zap.Any("user", user))

	hex, err := primitive.ObjectIDFromHex(user.Id)
	if err != nil {
		return nil, err
	}

	repoUser := repositories.NewUser(user.FirstName, user.LastName, user.Nickname, user.Email, user.Password, user.Country)
	repoUser.ID = hex

	repoUser, err = s.repository.UpdateUser(ctx, *repoUser)
	if err != nil {
		return nil, err
	}

	return lo.ToPtr(toUser(repoUser)), nil
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

	resp := lo.Map(repoUsers, func(item repositories.User, index int) User {
		return toUser(&item)
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

	return lo.ToPtr(toUser(repoUser)), nil
}

func toUser(user *repositories.User) User {
	return User{
		ID:        user.ID.Hex(),
		CreatedAt: user.CreatedAt.String(),
		UpdatedAt: user.UpdatedAt.String(),
		FirstName: user.FirstName,
		LastName:  user.LastName,
		Nickname:  user.Nickname,
		Email:     user.Email,
		Country:   user.Country,
	}
}
