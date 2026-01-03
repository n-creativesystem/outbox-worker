package rn

import (
	"errors"
	"strings"
)

const (
	rnDelimiter = ":"
	rnSections  = 6 // Match the number of sections in arn
	rnPrefix    = "rn:"
	arnPrefix   = "arn:"

	// sectionPartition = 1
	sectionService = 2
	// sectionRegion    = 3
	// sectionAccountID = 4
	sectionResource = 5

	// errors
	invalidPrefix   = "invalid prefix"
	invalidSections = "not enough sections"
)

func Parse(rn string) (RN, error) {
	if !isPrefix(rn) {
		return RN{}, errors.New(invalidPrefix)
	}
	sections := strings.SplitN(rn, rnDelimiter, rnSections)
	if len(sections) != rnSections {
		return RN{}, errors.New(invalidSections)
	}
	return RN{
		Service:  sections[sectionService],
		Resource: sections[sectionResource],
	}, nil
}

func isPrefix(rn string) bool {
	return strings.HasPrefix(rn, rnPrefix) || strings.HasPrefix(rn, arnPrefix)
}

func IsRN(rn string) bool {
	return isPrefix(rn) && strings.Count(rn, ":") >= rnSections-1
}

type RN struct {
	Service  string
	Resource string
}
