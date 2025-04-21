package extraresources

import "context"

type FollowingsProvider interface {
	ListFollowers(ctx context.Context, uuid int) (uuids []int, error error)
}
