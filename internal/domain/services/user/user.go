package user

import (
	"context"
	errs "errors"

	"github.com/go-playground/validator/v10"
	"github.com/samber/lo"
	"github.com/xBlaz3kx/faceit-task/internal/repositories"
	"github.com/xBlaz3kx/faceit-task/pkg/notifier"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.uber.org/zap"
)

var validate = validator.New()

type userServiceImpl struct {
	repository repositories.UserRepository
	logger     *zap.Logger
	notifier   *notifier.Notifier[ChangeStreamData]
}

func NewUserService(repository repositories.UserRepository, notifier *notifier.Notifier[ChangeStreamData]) *userServiceImpl {
	return &userServiceImpl{
		repository: repository,
		logger:     zap.L().Named("user-service"),
		notifier:   notifier,
	}
}

// AddUser adds a new user to the database.
func (s *userServiceImpl) AddUser(ctx context.Context, user NewUser) (*User, error) {
	s.logger.Info("Adding a new user", zap.Any("user", user))

	// Validate the user
	err := validate.Struct(user)
	if err != nil {
		return nil, errs.Join(ErrValidation, err)
	}

	repoUser := repositories.NewUser(user.FirstName, user.LastName, user.Nickname, user.Email, user.Password, user.Country)
	err = s.repository.AddUser(ctx, repoUser)
	if err != nil {
		return nil, err
	}

	response := toUser(repoUser)

	// Notify the subscribers about the update
	s.notifier.Broadcast(ChangeStreamData{
		OperationType: ChangeStreamOperationInsert,
		User:          response,
	})

	return lo.ToPtr(response), nil
}

// UpdateUser updates a user in the database.
func (s *userServiceImpl) UpdateUser(ctx context.Context, user UpdateUser) (*User, error) {
	s.logger.Info("Updating a user", zap.Any("user", user))

	hex, err := primitive.ObjectIDFromHex(user.Id)
	if err != nil {
		return nil, err
	}

	// todo check if password is updated - we don't want to overwrite the password with the empty string

	repoUser := repositories.NewUser(user.FirstName, user.LastName, user.Nickname, user.Email, user.Password, user.Country)
	repoUser.ID = hex

	repoUser, err = s.repository.UpdateUser(ctx, *repoUser)
	if err != nil {
		return nil, err
	}

	response := toUser(repoUser)

	// Notify the subscribers about the update
	s.notifier.Broadcast(ChangeStreamData{
		OperationType: ChangeStreamOperationUpdate,
		User:          response,
	})

	return lo.ToPtr(response), nil
}

// DeleteUser deletes a user from the database.
func (s *userServiceImpl) DeleteUser(ctx context.Context, id string) error {
	s.logger.Info("Deleting a user", zap.String("id", id))

	// Notify the subscribers about the deletion
	s.notifier.Broadcast(ChangeStreamData{
		OperationType: ChangeStreamOperationDelete,
		User:          User{ID: id},
	})

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

// GetChangeStreamChannel returns a channel that will be used to notify the clients about the changes in the database.
func (s *userServiceImpl) GetChangeStreamChannel(clientId string) <-chan ChangeStreamData {
	// Each client should have its own stream and all the clients should receive the same update.
	return s.notifier.AddSubscriber(clientId)
}

// RemoveStream the client from the stream multiplexer
func (s *userServiceImpl) RemoveStream(clientId string) error {
	s.notifier.RemoveSubscriber(clientId)
	return nil
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
