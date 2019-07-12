package goboilerplate

import "errors"

var (
	ErrNotFound = errors.New("Your requested item does not exists")
)
