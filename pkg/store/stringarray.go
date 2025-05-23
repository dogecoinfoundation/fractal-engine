package store

import (
	"database/sql/driver"
	"encoding/json"
	"errors"
)

type StringArray []string

// Scan implements the sql.Scanner interface
func (s *StringArray) Scan(value interface{}) error {
	bytes, ok := value.([]byte)
	if !ok {
		str, ok := value.(string)
		if !ok {
			return errors.New("type assertion to string failed")
		}
		bytes = []byte(str)
	}
	return json.Unmarshal(bytes, s)
}

// Value implements the driver.Valuer interface (optional if you want to write back to DB)
func (s StringArray) Value() (driver.Value, error) {
	return json.Marshal(s)
}
