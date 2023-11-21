package kypo

import (
	"errors"
	"fmt"
)

var ErrNotFound = errors.New("not found")

type Error struct {
	ResourceName string
	Identifier   any
	Err          error
}

func (e *Error) Error() string {
	return fmt.Sprintf("resource %s %v: %s", e.ResourceName, e.Identifier, e.Err.Error())
}

func (e *Error) Unwrap() error {
	return e.Err
}
