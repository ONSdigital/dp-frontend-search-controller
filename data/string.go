package data

import (
	"context"
	"fmt"
	"net/url"
	"regexp"

	"github.com/ONSdigital/log.go/v2/log"
)

const minQueryLength = 3

// reviewQueryString performs basic checks on the string entered by the user
func reviewQueryString(ctx context.Context, urlQuery url.Values) bool {
	validationProblem, err := checkForNonSpaceCharacters(ctx, urlQuery.Get("q"))
	if err != nil {
		log.Error(ctx, "query string did not have sufficient non-space characters", err)
	}
	return validationProblem
}

func checkForNonSpaceCharacters(ctx context.Context, queryString string) (bool, error) {
	var regexString string
	for i := 0; i < minQueryLength; i++ {
		regexString += `\S\s*`
	}

	match, err := regexp.MatchString(regexString, queryString)
	if err != nil {
		log.Error(ctx, "unable to check query string against regex", err)
		return false, err
	}

	if !match {
		log.Info(ctx, fmt.Sprintf("the query string did not match the regex, %v non-space characters required", minQueryLength))
		return true, nil
	}

	return false, nil
}
