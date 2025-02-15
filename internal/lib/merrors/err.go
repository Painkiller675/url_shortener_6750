package merrors

import "errors"

var (
	ErrURLNotFound      = errors.New("url not found")
	ErrURLOrAliasExists = errors.New("url or alias exists")
	ErrURLIsDel         = errors.New("url was deleted")
	ErrUserNotFound     = errors.New("user not found")
)
