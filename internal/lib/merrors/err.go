package merrors

import "errors"

var (
	ErrURLNotFound      = errors.New("url not found")
	ErrURLExists        = errors.New("url exists")
	ErrURLOrAliasExists = errors.New("url or alias exists")
)
