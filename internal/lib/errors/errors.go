package errors

import (
	"fmt"
)

func Fail(op string, err error) error {
	return fmt.Errorf("%s: %w", op, err)
}
