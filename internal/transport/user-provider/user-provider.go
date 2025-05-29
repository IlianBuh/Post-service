package userprovider

import (
	"context"
	"fmt"
	"log/slog"

	"time"

	"github.com/IlianBuh/Post-service/internal/lib/errors"
	"github.com/IlianBuh/Post-service/internal/lib/logger/sl"
	userinfov1 "github.com/IlianBuh/SSO_Protobuf/gen/go/userinfo"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

type UserProvider struct {
	log        *slog.Logger
	timeout    time.Duration
	UserClient userinfov1.UserInfoClient
	connection *grpc.ClientConn
}

func New(
	log *slog.Logger,
	host string,
	port int,
	timeout time.Duration,
) (*UserProvider, error) {
	const op = "user-provider.New"

	cc, err := grpc.NewClient(
		composeServerAddress(host, port),
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	)
	if err != nil {
		return nil, errors.Fail(op, err)
	}

	c := userinfov1.NewUserInfoClient(cc)

	return &UserProvider{
		log:        log,
		timeout:    timeout,
		UserClient: c,
		connection: cc,
	}, nil
}

func composeServerAddress(host string, port int) string {
	return fmt.Sprintf("%s:%d", host, port)
}

func (u *UserProvider) Exists(ctx context.Context, uuid int) (isExists bool, err error) {
	const op = "user-provider.Exists"

	resp, err := u.UserClient.UsersExist(
		ctx,
		&userinfov1.UsersExistRequest{Uuid: []int32{int32(uuid)}},
	)
	if err != nil {
		return false, errors.Fail(op, err)
	}

	return resp.Exist, nil
}

func (u *UserProvider) Stop() {
	const op = "user-provider.Stop"
	log := u.log.With(slog.String("op", op))
	log.Info("stopping user-provider", slog.String("op", op))

	err := u.connection.Close()
	if err != nil {
		log.Error("failed to close connection", sl.Err(err))
		return
	}

	log.Info("connection is closed")
}
