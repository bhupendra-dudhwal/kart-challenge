package utils

import "errors"

var (
	ErrDuplicateKey error = errors.New("document already exists")
	ErrNoData       error = errors.New("data does not exists")
)
