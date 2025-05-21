package grpcserver

import (
	"context"
	"errors"

	"time"

	"github.com/IlianBuh/Post-service/internal/service/posts"
	"github.com/IlianBuh/Post-service/internal/transport/validate"
	postv1 "github.com/IlianBuh/Posts-Protobuf/gen/go"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type PostService interface {

	// Create creates new post
	Create(
		ctx context.Context,
		userId int,
		login string,
		header string,
		content string,
		themes []string,
	) (int, error)

	// Update updates all post fields with postId.
	// If the value is the default value (zero value), the old value will be saved.
	// User id is used to verify if  the user is a creator
	Update(
		ctx context.Context,
		userId int,
		postId int,
		header string,
		content string,
		themes []string,
	) error

	// Delete deletes all post related entries..
	// User id is used to verify if  the user is a creator
	Delete(
		ctx context.Context,
		postId int,
		userId int,
	) error
}

type ServerAPI struct {
	postv1.UnimplementedPostServer
	srvc    PostService
	timeout time.Duration
}

// Register registers serverAPI on srv grpc-server
func Register(srv grpc.ServiceRegistrar, post PostService, timeout time.Duration) {
	postv1.RegisterPostServer(srv, &ServerAPI{srvc: post, timeout: timeout})
}

// Create makes request to service layer to create a new post
func (s *ServerAPI) Create(ctx context.Context, req *postv1.CreateRequest) (*postv1.CreateResponse, error) {
	var err error
	if err = ctx.Err(); err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	if err = validate.Id(req.GetUserId()); err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}
	if err = validate.Header(req.GetHeader()); err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	ctx, cancel := context.WithTimeout(ctx, s.timeout)
	defer cancel()

	postId, err := s.srvc.Create(
		ctx,
		int(req.GetUserId()),
		req.GetLogin(),
		req.GetHeader(),
		req.GetContent(),
		req.GetThemes(),
	)
	if err != nil {
		if errors.Is(err, posts.ErrUserNotFound) {
			return nil, status.Error(codes.InvalidArgument, "user does not exist")
		}
		return nil, status.Error(codes.Internal, codes.Internal.String())
	}

	return &postv1.CreateResponse{PostId: int64(postId)}, nil
}

// Update makes request to service layer to change the existing post
func (s *ServerAPI) Update(ctx context.Context, req *postv1.UpdateRequest) (*postv1.UpdateResponse, error) {
	var err error
	if err = ctx.Err(); err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	if err = validate.Id(req.GetUserId()); err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}
	if err = validate.Id(req.GetPostId()); err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}
	if err = validate.Header(req.GetHeader()); err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	ctx, cnl := context.WithTimeout(ctx, s.timeout)
	defer cnl()

	err = s.srvc.Update(
		ctx,
		int(req.GetUserId()),
		int(req.GetPostId()),
		req.GetHeader(),
		req.GetContent(),
		req.GetThemes(),
	)
	if err != nil {
		// TODO : handle some errors

		return nil, status.Error(codes.Internal, "Intenal")
	}

	return &postv1.UpdateResponse{}, nil
}

// Delete makes request to service layer to delete the existing post
func (s *ServerAPI) Delete(ctx context.Context, req *postv1.DeleteRequest) (*postv1.DeleteResponse, error) {
	var err error
	if err = ctx.Err(); err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	if err = validate.Id(req.GetPostId()); err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}
	if err = validate.Id(req.GetUserId()); err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	ctx, cnl := context.WithTimeout(ctx, s.timeout)
	defer cnl()

	err = s.srvc.Delete(ctx, int(req.GetPostId()), int(req.GetUserId()))
	if err != nil {
		// TODO : handle some errors

		return nil, status.Error(codes.Internal, "Inernal")
	}

	return &postv1.DeleteResponse{}, nil
}
