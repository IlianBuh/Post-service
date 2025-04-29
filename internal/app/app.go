package app

import (
	"context"
	"log/slog"
	"sync"

	"github.com/IlianBuh/Post-service/internal/app/app"
	"github.com/IlianBuh/Post-service/internal/config/grpcobj"
	cfgStorage "github.com/IlianBuh/Post-service/internal/config/storage"
	cfgUsrPrvdr "github.com/IlianBuh/Post-service/internal/config/user-provider"
	"github.com/IlianBuh/Post-service/internal/service/event-worker"
	"github.com/IlianBuh/Post-service/internal/service/posts"
	"github.com/IlianBuh/Post-service/internal/storage/postgres"
	"github.com/IlianBuh/Post-service/internal/transport/kafka"
	userprovider "github.com/IlianBuh/Post-service/internal/transport/user-provider"
)

type App struct {
	log           *slog.Logger
	DB            *postgres.Storage
	EventWorker   *eventworker.Worker
	GRPCApp       *grpcapp.App
	EventProducer *kafka.Producer
	UserProvider  *userprovider.UserProvider
}

func New(
	log *slog.Logger,
	cfgGRPC grpcobj.GRPCObj,
	cfgStrg cfgStorage.Storage,
	cfgUsrPrvdr cfgUsrPrvdr.UserProvider,
	// cfg
) *App {
	const op = "app.New"
	fail := func(err error) {
		panic(op + err.Error())
	}
	// TODO : init storage
	repo, err := postgres.New(
		cfgStrg.User,
		cfgStrg.Password,
		cfgStrg.Host,
		cfgStrg.Port,
		cfgStrg.DBName,
		cfgStrg.Timeout,
	)
	if err != nil {
		fail(err)
	}

	// TODO : init user provider
	usrPrvdr, err := userprovider.New(
		log,
		cfgUsrPrvdr.Host,
		cfgUsrPrvdr.Port,
		cfgUsrPrvdr.Timeout.Duration,
	)
	if err != nil {
		fail(err)
	}

	postService := posts.New(
		log, repo, repo, repo, cfgGRPC.Timeout.Duration, usrPrvdr,
	)

	grpcapp := grpcapp.New(log, cfgGRPC.Port, postService, cfgGRPC.Timeout.Duration)

	// TODO : init kafka producer
	producer, err := kafka.NewProducer(
		context.Background(),
		log,
		cfgKafka.addrs,
		cfgKafka.Timeout,
		cfgKafka.Retries,
	)
	if err != nil {
		fail(op, err)
	}

	// TODO : init event-worker
	worker := eventworker.New(
		log,
		cfgEventWorker.PageSize,
		repo,
		repo,
		repo,
		cfgEventWorker.Inteval,
	)

	return &App{
		log:           log,
		UserProvider:  usrPrvdr,
		DB:            repo,
		GRPCApp:       grpcapp,
		EventWorker:   worker,
		EventProducer: producer,
	}
}

func (a *App) Start() {
	const op = "app.Start"
	log := a.log.With(slog.String("op", op))
	log.Info("starting application")

	a.EventWorker.Start(context.Background())

	go a.GRPCApp.MustRun()

	log.Info("application started")
}

func (a *App) Stop() {
	const op = "app.Stop"
	log := a.log.With(slog.String("op", op))
	log.Info("stopping application")

	var wg sync.WaitGroup

	wg.Add(5)
	go func() {
		defer wg.Done()
		a.EventProducer.Stop()
	}()
	go func() {
		defer wg.Done()
		a.EventWorker.Stop()
	}()
	go func() {
		defer wg.Done()
		a.DB.Stop()
	}()
	go func() {
		defer wg.Done()
		a.GRPCApp.Stop()
	}()
	go func() {
		defer wg.Done()
		a.UserProvider.Stop()
	}()

	wg.Wait()

	log.Info("application is stopped")
}
