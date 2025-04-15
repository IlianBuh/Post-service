package duration

import (
	"encoding/json"
	"time"
)

// Duration is custom duration type to unmarshal json configuration
type Duration struct {
	time.Duration
}

// UnmarshalJSON unmarshal data from json to duratoin. Only string format of json data is supported
func (d *Duration) UnmarshalJSON(b []byte) error {
	value := ""

	err := json.Unmarshal(b, &value)
	if err != nil {
		return err
	}

	d.Duration, err = time.ParseDuration(value)
	if err != nil {
		return err
	}

	return nil
}
