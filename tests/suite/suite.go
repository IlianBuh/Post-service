package suite

import (
	"context"
	"log/slog"
	"os"
	"testing"

	"github.com/IlianBuh/Post-service/internal/config"
	eventworker "github.com/IlianBuh/Post-service/internal/service/event-worker"
	"github.com/IlianBuh/Post-service/internal/service/posts"
	"github.com/IlianBuh/Post-service/internal/storage/postgres"
	"github.com/IlianBuh/Post-service/internal/transport/kafka"
	"github.com/IlianBuh/Post-service/tests/mocks"
)

type Suite struct {
	Post *posts.PostService
	ctx  context.Context
}

func NewSuite(t *testing.T, cfg *config.Config) *Suite {
	t.Parallel()
	t.Helper()

	cfgStrg := cfg.Storage
	repo, err := postgres.New(
		cfgStrg.User,
		cfgStrg.Password,
		cfgStrg.Host,
		cfgStrg.Port,
		cfgStrg.DBName,
		cfgStrg.Timeout,
	)
	if err != nil {
		t.Fatalf("failed to create storage: %v", err)
	}

	// TODO : init user provider
	usrPrvdr := mocks.UserMock{}

	postService := posts.New(
		slog.New(
			slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelDebug}),
		), repo, repo, repo, cfg.GRPC.Timeout.Duration, usrPrvdr,
	)

	// TODO : init kafka producer
	cfgKafka := cfg.Kafka
	producer, err := kafka.NewProducer(
		context.Background(),
		slog.New(
			slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelDebug}),
		),
		cfgKafka.Addrs,
		cfgKafka.Timeout,
		cfgKafka.Retries,
	)
	if err != nil {
		t.Fatalf("failed to create storage: %v", err)
	}

	// TODO : init event-worker
	worker := eventworker.New(
		slog.New(
			slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelDebug}),
		),
		cfg.EventWorker.PageSize,
		repo,
		repo,
		repo,
		producer,
		cfg.EventWorker.Interval.Duration,
	)

	worker.Start(context.Background())

	s := &Suite{
		Post: postService,
		ctx:  context.Background(),
	}

	return s
}
