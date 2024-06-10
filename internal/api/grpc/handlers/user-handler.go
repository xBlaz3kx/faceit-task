package handlers

import (
	"context"
	"errors"

	"github.com/google/uuid"
	"github.com/samber/lo"
	"github.com/xBlaz3kx/faceit-task/internal/domain/services/user"
	"github.com/xBlaz3kx/faceit-task/internal/repositories"
	v1 "github.com/xBlaz3kx/faceit-task/pkg/proto/v1"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.uber.org/zap"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/emptypb"
)

type UserGrpcHandler struct {
	v1.UnimplementedUserServer
	userService user.UserService
	logger      *zap.Logger
}

func NewUserGrpcHandler(userService user.UserService) *UserGrpcHandler {
	return &UserGrpcHandler{
		userService: userService,
		logger:      zap.L().Named("user-grpc-handler"),
	}
}

func (s *UserGrpcHandler) CreateUser(ctx context.Context, request *v1.CreateUserRequest) (*v1.CreateUserResponse, error) {
	serviceReq := user.NewUser{
		FirstName: request.GetFirstName(),
		LastName:  request.GetLastName(),
		Nickname:  request.GetNickname(),
		Email:     request.GetEmail(),
		Password:  request.GetPassword(),
		Country:   request.GetCountry(),
	}

	usr, err := s.userService.AddUser(ctx, serviceReq)
	switch {
	case err == nil:
		return &v1.CreateUserResponse{
			User: toGrpcUser(usr),
		}, nil
	case errors.Is(err, user.ErrValidation):
		return nil, status.Errorf(codes.FailedPrecondition, "failed to validate the user: %v", err.Error())
	default:
		return nil, status.Error(codes.Internal, "unknown error occurred while creating the user")
	}
}

func (s *UserGrpcHandler) GetUser(ctx context.Context, request *v1.GetUserRequest) (*v1.GetUserResponse, error) {
	user, err := s.userService.GetUser(ctx, request.GetId())
	switch {
	case err == nil:
		return &v1.GetUserResponse{
			User: toGrpcUser(user),
		}, nil
	case errors.Is(err, repositories.ErrUserNotFound):
		return nil, status.Errorf(codes.NotFound, "user with id %s not found", request.GetId())
	case errors.Is(err, primitive.ErrInvalidHex):
		return nil, status.Errorf(codes.InvalidArgument, "the provided id is not a valid hex string")
	default:
		return nil, status.Error(codes.Internal, "unknown error occurred while getting the user")
	}
}

func (s *UserGrpcHandler) UpdateUser(ctx context.Context, request *v1.UpdateUserRequest) (*v1.UpdateUserResponse, error) {
	updateUserReq := user.UpdateUser{
		Id:        request.GetId(),
		FirstName: request.GetFirstName(),
		LastName:  request.GetLastName(),
		Nickname:  request.GetNickname(),
		Email:     request.GetEmail(),
		Password:  request.GetPassword(),
		Country:   request.GetCountry(),
	}

	user, err := s.userService.UpdateUser(ctx, updateUserReq)
	switch {
	case err == nil:
		return &v1.UpdateUserResponse{
			User: toGrpcUser(user),
		}, nil
	case errors.Is(err, repositories.ErrUserNotFound):
		return nil, status.Errorf(codes.NotFound, "user with id %s not found", request.GetId())
	case errors.Is(err, primitive.ErrInvalidHex):
		return nil, status.Errorf(codes.InvalidArgument, "the provided id is not a valid hex string")
	default:
		return nil, status.Error(codes.Internal, "unknown error occurred while updating the user")
	}
}

func (s *UserGrpcHandler) DeleteUser(ctx context.Context, request *v1.DeleteUserRequest) (*v1.DeleteUserResponse, error) {
	err := s.userService.DeleteUser(ctx, request.GetId())
	switch {
	case err == nil:
		return &v1.DeleteUserResponse{
			Status: v1.DeleteStatus_OK,
		}, nil
	case errors.Is(err, primitive.ErrInvalidHex):
		return nil, status.Errorf(codes.InvalidArgument, "the provided id is not a valid hex string")
	case errors.Is(err, repositories.ErrUserNotFound):
		return nil, status.Errorf(codes.NotFound, "user with id %s not found", request.GetId())
	default:
		return nil, status.Error(codes.Internal, "unknown error occurred while deleting the user")
	}
}

func (s *UserGrpcHandler) GetUsers(ctx context.Context, request *v1.ListUsersRequest) (*v1.ListUsersResponse, error) {
	users, err := s.userService.GetUsers(ctx, user.Query{
		FirstName: request.FirstName,
		LastName:  request.LastName,
		Nickname:  request.Nickname,
		Country:   request.Country,
		Email:     request.Email,
		Limit:     request.Limit,
		Offset:    request.Page,
	})
	if err != nil {
		return nil, status.Error(codes.Internal, "unknown error occurred while getting the users")
	}

	userList := lo.Map(users, func(item user.User, index int) *v1.UserModel {
		return toGrpcUser(&item)
	})

	return &v1.ListUsersResponse{
		Users: userList,
	}, nil
}

func (s *UserGrpcHandler) Watch(_ *emptypb.Empty, server v1.User_WatchServer) error {
	// For simplicity, we will generate a random client id for creating and storing a change stream for a client
	clientId := uuid.New().String()

	// Create a new change stream channel for the client
	streamChannel := s.userService.GetChangeStreamChannel(clientId)
	// After the client disconnects, remove the change stream channel
	defer s.userService.RemoveStream(clientId)

	for {
		select {
		// Check if the client has disconnected or the context has been canceled
		case <-server.Context().Done():
			err := server.Context().Err()

			if errors.Is(err, context.Canceled) {
				s.logger.Info("Client disconnected")
			} else {
				s.logger.Error("Client disconnected with error", zap.Error(err))
			}

			return nil
		case change, closed := <-streamChannel:
			if !closed {
				return nil
			}

			response := &v1.WatchStreamResponse{
				ChangeType: toChangeType(change.OperationType),
				User:       toGrpcUser(&change.User),
			}

			// Send the update to the stream
			err := server.Send(response)
			if err != nil {
				return err
			}
		}
	}
}

func (s *UserGrpcHandler) mustEmbedUnimplementedUserServer() {
}

func toChangeType(operationType user.ChangeStreamOperation) v1.ChangeType {
	switch operationType {
	case user.ChangeStreamOperationInsert:
		return v1.ChangeType_INSERT
	case user.ChangeStreamOperationUpdate:
		return v1.ChangeType_UPDATE
	case user.ChangeStreamOperationDelete:
		return v1.ChangeType_DELETE
	default:
		// This should never happen
		return -1
	}
}

func toGrpcUser(user *user.User) *v1.UserModel {
	return &v1.UserModel{
		Id:       user.ID,
		Name:     user.FirstName,
		LastName: user.LastName,
		Nickname: user.Nickname,
		Email:    user.Email,
		Country:  user.Country,
	}
}
