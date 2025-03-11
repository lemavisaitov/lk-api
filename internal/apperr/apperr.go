package apperr

import (
	"errors"
)

type NotFoundError struct {
	Err error
}

func (e NotFoundError) Error() string {
	return e.Err.Error()
}

func (e NotFoundError) Is(target error) bool {
	// Сравниваем с целевой ошибкой
	var notFoundError NotFoundError
	ok := errors.As(target, &notFoundError)
	return ok
}

var ErrNotFound = NotFoundError{Err: errors.New("not found")}
