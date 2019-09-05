package goboilerplate

import (
	"fmt"

	"github.com/pkg/errors"
)

// ErrNotFound is used when entry not found.
var ErrNotFound = errors.New("not found")

// ConstraintError is used when a certain constraint is broken, i.e. on validation user input.
type ConstraintError string

func (e ConstraintError) Error() string {
	return string(e)
}

// ConstraintErrorf creates new interface value of type error
func ConstraintErrorf(text string, value ...interface{}) ConstraintError {
	return ConstraintError(fmt.Sprintf(text, value...))
}

// UnauthorizedError is used when user is not authenticated.
type UnauthorizedError string

func (e UnauthorizedError) Error() string {
	return string(e)
}

// UnauthorizedErrorf constructs UnauthorizedError with formatted message.
func UnauthorizedErrorf(format string, a ...interface{}) UnauthorizedError {
	return UnauthorizedError(fmt.Sprintf(format, a...))
}
