package merrors

import "errors"

var (
	ErrURLNotFound      = errors.New("url not found")
	ErrURLOrAliasExists = errors.New("url or alias exists")
	ErrALiasNotFound    = errors.New("alias not found")
)
