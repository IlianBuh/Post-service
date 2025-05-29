package mocks

import "context"

type UserMock struct{}

func (UserMock) Exists(ctx context.Context, uuid int) (isExists bool, err error) {
	return true, nil
}
