package service

import (
	"context"
	"fmt"
	"time"

	user "api_gateway/pkg/api"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/status"
)

type UserService struct {
	client user.UserServiceClient
	conn   *grpc.ClientConn
}

func NewUserService(addr string) (*UserService, error) {
	conn, err := grpc.Dial(
		addr,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
		grpc.WithDefaultCallOptions(
			grpc.MaxCallRecvMsgSize(1024*1024*10), // 10MB
			grpc.MaxCallSendMsgSize(1024*1024*10), // 10MB
		),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to user-service: %w", err)
	}

	return &UserService{
		client: user.NewUserServiceClient(conn),
		conn:   conn,
	}, nil
}

func (us *UserService) Close() error {
	if us.conn != nil {
		return us.conn.Close()
	}
	return nil
}

func (us *UserService) Create(ctx context.Context, uuid, email, userName, avatar, aboutMe string, friends []string) (*user.UserResponse, error) {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	// Валидация
	if email == "" {
		return nil, fmt.Errorf("email is required")
	}
	if userName == "" {
		return nil, fmt.Errorf("user name is required")
	}

	resp, err := us.client.CreateUser(ctx, &user.CreateUserRequest{
		Uuid:     uuid,
		Email:    email,
		UserName: userName,
		Avatar:   avatar,
		AboutMe:  aboutMe,
		Friends:  friends,
	})
	if err != nil {
		return nil, grpcErrorToError(err)
	}
	return resp, nil
}

func (us *UserService) Update(ctx context.Context, uuid, email, userName, avatar, aboutMe string, friends []string) (*user.UserResponse, error) {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	// Валидация
	if uuid == "" {
		return nil, fmt.Errorf("uuid is required")
	}

	resp, err := us.client.UpdateUser(ctx, &user.UpdateUserRequest{
		Uuid:     uuid,
		Email:    email,
		UserName: userName,
		Avatar:   avatar,
		AboutMe:  aboutMe,
		Friends:  friends,
	})
	if err != nil {
		return nil, grpcErrorToError(err)
	}
	return resp, nil
}

func (us *UserService) Delete(ctx context.Context, uuid string) (*user.DeleteUserResponse, error) {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	// Валидация
	if uuid == "" {
		return nil, fmt.Errorf("uuid is required")
	}

	resp, err := us.client.DeleteUser(ctx, &user.DeleteUserRequest{Uuid: uuid})
	if err != nil {
		return nil, grpcErrorToError(err)
	}
	return resp, nil
}

func (us *UserService) Get(ctx context.Context, uuid string) (*user.UserResponse, error) {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	// Валидация
	if uuid == "" {
		return nil, fmt.Errorf("uuid is required")
	}

	resp, err := us.client.AboutMeUser(ctx, &user.AboutMeRequest{Uuid: uuid})
	if err != nil {
		return nil, grpcErrorToError(err)
	}
	return resp, nil
}

func (us *UserService) SetOnline(ctx context.Context, uuid string, ttlSeconds int32) (*user.StatusResponse, error) {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	// Валидация
	if uuid == "" {
		return nil, fmt.Errorf("uuid is required")
	}

	resp, err := us.client.SetOnline(ctx, &user.SetOnlineRequest{
		Uuid:       uuid,
		TtlSeconds: ttlSeconds,
	})
	if err != nil {
		return nil, grpcErrorToError(err)
	}
	return resp, nil
}

func (us *UserService) IsOnline(ctx context.Context, uuid string) (*user.IsOnlineResponse, error) {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	// Валидация
	if uuid == "" {
		return nil, fmt.Errorf("uuid is required")
	}

	resp, err := us.client.IsOnline(ctx, &user.IsOnlineRequest{Uuid: uuid})
	if err != nil {
		return nil, grpcErrorToError(err)
	}
	return resp, nil
}

func (us *UserService) GetOnlineUsers(ctx context.Context) (*user.GetOnlineUsersResponse, error) {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	resp, err := us.client.GetOnlineUsers(ctx, &user.GetOnlineUsersRequest{})
	if err != nil {
		return nil, grpcErrorToError(err)
	}
	return resp, nil
}

// Вспомогательная функция для преобразования gRPC ошибок
func grpcErrorToError(err error) error {
	if st, ok := status.FromError(err); ok {
		switch st.Code() {
		case codes.NotFound:
			return fmt.Errorf("not found: %s", st.Message())
		case codes.InvalidArgument:
			return fmt.Errorf("invalid argument: %s", st.Message())
		case codes.AlreadyExists:
			return fmt.Errorf("already exists: %s", st.Message())
		case codes.PermissionDenied:
			return fmt.Errorf("permission denied: %s", st.Message())
		default:
			return fmt.Errorf("internal error: %s", st.Message())
		}
	}
	return err
}
