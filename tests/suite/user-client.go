package suite

import (
	userinfov1 "github.com/IlianBuh/SSO_Protobuf/gen/go/userinfo"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

type UserClient struct {
	Client userinfov1.UserInfoClient
}

func NewUserClient(addr string) (*UserClient, error) {
	cc, err := grpc.NewClient(
		addr,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	)
	if err != nil {
		return nil, err
	}

	client := userinfov1.NewUserInfoClient(cc)

	return &UserClient{
		Client: client,
	}, nil
}
