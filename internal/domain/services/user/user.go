package user

import (
	"context"

	"github.com/samber/lo"
	"github.com/xBlaz3kx/faceit-task/internal/repositories"
	"github.com/xBlaz3kx/faceit-task/pkg/notifier"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.uber.org/zap"
)

type userServiceImpl struct {
	repository repositories.UserRepository
	logger     *zap.Logger
	notifier   *notifier.Notifier[ChangeStreamData]
}

func NewUserService(repository repositories.UserRepository) *userServiceImpl {
	return &userServiceImpl{
		repository: repository,
		logger:     zap.L().Named("user-service"),
		notifier:   notifier.NewNotifier[ChangeStreamData](),
	}
}

// AddUser adds a new user to the database.
func (s *userServiceImpl) AddUser(ctx context.Context, user NewUser) (*User, error) {
	s.logger.Info("Adding a new user", zap.Any("user", user))

	repoUser := repositories.NewUser(user.FirstName, user.LastName, user.Nickname, user.Email, user.Password, user.Country)
	err := s.repository.AddUser(ctx, repoUser)
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

func (s *userServiceImpl) GetChangeStreamChannel(clientId string) <-chan ChangeStreamData {
	// todo multiplex the stream - each client should have its own stream and all the clients should receive the same update
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
