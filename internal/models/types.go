package models

import (
	"database/sql/driver"
	"encoding/json"
	"errors"
)

// JSON is a generic jsonb column backed by a Go map/slice value.
type JSON map[string]any

func (j JSON) Value() (driver.Value, error) { return json.Marshal(j) }
func (j *JSON) Scan(src any) error {
	if src == nil {
		*j = nil
		return nil
	}
	b, ok := src.([]byte)
	if !ok {
		return errors.New("JSON: type assertion to []byte failed")
	}
	return json.Unmarshal(b, j)
}

// StringSlice is a jsonb-backed []string (options, tags).
type StringSlice []string

func (s StringSlice) Value() (driver.Value, error) { return json.Marshal(s) }
func (s *StringSlice) Scan(src any) error {
	if src == nil {
		*s = nil
		return nil
	}
	b, ok := src.([]byte)
	if !ok {
		return errors.New("StringSlice: type assertion to []byte failed")
	}
	return json.Unmarshal(b, s)
}
