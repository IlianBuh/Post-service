package repository

import (
	"context"
)

type Deleter interface {
	// Delete deletes the record. Return values: postId, error
	Delete(
		ctx context.Context,
		postId int,
		userId int,
	) (int, error)
}
