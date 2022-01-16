package config

import (
	"encoding/json"
	"errors"
	"reflect"
	"time"

	"github.com/mitchellh/mapstructure"
)

type Duration struct {
	time.Duration
}

func (d Duration) MarshalJSON() ([]byte, error) {
	return json.Marshal(d.String())
}

func (d *Duration) UnmarshalJSON(b []byte) error {
	var v interface{}
	if err := json.Unmarshal(b, &v); err != nil {
		return err
	}

	switch value := v.(type) {
	case float64:
		d.Duration = time.Duration(value)
		return nil

	case string:
		var err error

		d.Duration, err = time.ParseDuration(value)
		if err != nil {
			return err
		}

		return nil

	default:
		return errors.New("invalid duration")
	}
}

// StringToCustomDurationHookFunc returns a DecodeHookFunc that converts
// strings to Duration.
func StringToCustomDurationHookFunc() mapstructure.DecodeHookFunc {
	return func(f reflect.Type, t reflect.Type, data interface{}) (interface{}, error) {
		if f.Kind() != reflect.String {
			return data, nil
		}

		if t != reflect.TypeOf(Duration{}) {
			return data, nil
		}

		// Convert it by parsing
		d, err := time.ParseDuration(data.(string))

		return Duration{d}, err
	}
}
