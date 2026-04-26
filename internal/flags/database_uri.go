package flags

import (
	"errors"
)

type DatabaseURI string

func (d DatabaseURI) String() string {
	return string(d)
}

func (d *DatabaseURI) Set(s string) error {
	if s == "" {
		return errors.New("database URI is empty")
	}
	*d = DatabaseURI(s)
	return nil
}
