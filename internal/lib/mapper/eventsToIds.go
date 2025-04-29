package mapper

import "github.com/IlianBuh/Post-service/internal/domain/models"

func EventsToIds(events []models.Event) []string {
	length := len(events)
	res := make([]string, length)

	for i, event := range events {
		res[i] = event.Id
	}

	return res
}
