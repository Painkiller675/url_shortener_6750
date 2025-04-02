// Package consists of the errors which could occur in the app.
package merrors

import "errors"

var ( // main errors
	// ErrURLNotFound - URL not found.
	ErrURLNotFound = errors.New("url not found")
	// ErrURLOrAliasExists - URL or alias don't exist.
	ErrURLOrAliasExists = errors.New("url or alias exists")
	// ErrURLIsDel - wanted URL was deleted.
	ErrURLIsDel = errors.New("url was deleted")
	// ErrUserNotFound - can't find the user in the database.
	ErrUserNotFound = errors.New("user not found")
)
