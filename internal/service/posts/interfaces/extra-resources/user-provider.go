package extraresources

import (
	"context"

	"github.com/IlianBuh/Post-service/internal/domain/models"
)

type UserProvider interface {
	Users(ctx context.Context, uuids []int) (users []models.User, err error)
	User(ctx context.Context, uuid int) (user models.User, err error)
	Exists(ctx context.Context, uuid int) (isExists bool, err error)
}
