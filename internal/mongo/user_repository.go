package mongo

import (
	"context"
	"time"

	"github.com/kamva/mgm/v3"
	"github.com/pkg/errors"
	"github.com/samber/lo"
	"github.com/xBlaz3kx/faceit-task/internal/domain/users"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.uber.org/zap"
)

type userRepository struct {
	logger *zap.Logger
}

func NewUserRepository() users.Repository {
	return &userRepository{
		logger: zap.L().Named("user-repository"),
	}
}

func (u *userRepository) AddUser(ctx context.Context, user *users.User) error {
	u.logger.Info("Adding user to the database")
	userCollection := mgm.Coll(&User{})

	// Check if a user with this email already exists
	cursor, err := userCollection.Find(ctx, bson.M{"email": user.Email})
	if err != nil {
		return err
	}

	if cursor.RemainingBatchLength() > 0 {
		return users.ErrUserAlreadyExists
	}

	userEntity := toEntity(user)
	return mgm.Coll(&User{}).CreateWithCtx(ctx, &userEntity)

}

func (u *userRepository) UpdateUser(ctx context.Context, user users.User) (*users.User, error) {
	u.logger.Info("Updating user in the database")

	userEntity := toEntity(&user)
	err := mgm.Coll(&User{}).UpdateWithCtx(ctx, &userEntity)
	switch {
	case err == nil:
		res := toUser(&userEntity)
		return res, nil
	case errors.Is(err, mongo.ErrNoDocuments):
		return nil, users.ErrUserNotFound
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

	one, err := mgm.Coll(&User{}).DeleteOne(ctx, bson.M{"_id": hex})
	switch {
	case err == nil:
		if one.DeletedCount == 0 {
			return users.ErrUserNotFound
		}

		return nil
	case errors.Is(err, mongo.ErrNoDocuments):
		return users.ErrUserNotFound
	default:
		return err
	}
}

func (u *userRepository) GetUser(ctx context.Context, id string) (*users.User, error) {
	u.logger.Info("Getting a user from the database", zap.String("id", id))

	user := &User{}
	err := mgm.Coll(&User{}).FindByIDWithCtx(ctx, id, user)
	switch {
	case err == nil:
		return toUser(user), nil
	case errors.Is(err, mongo.ErrNoDocuments):
		return nil, users.ErrUserNotFound
	default:
		return nil, err
	}
}

func (u *userRepository) GetUsers(ctx context.Context, firstName, lastName, nickname, country, email *string, limit, offset *int64) ([]users.User, error) {
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

	dbUsers := []User{}
	err := mgm.Coll(&User{}).SimpleFindWithCtx(ctx, &dbUsers, query, opts)
	if err != nil {
		return nil, err
	}

	response := make([]users.User, len(dbUsers))
	for i, user := range dbUsers {
		response[i] = *toUser(&user)
	}

	return response, nil
}

func (u *userRepository) Watch(ctx context.Context) (<-chan users.UserEvent, error) {
	_, _, database, err := mgm.DefaultConfigs()
	if err != nil {
		return nil, err
	}

	// Assuming no other service will be writing to the database, we can use the change stream to get the changes
	matchStage := bson.D{{
		"$match",
		bson.D{{
			"operationType",
			bson.D{{"$in", bson.A{"insert", "update", "delete"}}},
		},
		},
	}}

	opts := options.ChangeStream().SetMaxAwaitTime(2 * time.Second)
	changeStream, err := database.Watch(ctx, mongo.Pipeline{matchStage}, opts)
	if err != nil {
		return nil, err
	}

	//nolint:errcheck
	defer changeStream.Close(context.Background())

	userChan := make(chan users.UserEvent)

	// Stream to goroutine
	go func() {
		for {
			select {
			case <-ctx.Done():
				return
			default:

			}
		}
	}()

	return userChan, nil
}

func toEntity(user *users.User) User {
	hex, _ := primitive.ObjectIDFromHex(user.ID)

	entity := User{

		SchemaVersion: schemaVersion,
		FirstName:     user.FirstName,
		LastName:      user.LastName,
		Nickname:      user.Nickname,
		Email:         user.Email,
		Country:       user.Country,
	}
	entity.SetID(hex)
	return entity
}

func toUser(user *User) *users.User {
	return &users.User{
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
