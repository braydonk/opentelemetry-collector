// Copyright The OpenTelemetry Authors
// SPDX-License-Identifier: Apache-2.0

package ocbplugin

import (
	"bytes"
	"errors"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type mockPlugin struct {
	preGenerateErr  error
	postGenerateErr error
	preBuildErr     error
	postBuildErr    error
	minVersion      string

	preGenerateCalled  bool
	postGenerateCalled bool
	preBuildCalled     bool
	postBuildCalled    bool
	lastConfig         map[string]any
}

func (m *mockPlugin) PreGenerate(config map[string]any) error {
	m.preGenerateCalled = true
	m.lastConfig = config
	return m.preGenerateErr
}

func (m *mockPlugin) PostGenerate(config map[string]any) error {
	m.postGenerateCalled = true
	m.lastConfig = config
	return m.postGenerateErr
}

func (m *mockPlugin) PreBuild(config map[string]any) error {
	m.preBuildCalled = true
	m.lastConfig = config
	return m.preBuildErr
}

func (m *mockPlugin) PostBuild(config map[string]any) error {
	m.postBuildCalled = true
	m.lastConfig = config
	return m.postBuildErr
}

func (m *mockPlugin) MinOCBVersion() string {
	return m.minVersion
}

func TestRunPlugin_LifecycleActions(t *testing.T) {
	tests := []struct {
		name     string
		action   string
		validate func(t *testing.T, m *mockPlugin)
	}{
		{
			name:   "pre-generate",
			action: "pre-generate",
			validate: func(t *testing.T, m *mockPlugin) {
				assert.True(t, m.preGenerateCalled)
				assert.Equal(t, "foo", m.lastConfig["key"])
			},
		},
		{
			name:   "post-generate",
			action: "post-generate",
			validate: func(t *testing.T, m *mockPlugin) {
				assert.True(t, m.postGenerateCalled)
				assert.Equal(t, "foo", m.lastConfig["key"])
			},
		},
		{
			name:   "pre-build",
			action: "pre-build",
			validate: func(t *testing.T, m *mockPlugin) {
				assert.True(t, m.preBuildCalled)
				assert.Equal(t, "foo", m.lastConfig["key"])
			},
		},
		{
			name:   "post-build",
			action: "post-build",
			validate: func(t *testing.T, m *mockPlugin) {
				assert.True(t, m.postBuildCalled)
				assert.Equal(t, "foo", m.lastConfig["key"])
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := &mockPlugin{}
			input := "action: " + tt.action + "\nconfig:\n  key: foo\n"
			var stdout bytes.Buffer
			err := runPlugin(m, strings.NewReader(input), &stdout)
			require.NoError(t, err)
			tt.validate(t, m)
		})
	}
}

func TestRunPlugin_MinOCBVersion(t *testing.T) {
	m := &mockPlugin{minVersion: "v0.157.0"}
	input := "action: min-ocb-version\n"
	var stdout bytes.Buffer
	err := runPlugin(m, strings.NewReader(input), &stdout)
	require.NoError(t, err)
	assert.Equal(t, "v0.157.0\n", stdout.String())
}

func TestRunPlugin_ActionError(t *testing.T) {
	m := &mockPlugin{preBuildErr: errors.New("custom pre-build error")}
	input := "action: pre-build\nconfig:\n  key: foo\n"
	var stdout bytes.Buffer
	err := runPlugin(m, strings.NewReader(input), &stdout)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "custom pre-build error")
}

func TestRunPlugin_UnknownAction(t *testing.T) {
	m := &mockPlugin{}
	input := "action: invalid-action\n"
	var stdout bytes.Buffer
	err := runPlugin(m, strings.NewReader(input), &stdout)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "unknown plugin action \"invalid-action\"")
}

func TestRunPlugin_InvalidYAML(t *testing.T) {
	m := &mockPlugin{}
	input := ": invalid: yaml: ["
	var stdout bytes.Buffer
	err := runPlugin(m, strings.NewReader(input), &stdout)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "error decoding plugin input")
}
