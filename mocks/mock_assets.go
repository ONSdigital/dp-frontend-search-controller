package mocks

import "strings"

var cyLocale = []string{
	"[ValidationPatternMismatch]",
	"one = \"Enter a number\"",
	"[ValidationYearMissing]",
	"one = \"Enter a year\"",
	"[ValidationInvalidDate]",
	"one = \"Enter a real date\"",
}

var enLocale = []string{
	"[ValidationPatternMismatch]",
	"one = \"Enter a number\"",
	"[ValidationYearMissing]",
	"one = \"Enter a year\"",
	"[ValidationInvalidDate]",
	"one = \"Enter a real date\"",
}

func MockAssetFunction(name string) ([]byte, error) {
	if strings.Contains(name, ".cy.toml") {
		return []byte(strings.Join(cyLocale, "\n")), nil
	}
	return []byte(strings.Join(enLocale, "\n")), nil
}
