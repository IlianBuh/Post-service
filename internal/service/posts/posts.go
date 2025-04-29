package posts

import (
	"context"
	"log/slog"
	"time"

	"errors"

	"github.com/IlianBuh/Post-service/internal/lib/logger/sl"
	extraresources "github.com/IlianBuh/Post-service/internal/service/posts/interfaces/extra-resources"
	"github.com/IlianBuh/Post-service/internal/service/posts/interfaces/repository"
	"github.com/IlianBuh/Post-service/internal/storage"
	errs "github.com/IlianBuh/Post-service/pkg/errors"
)

type PostService struct {
	log      *slog.Logger
	svr      repository.Saver
	updtr    repository.Updater
	dltr     repository.Deleter
	timeout  time.Duration
	usrPrvdr extraresources.UserProvider
}

func New(
	log *slog.Logger,
	svr repository.Saver,
	updtr repository.Updater,
	dltr repository.Deleter,
	timeout time.Duration,
	usrPrvdr extraresources.UserProvider,
) *PostService {
	return &PostService{
		log:      log,
		svr:      svr,
		updtr:    updtr,
		dltr:     dltr,
		timeout:  timeout,
		usrPrvdr: usrPrvdr,
	}
}

// Create creates new post and returns new posts' id or error.
// Only [ErrInternal] or [ErrUserNotFound] can be returned
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
	defer log.Info("creating post ended")

	var err error
	sendErr := func(err error) (int, error) {
		return 0, errs.Fail(op, err)
	}

	if err = ctx.Err(); err != nil {
		log.Error("failed to update - context is canceled", sl.Err(err))
		return sendErr(ctx.Err())
	}
	ctx, cncl := context.WithTimeout(ctx, p.timeout)
	defer cncl()

	err = p.checkUserExisting(ctx, userId)
	if err != nil {
		return sendErr(err)
	}

	postId, err := p.svr.Save(ctx, userId, header, content, themes)
	if err != nil {
		log.Error("failed to save post", sl.Err(err))
		return sendErr(ErrInternal)
	}

	log.Info("post is saved")
	return postId, nil
}

// Update updates post and returns posts' id, which must be equal
// to postId or error.
// Only [ErrInternal], [ErrNotCreator] or [ErrNotFound] can be returned as an error
func (p *PostService) Update(
	ctx context.Context,
	postId int,
	userId int,
	header string,
	content string,
	themes []string,
) (error) {
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
	defer log.Info("updating post ended")

	var err error
	sendErr := func(err error) (error) {
		return errs.Fail(op, err)
	}

	if err = ctx.Err(); err != nil {
		log.Error("failed to update - context is canceled", sl.Err(err))
		return sendErr(ErrInternal)
	}
	ctx, cncl := context.WithTimeout(ctx, p.timeout)
	defer cncl()

	postId, err = p.updtr.Update(ctx, postId, userId, header, content, themes)
	if err != nil {
		switch {
		case errors.Is(err, storage.ErrNotFound):
			log.Warn(
				"post with the id is not found",
				slog.Int("post-id", postId),
				sl.Err(err),
			)
			return sendErr(ErrNotFound)
		case errors.Is(err, storage.ErrNotCreator):
			log.Warn(
				"user is not creator of the post",
				slog.Int("post-id", postId),
				slog.Int("user-id", userId),
				sl.Err(err),
			)
			return sendErr(ErrNotCreator)
		}

		log.Error("failed to update record", sl.Err(err))
		return sendErr(ErrInternal)
	}

	return nil
}

// Delete deletes post with postId. Return posts' id which must be
// equal to postId or error.
// Only [ErrInternal] or [ErrNotCreator] can be returned as error
func (p *PostService) Delete(
	ctx context.Context,
	postId int,
	userId int,
) error {
	const op = "post-service.Delete"
	log := p.log.With("op", op)
	log.Info("starting deleting post",
		slog.Int("post-id", postId),
		slog.Int("user-id", userId),
	)
	defer log.Info("deleting ended")

	var err error
	sendErr := func(err error) error {
		return errs.Fail(op, err)
	}

	if err = ctx.Err(); err != nil {
		log.Error("failed to update - context is canceled", sl.Err(err))
		return sendErr(ErrInternal)
	}

	ctx, cncl := context.WithTimeout(ctx, p.timeout)
	defer cncl()

	err = p.dltr.Delete(ctx, postId, userId)
	if err != nil {
		if errors.Is(err, storage.ErrNotCreator) {
			log.Warn(
				"user is not creator of the post",
				slog.Int("post-id", postId),
				slog.Int("user-id", userId),
				sl.Err(err),
			)
			return sendErr(ErrNotCreator)
		}

		log.Error("failed to delete post", sl.Err(err))
		return sendErr(ErrInternal)
	}

	return nil
}

// checkUserExisting checks if user exists. If user does not exist,
// return error, otherwise return nil.
//
// It can return either [ErrInternal] or [ErrUserNotFound]
func (p *PostService) checkUserExisting(
	ctx context.Context,
	userId int,
) error {
	const op = "post-service.checkUserExisting"
	log := p.log.With(slog.String("op", op))
	log.Info("starting to check user existing")
	defer log.Info("checking ended")

	ok, err := p.usrPrvdr.Exists(ctx, userId)
	if err != nil {
		log.Error("failed to check users' existsing", sl.Err(err))
		return errs.Fail(op, ErrInternal)
	}
	if !ok {
		log.Warn(
			"user does not exist",
			slog.Int("uuid", userId),
		)
		return errs.Fail(op, ErrUserNotFound)
	}

	return nil
}
