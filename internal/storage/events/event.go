package events

import (
	"encoding/json"
	"fmt"
	"time"

	e "github.com/IlianBuh/Post-service/pkg/errors"
)

const (
	TypeCteated = "created"
)

type EventPayload struct {
	UserId    int       `json:"user-id"`
	Header    string    `json:"header"`
	CreatedAt time.Time `json:"created-at"`
}

func CollectEventPayload(id int, header string, createdAt time.Time) (string, error) {
	const op = "event.CollectEventPayload"

	payload, err := json.Marshal(EventPayload{id, header, createdAt})
	if err != nil {
		return "", e.Fail(op, err)
	}

	return string(payload), nil
}

func CollectEventId(userId int) string {
	return fmt.Sprintf(`%d_%d`, userId, time.Now().Unix())
}
