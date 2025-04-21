package repository

import (
	"context"
)

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
