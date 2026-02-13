package grpc

import (
	"context"
	"time"

	"main/internal/service"
	user "main/pkg/api"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type userHandler struct {
	user.UnimplementedUserServiceServer
	userService service.UserService
}

func NewUserHandler(userService service.UserService) user.UserServiceServer {
	return &userHandler{
		userService: userService,
	}
}

func (h *userHandler) CreateUser(ctx context.Context, req *user.CreateUserRequest) (*user.UserResponse, error) {
	createReq := service.CreateUserRequest{
		UUID:     req.GetUuid(),
		Email:    req.GetEmail(),
		UserName: req.GetUserName(),
		Avatar:   req.GetAvatar(),
		AboutMe:  req.GetAboutMe(),
	}

	createdUser, err := h.userService.CreateUser(ctx, &createReq)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	return createdUser.ToProto(), nil
}

func (h *userHandler) GetUser(ctx context.Context, req *user.GetUserRequest) (*user.UserResponse, error) {
	u, err := h.userService.GetUser(ctx, req.GetUuid())
	if err != nil {
		if err.Error() == "user not found" {
			return nil, status.Error(codes.NotFound, err.Error())
		}
		return nil, status.Error(codes.Internal, err.Error())
	}
	return u.ToProto(), nil
}

func (h *userHandler) UpdateUser(ctx context.Context, req *user.UpdateUserRequest) (*user.UserResponse, error) {
	updateReq := service.UpdateUserRequest{
		Uuid:     req.GetUuid(),
		Email:    req.GetEmail(),
		UserName: req.GetUserName(),
		Avatar:   req.GetAvatar(),
		AboutMe:  req.GetAboutMe(),
	}

	updatedUser, err := h.userService.UpdateUser(ctx, &updateReq)
	if err != nil {
		if err.Error() == "user not found" {
			return nil, status.Error(codes.NotFound, err.Error())
		}
		return nil, status.Error(codes.Internal, err.Error())
	}

	return updatedUser.ToProto(), nil
}

func (h *userHandler) DeleteUser(ctx context.Context, req *user.DeleteUserRequest) (*user.DeleteUserResponse, error) {
	err := h.userService.DeleteUser(ctx, req.GetUuid())
	if err != nil {
		if err.Error() == "user not found" {
			return nil, status.Error(codes.NotFound, err.Error())
		}
		return nil, status.Error(codes.Internal, err.Error())
	}

	return &user.DeleteUserResponse{
		Success: true,
		Message: "user deleted successfully",
	}, nil
}

func (h *userHandler) ListUsers(ctx context.Context, req *user.ListUsersRequest) (*user.ListUsersResponse, error) {
	// TODO: реализовать если нужно
	return &user.ListUsersResponse{
		Users: []*user.UserResponse{},
		Total: 0,
	}, nil
}
func (h *userHandler) AboutMeUser(ctx context.Context, req *user.AboutMeRequest) (*user.UserResponse, error) {
	resp, err := h.userService.AboutMeUser(ctx, req)
	if err != nil {
		if err.Error() == "user not found" {
			return nil, status.Error(codes.NotFound, err.Error())
		}
		return nil, status.Error(codes.Internal, err.Error())
	}
	return resp, nil
}

func (h *userHandler) SetOnline(ctx context.Context, req *user.SetOnlineRequest) (*user.StatusResponse, error) {
	err := h.userService.SetOnlineUser(ctx, req.GetUuid(), time.Duration(req.GetTtlSeconds())*time.Second)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}
	return &user.StatusResponse{Success: true, Message: "user set online"}, nil
}

func (h *userHandler) SetOffline(ctx context.Context, req *user.SetOfflineRequest) (*user.StatusResponse, error) {
	err := h.userService.SetOffline(ctx, req.GetUuid())
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}
	return &user.StatusResponse{Success: true, Message: "user set offline"}, nil
}

func (h *userHandler) IsOnline(ctx context.Context, req *user.IsOnlineRequest) (*user.IsOnlineResponse, error) {
	online, err := h.userService.IsOnline(ctx, req.GetUuid())
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}
	return &user.IsOnlineResponse{Uuid: req.GetUuid(), Online: online}, nil
}

func (h *userHandler) GetOnlineUsers(ctx context.Context, req *user.GetOnlineUsersRequest) (*user.GetOnlineUsersResponse, error) {
	users, err := h.userService.GetOnlineUsers(ctx)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}
	return &user.GetOnlineUsersResponse{Uuids: users}, nil
}
