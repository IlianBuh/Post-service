package extraresources

import (
	"context"
	// "github.com/IlianBuh/Post-service/internal/domain/models"
)

type UserProvider interface {
	Exists(ctx context.Context, uuid int) (isExists bool, err error)
}
