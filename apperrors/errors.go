package apperrors

import (
	"errors"
)

var (
	ErrFilterNotFound                   = errors.New("filter type from client not available in data.go")
	ErrInternalServer                   = errors.New("internal server error")
	ErrInvalidFilter                    = errors.New("invalid filter type given")
	ErrInvalidConentTypeAndTopicFilters = errors.New("invalid content type and topic filters")
	ErrInvalidPage                      = errors.New("invalid page value, exceeding the default maximum search results")
	ErrInvalidQueryString               = errors.New("the query string did not meet requirements")
	ErrInvalidQueryCharLengthString     = errors.New("the query string is less than the required character length")
	ErrPageExceedsTotalPages            = errors.New("invalid page value, exceeding the total page value")
	ErrTopicNotFound                    = errors.New("topic not found")

	BadRequestMap = map[error]bool{
		ErrFilterNotFound:        true,
		ErrInvalidFilter:         true,
		ErrInvalidPage:           true,
		ErrPageExceedsTotalPages: true,
		ErrInvalidQueryString:    true,
		ErrTopicNotFound:         true,
	}

	// ErrMapForRenderBeforeAPICalls is a list of errors which leads to the search page being rendered before making any API calls
	ErrMapForRenderBeforeAPICalls = map[error]bool{
		ErrInvalidQueryString: true,
		ErrTopicNotFound:      true,
	}
)
