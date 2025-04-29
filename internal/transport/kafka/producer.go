package kafka

import (
	"context"
	"fmt"
	"time"

	"log/slog"

	"github.com/IBM/sarama"
	"github.com/IlianBuh/Post-service/internal/domain/models"
	"github.com/IlianBuh/Post-service/internal/lib/logger/sl"
)

const (
	initialRetryTime = 1
)

var (
	eventsTopic = "events"
)

type Producer struct {
	log        *slog.Logger
	producer   sarama.AsyncProducer
	retries    int
	maxTimeout int
}

// NewProducer creates new kafka producer.
func NewProducer(
	ctx context.Context,
	log *slog.Logger,
	addrs []string,
	maxTimeout int,
	retries int,
) (*Producer, error) {
	const op = "kafka.NewProducer"
	cfg := sarama.NewConfig()
	cfg.Producer.Return.Errors = false
	cfg.Producer.Retry.Max = retries
	cfg.Producer.Timeout = time.Duration(maxTimeout * int(time.Second))

	fmt.Println("start to testing")
	p, err := sarama.NewAsyncProducer(
		addrs,
		cfg,
	)

	fmt.Println("start to testing")
	if err != nil {
		return tryToCreateProducer(ctx, log, addrs, cfg, maxTimeout, retries)
	}

	return &Producer{
		log:        log,
		producer:   p,
		retries:    retries,
		maxTimeout: maxTimeout,
	}, nil
}

// tryToCreateProducer tries to make producer instance
func tryToCreateProducer(
	ctx context.Context,
	log *slog.Logger,
	addrs []string,
	cfg *sarama.Config,
	maxTimeout, retries int,
) (*Producer, error) {
	const op = "kafka.tryToCreateProducer"
	var (
		err error
		p   sarama.AsyncProducer
	)
	timeout := initialRetryTime

	for retries > 0 {
		retries--
		if err = ctx.Err(); err != nil {
			return nil, fail(op, err)
		}

		p, err = sarama.NewAsyncProducer(
			addrs,
			cfg,
		)

		if err == nil {
			return &Producer{
				log:        log,
				producer:   p,
				retries:    retries,
				maxTimeout: maxTimeout,
			}, nil
		}

		time.Sleep(time.Duration(maxTimeout * int(time.Second)))
		timeout *= 2
		if timeout > maxTimeout {
			timeout = maxTimeout
		}
	}

	return nil, fail(op, err)
}

// Send sends page of events to kafka
func (p *Producer) Send(ctx context.Context, page []models.Event) error {
	const op = "producer.Send"
	log := p.log.With(slog.String("op", op))
	var (
		err error
		msg *sarama.ProducerMessage
	)
	ctx, cncl := context.WithCancel(ctx)
	defer cncl()

	for _, event := range page {
		eventmsg := event.Id + ":" + event.Payload
		msg = &sarama.ProducerMessage{
			Topic:     eventsTopic,
			Value:     sarama.ByteEncoder(eventmsg),
			Key:       sarama.ByteEncoder(event.Id),
			Timestamp: time.Now(),
		}

		select {
		case p.producer.Input() <- msg:
		case <-ctx.Done():
			log.Info("failed to send all messages", sl.Err(err))
			return fail(op, ctx.Err())
		}
	}

	log.Info("all events was sent successfully")
	return nil
}

// Stop stops kafka producer, but the first trying to send all messages
func (p *Producer) Stop() {
	const op = "producer.Stop"
	p.log.Info("starting to stop producer", slog.String("op", op))
	err := p.producer.Close()
	if err != nil {
		p.log.Error(
			"error during closing",
			slog.String("op", op),
			sl.Err(err),
		)
	}
}

func fail(op string, err error) error {
	return fmt.Errorf("%s: %w", op, err)
}
