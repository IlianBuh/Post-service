package events

import (
	"fmt"
	"time"
)

const (
	TypeCteated = "created"
)

func CollectEventPayload(userId int, header string) string {
	return fmt.Sprintf(`{"id":%d, "header":"%s"}`, userId, header)
}

func CollectEventId(userId int) string {
	return fmt.Sprintf(`%d_%d`, userId, time.Now().Unix())
}
