package apperrors

import (
	"errors"
)

var (
	ErrInvalidFilter         = errors.New("invalid filter type given")
	ErrFilterNotFound        = errors.New("filter type from client not available in data.go")
	ErrInternalServer        = errors.New("internal server error")
	ErrInvalidPage           = errors.New("invalid page value, exceeding the default maximum search results")
	ErrPageExceedsTotalPages = errors.New("invalid page value, exceeding the total page value")

	BadRequestMap = map[error]bool{
		ErrInvalidFilter:         true,
		ErrInvalidPage:           true,
		ErrFilterNotFound:        true,
		ErrPageExceedsTotalPages: true,
	}
)
