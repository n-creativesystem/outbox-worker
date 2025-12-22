package errors

import "fmt"

type NotFoundKeyError struct {
	key string
}

func NewNotFoundKeyError(key string) NotFoundKeyError {
	return NotFoundKeyError{key: key}
}

func (e NotFoundKeyError) Error() string {
	return fmt.Errorf("not found %s", e.key).Error()
}
