package data

import (
	"context"
	"fmt"
	"net/url"
	"regexp"
	"strings"

	errs "github.com/ONSdigital/dp-frontend-search-controller/apperrors"
	"github.com/ONSdigital/log.go/v2/log"
)

// reviewQueryString performs basic checks on the string entered by the user
func reviewQueryString(ctx context.Context, urlQuery url.Values) error {
	return checkForNonSpaceCharacters(ctx, urlQuery.Get("q"))
}

func checkForNonSpaceCharacters(ctx context.Context, queryString string) error {
	const minQueryLength = 3
	var regexString = strings.Repeat(`\S\s*`, minQueryLength)

	match, err := regexp.MatchString(regexString, queryString)
	if err != nil {
		log.Error(ctx, "unable to check query string against regex", err)
		errVal := errs.ErrInvalidQueryString
		return errVal
	}

	if !match {
		log.Warn(ctx, fmt.Sprintf("the query string did not match the regex, %v non-space characters required", minQueryLength))
		errVal := errs.ErrInvalidQueryString
		return errVal
	}

	return nil
}
