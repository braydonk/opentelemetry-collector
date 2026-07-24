// Copyright The OpenTelemetry Authors
// SPDX-License-Identifier: Apache-2.0

package filter

import (
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"go.opentelemetry.io/collector/confmap"
	"go.opentelemetry.io/collector/confmap/confmaptest"
)

func readTestdataConfigYamls(t *testing.T, filename string) map[string][]Config {
	testFile := filepath.Join("testdata", filename)
	v, err := confmaptest.LoadConf(testFile)
	require.NoError(t, err)

	cfgs := map[string][]Config{}
	require.NoErrorf(t, v.Unmarshal(&cfgs, confmap.WithIgnoreUnused()), "unable to unmarshal yaml from file %v", testFile)
	return cfgs
}

func TestConfig(t *testing.T) {
	actualConfigs := readTestdataConfigYamls(t, "config.yaml")
	expectedConfigs := map[string][]Config{
		"regexp/default": {
			{
				Regexp: "one|two",
			},
		},
		"strict/default": {
			{
				Strict: "strict",
			},
		},
		"pattern/default": {
			{
				Pattern: "one*two",
			},
		},
	}

	for testName, actualCfg := range actualConfigs {
		t.Run(testName, func(t *testing.T) {
			expCfg, ok := expectedConfigs[testName]
			assert.True(t, ok)
			assert.Equal(t, expCfg, actualCfg)

			for _, cfg := range actualCfg {
				require.NoError(t, cfg.Validate())
			}
			fs := CreateFilter(actualCfg)
			assert.NotNil(t, fs)
		})
	}
}

func TestMatches(t *testing.T) {
	cfg := []Config{
		{
			Strict: "a",
		},
		{
			Strict: "b",
		},
		{
			Regexp: "a|b|c",
		},
		{
			Pattern: "x*z",
		},
		{
			Pattern: "f?o",
		},
		{
			Pattern: "test.num[0]",
		},
	}

	for _, c := range cfg {
		require.NoError(t, c.Validate())
	}
	fs := CreateFilter(cfg)

	assert.True(t, fs.Matches("a"))
	assert.True(t, fs.Matches("b"))
	assert.True(t, fs.Matches("c"))
	assert.True(t, fs.Matches("xz"))
	assert.True(t, fs.Matches("xyz"))
	assert.True(t, fs.Matches("xyabcz"))
	assert.True(t, fs.Matches("foo"))
	assert.True(t, fs.Matches("fao"))
	assert.True(t, fs.Matches("test.num[0]"))

	assert.False(t, fs.Matches("d"))
	assert.False(t, fs.Matches("fooo"))
	assert.False(t, fs.Matches("testXnum[0]"))
	assert.False(t, fs.Matches(123))
}

func TestConfigInvalid(t *testing.T) {
	actualConfigs := readTestdataConfigYamls(t, "config_invalid.yaml")
	expectedConfigs := map[string][]Config{
		"invalid/regexp": {
			{
				Regexp: "(.*[",
			},
		},
		"invalid/config_empty": {
			{
				Regexp: "",
				Strict: "",
			},
		},
		"invalid/config_both_set": {
			{
				Regexp: "1",
				Strict: "1",
			},
		},
		"invalid/pattern_and_strict": {
			{
				Pattern: "a*",
				Strict:  "a",
			},
		},
		"invalid/pattern_and_regexp": {
			{
				Pattern: "a*",
				Regexp:  "a.*",
			},
		},
		"invalid/pattern_strict_regexp": {
			{
				Pattern: "a*",
				Strict:  "a",
				Regexp:  "a.*",
			},
		},
	}

	for testName, actualCfg := range actualConfigs {
		t.Run(testName, func(t *testing.T) {
			expCfg, ok := expectedConfigs[testName]
			assert.True(t, ok)
			assert.Equal(t, expCfg, actualCfg)

			for _, cfg := range actualCfg {
				assert.Error(t, cfg.Validate())
			}
		})
	}
}

func TestWildcardPattern(t *testing.T) {
	tests := []struct {
		pattern  string
		input    string
		expected bool
	}{
		// Basic functionality.
		// No wildcards
		{pattern: "abc", input: "abc", expected: true},
		{pattern: "abc", input: "abcd", expected: false},
		// Asterisk (multi character wildcard)
		{pattern: "a*c", input: "ac", expected: true},
		{pattern: "a*c", input: "abbc", expected: true},
		{pattern: "a*c", input: "ab", expected: false},
		// Question mark (single character wildcard)
		{pattern: "a?c", input: "abc", expected: true},
		{pattern: "a?c", input: "ac", expected: false},
		{pattern: "a?c", input: "abbc", expected: false},
		// Treating any other regexp character as literal
		{pattern: "a.b", input: "a.b", expected: true},
		{pattern: "a.b", input: "axb", expected: false},
		{pattern: "a[b]", input: "a[b]", expected: true},
		{pattern: "a[b]", input: "ab", expected: false},
		{pattern: "a(b)", input: "a(b)", expected: true},
		{pattern: "a+b", input: "a+b", expected: true},
		{pattern: "a+b", input: "ab", expected: false},

		// Realistic metric name matching usage.
		{pattern: "http.server.request.count", input: "http.server.request.count", expected: true},
		{pattern: "http.server.request.count", input: "httpXserverXrequestXcount", expected: false},
		{pattern: "http.server.*", input: "http.server.request.duration", expected: true},
		{pattern: "http.server.*", input: "http.server.response.body.size", expected: true},
		{pattern: "http.server.*", input: "http.client.request.duration", expected: false},
		{pattern: "system.*.utilization", input: "system.memory.utilization", expected: true},
		{pattern: "system.*.utilization", input: "system.cpu.utilization", expected: true},
		{pattern: "system.*.utilization", input: "system.memory.usage", expected: false},
	}

	for _, tt := range tests {
		t.Run(tt.pattern+"_with_"+tt.input, func(t *testing.T) {
			re, err := WildcardCompile(tt.pattern)
			require.NoError(t, err)
			assert.Equal(t, tt.expected, re.MatchString(tt.input))

			mustRe := WildcardMustCompile(tt.pattern)
			assert.Equal(t, tt.expected, mustRe.MatchString(tt.input))
		})
	}
}
