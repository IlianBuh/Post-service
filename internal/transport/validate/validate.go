package validate

import (
	"fmt"
)

func Header(header string) error {
	if len(header) == 0 {
		return fmt.Errorf("%s", "header can't be empty")
	}

	return nil
}

func Id(id int64) error {
	if id < 0 {
		return fmt.Errorf("%s", "id must be natural number")
	}

	return nil
}
