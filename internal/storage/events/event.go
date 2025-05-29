package events

import (
	"encoding/json"
	"fmt"
	"time"

	e "github.com/IlianBuh/Post-service/internal/lib/errors"
)

const (
	TypeCteated = "created"
)

type EventPayload struct {
	Author    Author    `json:"author"`
	Header    string    `json:"header"`
	CreatedAt time.Time `json:"created-at"`
}

type Author struct {
	Id    int    `json:"id"`
	Login string `json:"login"`
}

func CollectEventPayload(id int, login string, header string, createdAt time.Time) (string, error) {
	const op = "event.CollectEventPayload"

	payload, err := json.Marshal(
		EventPayload{
			Author{
				Id:    id,
				Login: login,
			},
			header,
			createdAt,
		},
	)
	if err != nil {
		return "", e.Fail(op, err)
	}

	return string(payload), nil
}

func CollectEventId(userId int) string {
	return fmt.Sprintf(`%d_%d`, userId, time.Now().Unix())
}
