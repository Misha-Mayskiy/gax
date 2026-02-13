package userserviceclient

import (
	"context"
	userpb "main/pkg/api_user_service"

	"github.com/rs/zerolog"
	"google.golang.org/grpc"
)

func NewUserClient(addr string, log zerolog.Logger) userpb.UserServiceClient {
	conn, err := grpc.Dial(addr, grpc.WithInsecure())
	if err != nil {
		log.Fatal().Msgf("failed to connect to user-service: %v", err)
	}
	return userpb.NewUserServiceClient(conn)
}

func GetUserInfo(client userpb.UserServiceClient, uuid string) (*userpb.UserResponse, error) {
	ctx := context.Background()
	resp, err := client.AboutMeUser(ctx, &userpb.AboutMeRequest{Uuid: uuid})
	if err != nil {
		return nil, err
	}
	return resp, nil
}
