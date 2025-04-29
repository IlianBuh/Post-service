package eventworker

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"time"

	"github.com/IlianBuh/Post-service/internal/domain/models"
	"github.com/IlianBuh/Post-service/internal/lib/logger/sl"
	"github.com/IlianBuh/Post-service/internal/lib/mapper"
	"github.com/IlianBuh/Post-service/internal/storage"
)

type PageProvider interface {
	EventPage(ctx context.Context, limit int) ([]models.Event, error)
}

type Deleter interface {
	DeleteEvent(ctx context.Context, ids []string) error
}

type Reserver interface {
	Reserve(ctx context.Context, ids []string) error
}

type Sender interface {
	Send(ctx context.Context, page []models.Event) error
}

type Worker struct {
	log          *slog.Logger
	pageSize     int
	pageProvider PageProvider
	deleter      Deleter
	reserver     Reserver
	sender       Sender
	stop         chan struct{}
	ticker       *time.Ticker
	timeout      time.Duration
}

func New(
	log *slog.Logger,
	pageSize int,
	pageProvider PageProvider,
	reserver Reserver,
	deleter Deleter,
	interval time.Duration,
) *Worker {
	return &Worker{
		log:          log,
		pageSize:     pageSize,
		pageProvider: pageProvider,
		deleter:      deleter,
		reserver:     reserver,
		stop:         make(chan struct{}),
	}
}

func (w *Worker) Start(ctx context.Context, interval time.Duration) error {
	const op = "eventworker.Start"
	log := w.log.With(slog.String("op", op))

	w.ticker = time.NewTicker(interval)

	go func() {
		defer func() {
			w.stop <- struct{}{}
		}()

		for {
			select {
			case <-w.stop:
				log.Info("stop signal is received")
				return
			default:
			}

			select {
			case <-w.stop:
				log.Info("stop signal is received")
				return
			case <-w.ticker.C:
			}

			err := w.handleEvents()
			if err != nil {
				log.Error("failed to handle events", sl.Err(err))
			}
		}

	}()

	return nil
}

func (w *Worker) Stop() {
	const op = "eventworker.Stop"
	w.log.Info("starting to stop worker", slog.String("op", op))

	w.stop <- struct{}{}
	<-w.stop

	close(w.stop)
}

func (w *Worker) handleEvents() error {
	const op = "eventworker.handleEvents"
	log := w.log.With(slog.String("op", op))
	log.Info("starting to handle events")

	ctx, cncl := context.WithTimeout(context.Background(), w.timeout)
	defer cncl()

	page, err := w.pageProvider.EventPage(ctx, w.pageSize)
	if err != nil {

		log.Error("failed to get event page", sl.Err(err))
		return fail(op, err)
	}

	ids := mapper.EventsToIds(page)

	err = w.reserver.Reserve(ctx, ids)
	if err != nil {
		if errors.Is(err, storage.ErrNoEvents) {
			log.Info("no new events")
			return nil
		}

		log.Info("failed to reserve events", sl.Err(err))
		return fail(op, err)
	}

	err = w.sender.Send(ctx, page)
	if err != nil {
		log.Error("failed to send events", sl.Err(err))
		return fail(op, err)
	}

	err = w.deleter.DeleteEvent(ctx, ids)
	if err != nil {
		log.Error("failed to delete events", sl.Err(err))
		return fail(op, err)
	}

	return nil
}

func fail(op string, err error) error {
	return fmt.Errorf("%s: %w", op, err)
}
