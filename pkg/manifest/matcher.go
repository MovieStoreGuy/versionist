package manifest

import "regexp"

type (
	// Matcher defines a shared interface
	// to allow various pattern match solutions
	Matcher interface {
		MatchString(str string) bool
	}

	matchString string
)

var (
	_ Matcher = (*matchString)(nil)
	_ Matcher = (*regexp.Regexp)(nil)
)

func (ms matchString) MatchString(str string) bool { return string(ms) == str }
