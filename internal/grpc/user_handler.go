package grpc

import (
	"context"
	"errors"

	"github.com/samber/lo"
	"github.com/xBlaz3kx/faceit-task/internal/domain/users"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.uber.org/zap"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/emptypb"
)

type UserGrpcHandler struct {
	UnimplementedUserServer
	userService users.Service
	logger      *zap.Logger
}

func NewUserGrpcHandler(userService users.Service) *UserGrpcHandler {
	return &UserGrpcHandler{
		userService: userService,
		logger:      zap.L().Named("user-grpc-handler"),
	}
}

func (s *UserGrpcHandler) CreateUser(ctx context.Context, request *CreateUserRequest) (*CreateUserResponse, error) {
	serviceReq := users.NewUser{
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
		return &CreateUserResponse{
			User: toGrpcUser(usr),
		}, nil
	case errors.Is(err, users.ErrValidation):
		return nil, status.Errorf(codes.FailedPrecondition, "failed to validate the user: %v", err.Error())
	case errors.Is(err, users.ErrUserAlreadyExists):
		return nil, status.Errorf(codes.AlreadyExists, "user with email %s already exists", request.GetEmail())
	default:
		return nil, status.Error(codes.Internal, "unknown error occurred while creating the user")
	}
}

func (s *UserGrpcHandler) GetUser(ctx context.Context, request *GetUserRequest) (*GetUserResponse, error) {
	user, err := s.userService.GetUser(ctx, request.GetId())
	switch {
	case err == nil:
		return &GetUserResponse{
			User: toGrpcUser(user),
		}, nil
	case errors.Is(err, users.ErrUserNotFound):
		return nil, status.Errorf(codes.NotFound, "user with id %s not found", request.GetId())
	case errors.Is(err, primitive.ErrInvalidHex):
		return nil, status.Errorf(codes.InvalidArgument, "the provided id is not a valid hex string")
	default:
		return nil, status.Error(codes.Internal, "unknown error occurred while getting the user")
	}
}

func (s *UserGrpcHandler) UpdateUser(ctx context.Context, request *UpdateUserRequest) (*UpdateUserResponse, error) {
	updateUserReq := users.UpdateUser{
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
		return &UpdateUserResponse{
			User: toGrpcUser(user),
		}, nil
	case errors.Is(err, users.ErrUserNotFound):
		return nil, status.Errorf(codes.NotFound, "user with id %s not found", request.GetId())
	case errors.Is(err, primitive.ErrInvalidHex):
		return nil, status.Errorf(codes.InvalidArgument, "the provided id is not a valid hex string")
	default:
		return nil, status.Error(codes.Internal, "unknown error occurred while updating the user")
	}
}

func (s *UserGrpcHandler) DeleteUser(ctx context.Context, request *DeleteUserRequest) (*DeleteUserResponse, error) {
	err := s.userService.DeleteUser(ctx, request.GetId())
	switch {
	case err == nil:
		return &DeleteUserResponse{
			Status: DeleteStatus_OK,
		}, nil
	case errors.Is(err, primitive.ErrInvalidHex):
		return nil, status.Errorf(codes.InvalidArgument, "the provided id is not a valid hex string")
	case errors.Is(err, users.ErrUserNotFound):
		return nil, status.Errorf(codes.NotFound, "user with id %s not found", request.GetId())
	default:
		return nil, status.Error(codes.Internal, "unknown error occurred while deleting the user")
	}
}

func (s *UserGrpcHandler) GetUsers(ctx context.Context, request *ListUsersRequest) (*ListUsersResponse, error) {
	getUsers, err := s.userService.GetUsers(ctx, toQuery(request))
	if err != nil {
		return nil, status.Error(codes.Internal, "unknown error occurred while getting the users")
	}

	userList := lo.Map(getUsers, func(item users.User, index int) *UserModel {
		return toGrpcUser(&item)
	})

	return &ListUsersResponse{
		Users: userList,
	}, nil
}

func toQuery(request *ListUsersRequest) users.Query {
	return users.Query{
		FirstName: request.FirstName,
		LastName:  request.LastName,
		Nickname:  request.Nickname,
		Country:   request.Country,
		Email:     request.Email,
		Limit:     request.Limit,
		Offset:    request.Page,
	}
}

func (s *UserGrpcHandler) Watch(_ *emptypb.Empty, server User_WatchServer) error {

	changeStream, err := s.userService.Watch(server.Context())
	if err != nil {
		s.logger.Error("Failed to watch the users", zap.Error(err))
		return status.Error(codes.Internal, "unknown error occurred while watching the users")
	}

	for {
		select {
		case changeEvent, ok := <-changeStream:
			if !ok {
				s.logger.Error("Change stream closed")
				return nil
			}

			response := toStreamResponse(changeEvent)
			err := server.Send(response)
			if err != nil {
				return err
			}

		case <-server.Context().Done():
			// Check if the client has disconnected or the context has been canceled
			err := server.Context().Err()
			if errors.Is(err, context.Canceled) {
				s.logger.Info("Client disconnected")
			} else {
				s.logger.Error("Client disconnected with error", zap.Error(err))
			}

			return nil
		}
	}
}

func (s *UserGrpcHandler) mustEmbedUnimplementedUserServer() {
}

func toStreamResponse(change users.UserEvent) *WatchStreamResponse {
	return &WatchStreamResponse{
		ChangeType: toChangeType(change.ChangeType),
		User:       toGrpcUser(&change.User),
	}
}

func toChangeType(opType string) ChangeType {
	switch opType {
	case "insert":
		return ChangeType_INSERT
	case "update":
		return ChangeType_UPDATE
	case "delete":
		return ChangeType_DELETE
	default:
		return -1
	}
}

func toGrpcUser(user *users.User) *UserModel {
	return &UserModel{
		Id:       user.ID,
		Name:     user.FirstName,
		LastName: user.LastName,
		Nickname: user.Nickname,
		Email:    user.Email,
		Country:  user.Country,
	}
}
