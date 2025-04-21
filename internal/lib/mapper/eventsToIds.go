package mapper

import "github.com/IlianBuh/Post-service/internal/domain/models"

func EventsToIds(events []models.Event) []int {
	length := len(events)
	res := make([]int, length)

	for i := 0; i < length; i++ {
		res[i] = events[i].Id
	}

	return res
}
