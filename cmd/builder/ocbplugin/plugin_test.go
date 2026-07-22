// Copyright The OpenTelemetry Authors
// SPDX-License-Identifier: Apache-2.0

package ocbplugin

import (
	"errors"
	"testing"

	"github.com/hashicorp/go-plugin"
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

func TestOCBPluginRPC(t *testing.T) {
	mock := &mockPlugin{
		minVersion: "v0.120.0",
	}

	client, _ := plugin.TestPluginRPCConn(t, map[string]plugin.Plugin{
		PluginName: &ocbPluginRPC{Impl: mock},
	}, nil)

	raw, err := client.Dispense(PluginName)
	require.NoError(t, err)

	p, ok := raw.(OCBPlugin)
	require.True(t, ok)

	assert.Equal(t, "v0.120.0", p.MinOCBVersion())

	cfg := map[string]any{"key": "value"}
	err = p.PreBuild(cfg)
	require.NoError(t, err)
	assert.True(t, mock.preBuildCalled)
	assert.Equal(t, "value", mock.lastConfig["key"])

	err = p.PostBuild(cfg)
	require.NoError(t, err)
	assert.True(t, mock.postBuildCalled)
}

func TestOCBPluginRPCError(t *testing.T) {
	mock := &mockPlugin{
		preBuildErr: errors.New("prebuild failed"),
		minVersion:  "0.100.0",
	}

	client, _ := plugin.TestPluginRPCConn(t, map[string]plugin.Plugin{
		PluginName: &ocbPluginRPC{Impl: mock},
	}, nil)

	raw, err := client.Dispense(PluginName)
	require.NoError(t, err)

	p, ok := raw.(OCBPlugin)
	require.True(t, ok)

	err = p.PreBuild(map[string]any{})
	require.Error(t, err)
	assert.Contains(t, err.Error(), "prebuild failed")
}

func TestOCBPluginRPCStandardErrors(t *testing.T) {
	mock := &mockPlugin{
		preGenerateErr:  ErrUnsupportedActionPreGenerate,
		postGenerateErr: ErrUnsupportedActionPostGenerate,
		preBuildErr:     ErrUnsupportedActionPreBuild,
		postBuildErr:    ErrUnsupportedActionPostBuild,
	}

	client, _ := plugin.TestPluginRPCConn(t, map[string]plugin.Plugin{
		PluginName: &ocbPluginRPC{Impl: mock},
	}, nil)

	raw, err := client.Dispense(PluginName)
	require.NoError(t, err)

	p, ok := raw.(OCBPlugin)
	require.True(t, ok)

	err = p.PreGenerate(map[string]any{})
	require.Error(t, err)
	assert.True(t, errors.Is(err, ErrUnsupportedActionPreGenerate))

	err = p.PostGenerate(map[string]any{})
	require.Error(t, err)
	assert.True(t, errors.Is(err, ErrUnsupportedActionPostGenerate))

	err = p.PreBuild(map[string]any{})
	require.Error(t, err)
	assert.True(t, errors.Is(err, ErrUnsupportedActionPreBuild))

	err = p.PostBuild(map[string]any{})
	require.Error(t, err)
	assert.True(t, errors.Is(err, ErrUnsupportedActionPostBuild))
}

func TestValidateVersion(t *testing.T) {
	tests := []struct {
		name        string
		minVersion  string
		hostVersion string
		wantErr     bool
	}{
		{
			name:        "host satisfies min version",
			minVersion:  "v0.120.0",
			hostVersion: "v0.121.0",
			wantErr:     false,
		},
		{
			name:        "host equals min version",
			minVersion:  "0.120.0",
			hostVersion: "0.120.0",
			wantErr:     false,
		},
		{
			name:        "host below min version",
			minVersion:  "v0.121.0",
			hostVersion: "v0.120.0",
			wantErr:     true,
		},
		{
			name:        "invalid min version",
			minVersion:  "invalid",
			hostVersion: "v0.120.0",
			wantErr:     true,
		},
		{
			name:        "invalid host version",
			minVersion:  "v0.120.0",
			hostVersion: "invalid",
			wantErr:     true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validateVersion(tt.minVersion, tt.hostVersion)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}
