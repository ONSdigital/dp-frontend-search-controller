package data

import (
	"context"
	"net/url"
	"regexp"

	errs "github.com/ONSdigital/dp-frontend-search-controller/apperrors"
	"github.com/ONSdigital/log.go/v2/log"
)

var regexString = `^\s+$`

// reviewQueryString performs basic checks on the string entered by the user
func reviewQueryString(ctx context.Context, urlQuery url.Values) error {
	q := urlQuery.Get("q")

	nonSpaceCharErr := checkForNonSpaceCharacters(ctx, q)
	if nonSpaceCharErr != nil {
		return nonSpaceCharErr
	}

	specialCharErr := checkForSpecialCharacters(ctx, q)
	if specialCharErr != nil {
		return specialCharErr
	}

	return nil
}

func checkForNonSpaceCharacters(ctx context.Context, queryString string) error {
	match, err := regexp.MatchString(regexString, queryString)
	if err != nil {
		log.Error(ctx, "unable to check query string against regex", err)
		errVal := errs.ErrInvalidQueryString
		return errVal
	}

	if match {
		log.Warn(ctx, "the query string consists only of whitespace characters")
		errVal := errs.ErrQueryOnlyWhitespace
		return errVal
	}

	return nil
}

func checkForSpecialCharacters(ctx context.Context, str string) error {
	re := regexp.MustCompile("[[:^ascii:]]")

	match := re.MatchString(str)

	if match {
		log.Warn(ctx, "the query string contains special characters")
		errVal := errs.ErrInvalidQueryString
		return errVal
	}

	return nil
}
