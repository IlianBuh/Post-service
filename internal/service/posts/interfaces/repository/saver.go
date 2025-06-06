package repository

import (
	"context"
)

type Saver interface {
	// Save saves the record. Return values: postId, error
	Save(
		ctx context.Context,
		userId int,
		login string,
		header string,
		contetn string,
		themes []string,
	) (int, error)
}
