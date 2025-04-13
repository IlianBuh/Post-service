package posts

import (
	"context"
	"fmt"
	"github.com/IlianBuh/Post-service/internal/lib/logger/sl"
	"log/slog"
	"time"
)

type PostService struct {
	log     *slog.Logger
	svr     Saver
	updtr   Updater
	dltr    Deleter
	timeout time.Duration
}

type Saver interface {
	// Save saves the record. Return values: postId, error
	Save(
		ctx context.Context,
		userId int,
		header string,
		contetn string,
		themes []string,
	) (int, error)
}
type Updater interface {
	// Update updates the record. Return values: postId, error
	Update(
		ctx context.Context,
		postId int,
		userId int,
		header string,
		contetn string,
		themes []string,
	) (int, error)
}
type Deleter interface {
	// Delete deletes the record. Return values: postId, error
	Delete(
		ctx context.Context,
		postId int,
		userId int,
	) (int, error)
}

func New(
	log *slog.Logger,
	svr Saver,
	updtr Updater,
	dltr Deleter,
	timeout time.Duration,
) *PostService {
	return &PostService{
		log:     log,
		svr:     svr,
		updtr:   updtr,
		dltr:    dltr,
		timeout: timeout,
	}
}

// Create creates new post
func (p *PostService) Create(
	ctx context.Context,
	userId int,
	header string,
	content string,
	themes []string,
) (int, error) {
	const op = "post-service.Create"
	log := p.log.With(slog.String("op", op))
	log.Info(
		"Starting create new post",
		slog.Int("user-id", userId),
		slog.String("header", header),
		slog.String("content", content),
		slog.Any("themes", themes),
	)
	var err error

	if err = ctx.Err(); err != nil {
		log.Error("failed to update - context is canceled", sl.Err(err))
		return 0, fmt.Errorf("%s: %w", op, ErrInternal)
	}

	ctx, cncl := context.WithTimeout(ctx, p.timeout)
	defer cncl()

	postId, err := p.svr.Save(ctx, userId, header, content, themes)
	if err != nil {
		log.Error("failed to save post", sl.Err(err))
		return 0, fmt.Errorf("%s: %w", op, ErrInternal)
	}

	log.Info("post is saved")
	return postId, nil
}

func (p *PostService) Update(
	ctx context.Context,
	postId int,
	userId int,
	header string,
	content string,
	themes []string,
) (int, error) {
	const op = "post-service.Update"
	log := p.log.With("op", op)
	log.Info(
		"Starting update post",
		slog.Int("post-id", postId),
		slog.Int("user-id", userId),
		slog.String("header", header),
		slog.String("content", content),
		slog.Any("themes", themes),
	)
	var err error

	if err = ctx.Err(); err != nil {
		log.Error("failed to update - context is canceled", sl.Err(err))
		return 0, fmt.Errorf("%s: %w", op, ErrInternal)
	}

	ctx, cncl := context.WithTimeout(ctx, p.timeout)
	defer cncl()

	postId, err = p.updtr.Update(ctx, postId, userId, header, content, themes)
	if err != nil {
		// TODO : check storage errors

		log.Error("failed to update record", sl.Err(err))
		return 0, fmt.Errorf("%s: %w", op, ErrInternal)
	}

	log.Info("post is updated")
	return postId, nil
}

func (p *PostService) Delete(
	ctx context.Context,
	postId int,
	userId int,
) (int, error) {
	const op = "post-service.Delete"
	log := p.log.With("op", op)
	log.Info("starting deleting post",
		slog.Int("post-id", postId),
		slog.Int("user-id", userId),
	)
	var err error

	if err = ctx.Err(); err != nil {
		log.Error("failed to update - context is canceled", sl.Err(err))
		return 0, fmt.Errorf("%s: %w", op, ErrInternal)
	}

	ctx, cncl := context.WithTimeout(ctx, p.timeout)
	defer cncl()

	postId, err = p.dltr.Delete(ctx, postId, userId)
	if err != nil {
		// TODO : handle storage errors

		log.Error("failed to delete post", sl.Err(err))
		return 0, fmt.Errorf("%s: %w", op, ErrInternal)
	}

	log.Info("post is deleted")
	return postId, nil
}
