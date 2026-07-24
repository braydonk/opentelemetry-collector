// Copyright The OpenTelemetry Authors
// SPDX-License-Identifier: Apache-2.0

package filter // import "go.opentelemetry.io/collector/filter"

import (
	"fmt"
	"regexp"
	"strings"
)

func validateConfig(c *Config) error {
	if err := exactlyOneSet(c.Strict, c.Regexp, string(c.Pattern)); err != nil {
		return fmt.Errorf("only one of string, regex, or pattern are allowed: %w", err)
	}

	if c.Regexp != "" {
		_, err := regexp.Compile(c.Regexp)
		if err != nil {
			return err
		}
	}

	if c.Pattern != "" {
		_, err := WildcardCompile(c.Pattern)
		if err != nil {
			return err
		}
	}

	return nil
}

func exactlyOneSet(values ...string) error {
	countSet := 0
	for _, value := range values {
		if value != "" {
			countSet++
		}
	}
	if countSet != 1 {
		return fmt.Errorf("%d values were set", countSet)
	}
	return nil
}

type combinedFilter struct {
	stricts map[any]struct{}
	regexes []*regexp.Regexp
}

// CreateFilter creates a Filter out of a set of Config configuration objects.
func CreateFilter(configs []Config) Filter {
	cf := &combinedFilter{
		stricts: make(map[any]struct{}),
	}
	for _, config := range configs {
		if config.Strict != "" {
			cf.stricts[config.Strict] = struct{}{}
			continue
		}

		var re *regexp.Regexp
		if config.Regexp != "" {
			// Validate() call above ensures that the regex is valid.
			re = regexp.MustCompile(config.Regexp)
		} else if config.Pattern != "" {
			// Validate() call above ensures that the pattern is valid.
			re = WildcardMustCompile(config.Pattern)
		}

		cf.regexes = append(cf.regexes, re)
	}
	return cf
}

func (cf *combinedFilter) Matches(toMatch any) bool {
	_, ok := cf.stricts[toMatch]
	if ok {
		return ok
	}
	if str, ok := toMatch.(string); ok {
		for _, re := range cf.regexes {
			if re.MatchString(str) {
				return true
			}
		}
	}
	return false
}

func WildcardMustCompile(wp string) *regexp.Regexp {
	return regexp.MustCompile(wildcardAsRegexpString(wp))
}

func WildcardCompile(wp string) (*regexp.Regexp, error) {
	return regexp.Compile(wildcardAsRegexpString(wp))
}

func wildcardAsRegexpString(wp string) string {
	escapedWp := regexp.QuoteMeta(string(wp))
	escapedWp = "^" + escapedWp + "$"
	escapedWp = strings.ReplaceAll(escapedWp, `\?`, ".")
	escapedWp = strings.ReplaceAll(escapedWp, `\*`, ".*")
	return escapedWp
}
