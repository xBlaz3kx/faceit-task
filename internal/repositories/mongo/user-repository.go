package mongo

import (
	"context"

	"github.com/kamva/mgm/v3"
	"github.com/pkg/errors"
	"github.com/samber/lo"
	"github.com/xBlaz3kx/faceit-task/internal/repositories"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.uber.org/zap"
)

type userRepository struct {
	logger *zap.Logger
}

func NewUserRepository() *userRepository {
	return &userRepository{
		logger: zap.L().Named("user-repository"),
	}
}

func (u *userRepository) AddUser(ctx context.Context, user *repositories.User) error {
	u.logger.Info("Adding user to the database")
	userCollection := mgm.Coll(&repositories.User{})

	// Check if a user with this email already exists
	cursor, err := userCollection.Find(ctx, bson.M{"email": user.Email})
	if err != nil {
		return err
	}

	if cursor.RemainingBatchLength() > 0 {
		return repositories.ErrUserAlreadyExists
	}

	return mgm.Coll(&repositories.User{}).CreateWithCtx(ctx, user)

}

func (u *userRepository) UpdateUser(ctx context.Context, user repositories.User) (*repositories.User, error) {
	u.logger.Info("Updating user in the database")
	err := mgm.Coll(&repositories.User{}).UpdateWithCtx(ctx, &user)
	switch {
	case err == nil:
		return &user, nil
	case errors.Is(err, mongo.ErrNoDocuments):
		return nil, repositories.ErrUserNotFound
	default:
		return nil, err
	}
}

func (u *userRepository) DeleteUser(ctx context.Context, id string) error {
	u.logger.Info("Deleting user from the database", zap.String("id", id))

	hex, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return err
	}

	one, err := mgm.Coll(&repositories.User{}).DeleteOne(ctx, bson.M{"_id": hex})
	switch {
	case err == nil:
		if one.DeletedCount == 0 {
			return repositories.ErrUserNotFound
		}

		return nil
	case errors.Is(err, mongo.ErrNoDocuments):
		return repositories.ErrUserNotFound
	default:
		return err
	}
}

func (u *userRepository) GetUser(ctx context.Context, id string) (*repositories.User, error) {
	u.logger.Info("Getting a user from the database", zap.String("id", id))

	user := &repositories.User{}
	err := mgm.Coll(&repositories.User{}).FindByIDWithCtx(ctx, id, user)
	switch {
	case err == nil:
		return user, nil
	case errors.Is(err, mongo.ErrNoDocuments):
		return nil, repositories.ErrUserNotFound
	default:
		return nil, err
	}
}

func (u *userRepository) GetUsers(ctx context.Context, firstName, lastName, nickname, country, email *string, limit, offset *int64) ([]repositories.User, error) {
	// Create a query
	query := bson.M{}
	if firstName != nil {
		query["first_name"] = *firstName
	}

	if lastName != nil {
		query["last_name"] = *lastName
	}

	if nickname != nil {
		query["nickname"] = *nickname
	}

	if country != nil {
		query["country"] = *country
	}

	if email != nil {
		query["email"] = *email
	}

	u.logger.Info("Getting users from the database", zap.Any("query", query))

	// Set default limit and offset for pagination
	if limit == nil {
		defaultLimit := int64(30)
		limit = &defaultLimit
	}

	if offset == nil {
		defaultOffset := int64(0)
		offset = &defaultOffset
	}

	opts := &options.FindOptions{
		Limit: lo.ToPtr(*limit),
		Skip:  lo.ToPtr(*offset),
		// Sorting by created_at in descending order
		Sort: bson.D{{"created_at", -1}},
	}

	users := []repositories.User{}
	err := mgm.Coll(&repositories.User{}).SimpleFindWithCtx(ctx, &users, query, opts)
	if err != nil {
		return nil, err
	}

	return users, nil
}
