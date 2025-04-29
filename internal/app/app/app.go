package grpcapp

import (
	"fmt"
	"log/slog"
	"net"
	"time"

	"github.com/IlianBuh/Post-service/internal/service/posts"
	grpcserver "github.com/IlianBuh/Post-service/internal/transport/grpc-server"
	"github.com/IlianBuh/Post-service/pkg/errors"
	"github.com/grpc-ecosystem/go-grpc-middleware/v2/interceptors/recovery"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type App struct {
	log      *slog.Logger
	port     int
	grpcsrvr *grpc.Server
}

func New(
	log *slog.Logger,
	port int,
	post *posts.PostService,
	timeout time.Duration,
) *App {
	recoveryOpt := []recovery.Option{
		recovery.WithRecoveryHandler(
			func(p any) error {
				log.Error("recover panic", slog.Any("panic", p))

				return status.Errorf(codes.Internal, "internal error")
			},
		),
	}

	grpcsrvr := grpc.NewServer(
		grpc.ChainUnaryInterceptor(
			recovery.UnaryServerInterceptor(recoveryOpt...),
		),
	)

	grpcserver.Register(grpcsrvr, post, timeout)

	return &App{
		log:      log,
		port:     port,
		grpcsrvr: grpcsrvr,
	}
}

func (a *App) MustRun() {
	if err := a.Run(); err != nil {
		panic("failed to run application: " + err.Error())
	}
}

func (a *App) Run() error {
	const op = "grpcapp.Run"
	log := a.log.With(slog.String("op", op))
	log.Info("starting application")

	l, err := net.Listen("tcp", fmt.Sprintf(":%d", a.port))
	if err != nil {
		return errors.Fail(op, err)
	}

	if err = a.grpcsrvr.Serve(l); err != nil {
		return errors.Fail(op, err)
	}

	return nil
}

func (a *App) Stop() {
	const op = "grpcapp.Stop"

	a.log.Info("stop grpc application", slog.String("op", op))

	a.grpcsrvr.GracefulStop()

	a.log.Info("grpc appliaction stopped", slog.String("op", op))
}
