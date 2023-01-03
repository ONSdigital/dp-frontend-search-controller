package apperrors

import (
	"errors"
)

var (
	ErrContentTypeNotFound          = errors.New("content type not found")
	ErrInternalServer               = errors.New("internal server error")
	ErrInvalidPage                  = errors.New("invalid page value, exceeding the default maximum search results")
	ErrInvalidQueryString           = errors.New("the query string did not meet requirements")
	ErrInvalidQueryCharLengthString = errors.New("the query string is less than the required character length")
	ErrPageExceedsTotalPages        = errors.New("invalid page value, exceeding the total page value")
	ErrTopicNotFound                = errors.New("topic not found")

	BadRequestMap = map[error]bool{
		ErrContentTypeNotFound:   true,
		ErrInvalidPage:           true,
		ErrInvalidQueryString:    true,
		ErrPageExceedsTotalPages: true,
		ErrTopicNotFound:         true,
	}

	// ErrMapForRenderBeforeAPICalls is a list of errors which leads to the search page being rendered before making any API calls
	ErrMapForRenderBeforeAPICalls = map[error]bool{
		ErrInvalidQueryString: true,
	}
)
