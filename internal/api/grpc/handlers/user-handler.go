package handlers

import (
	"context"

	"github.com/samber/lo"
	"github.com/xBlaz3kx/faceit-task/internal/domain/services/user"
	v1 "github.com/xBlaz3kx/faceit-task/pkg/proto/v1"
	"go.uber.org/zap"
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

	user, err := s.userService.AddUser(ctx, serviceReq)
	// todo handle error cases
	if err != nil {
		return nil, err
	}

	return &v1.CreateUserResponse{
		User: toGrpcUser(user),
	}, nil
}

func (s *UserGrpcHandler) GetUser(ctx context.Context, request *v1.GetUserRequest) (*v1.GetUserResponse, error) {
	user, err := s.userService.GetUser(ctx, request.GetId())
	if err != nil {
		return nil, err
	}

	return &v1.GetUserResponse{
		User: toGrpcUser(user),
	}, nil
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
	if err != nil {
		return nil, err
	}

	return &v1.UpdateUserResponse{
		User: toGrpcUser(user),
	}, nil
}

func (s *UserGrpcHandler) DeleteUser(ctx context.Context, request *v1.DeleteUserRequest) (*v1.DeleteUserResponse, error) {
	err := s.userService.DeleteUser(ctx, request.GetId())
	if err != nil {
		return nil, err
	}

	return &v1.DeleteUserResponse{
		Status: "Success",
	}, nil
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
		return nil, err
	}

	userList := lo.Map(users, func(item user.User, index int) *v1.UserModel {
		return toGrpcUser(&item)
	})

	return &v1.ListUsersResponse{
		Users: userList,
	}, nil
}

func (s *UserGrpcHandler) mustEmbedUnimplementedUserServer() {
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
