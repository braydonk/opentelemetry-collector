// Copyright The OpenTelemetry Authors
// SPDX-License-Identifier: Apache-2.0

package builder

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
)

func TestPluginSourceConfigInstall_MissingSource(t *testing.T) {
	cfg, err := NewDefaultConfig()
	require.NoError(t, err)
	pluginDir := t.TempDir()

	p := PluginSourceConfig{}

	err = p.Install(cfg, pluginDir)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "gomod must be specified")
}

func TestPluginSourceConfigInstall_NonExistentPluginDir(t *testing.T) {
	cfg, err := NewDefaultConfig()
	require.NoError(t, err)
	nonExistentDir := filepath.Join(t.TempDir(), "does_not_exist")

	p := PluginSourceConfig{
		GoMod: t.TempDir(),
	}

	err = p.Install(cfg, nonExistentDir)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "does not exist")
}

func TestPluginSourceConfigInstall_FromPath(t *testing.T) {
	pluginSourceDir := t.TempDir()
	goModContent := `module example.com/dummyplugin

go 1.25.0
`
	require.NoError(t, os.WriteFile(filepath.Join(pluginSourceDir, "go.mod"), []byte(goModContent), 0600))

	mainGoContent := `package main

import (
	"fmt"
	"io"
	"os"
)

func main() {
	input, _ := io.ReadAll(os.Stdin)
	if string(input) != "" {
		fmt.Printf("received:%s\n", string(input))
	}
}
`
	require.NoError(t, os.WriteFile(filepath.Join(pluginSourceDir, "main.go"), []byte(mainGoContent), 0600))

	outputDir := t.TempDir()
	pluginDir := filepath.Join(outputDir, "hooks")
	require.NoError(t, os.MkdirAll(pluginDir, 0750))

	cfg, err := NewDefaultConfig()
	require.NoError(t, err)
	cfg.Logger = zap.NewNop()
	cfg.Distribution.OutputPath = outputDir
	cfg.Distribution.Go = "go"

	p := PluginSourceConfig{
		GoMod: pluginSourceDir,
	}

	err = p.Install(cfg, pluginDir)
	require.NoError(t, err)

	expectedBinaryPath := p.InstallPath(pluginDir)
	info, err := os.Stat(expectedBinaryPath)
	require.NoError(t, err)
	assert.False(t, info.IsDir())

	// Test running the installed plugin via stdin
	ip := &InstalledPlugin{path: expectedBinaryPath}
	err = ip.RunPreGenerate(map[string]any{"test": "val"})
	require.NoError(t, err)
}

func TestParseRemoteGoMod(t *testing.T) {
	tests := []struct {
		input       string
		wantMod     string
		wantVersion string
	}{
		{
			input:       "github.com/org/repo@v1.2.3",
			wantMod:     "github.com/org/repo",
			wantVersion: "v1.2.3",
		},
		{
			input:       "github.com/org/repo v1.2.3",
			wantMod:     "github.com/org/repo",
			wantVersion: "v1.2.3",
		},
		{
			input:       "github.com/org/repo",
			wantMod:     "github.com/org/repo",
			wantVersion: "latest",
		},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			mod, version := parseRemoteGoMod(tt.input)
			assert.Equal(t, tt.wantMod, mod)
			assert.Equal(t, tt.wantVersion, version)
		})
	}
}

func TestHookSkipping(t *testing.T) {
	cfg, err := NewDefaultConfig()
	require.NoError(t, err)
	cfg.Logger = zap.NewNop()
	cfg.Hooks = &BuildHooks{
		PreGenerate:  HookCollection{{PluginSourceConfig: PluginSourceConfig{GoMod: "dummy"}}},
		PostGenerate: HookCollection{{PluginSourceConfig: PluginSourceConfig{GoMod: "dummy"}}},
		PreBuild:     HookCollection{{PluginSourceConfig: PluginSourceConfig{GoMod: "dummy"}}},
		PostBuild:    HookCollection{{PluginSourceConfig: PluginSourceConfig{GoMod: "dummy"}}},
	}

	plugins := InstalledPlugins{} // empty, so any run would fail with unrecognized plugin

	// When SkipGenerate is true, pre/post generate hooks must be skipped without error
	cfg.SkipGenerate = true
	assert.NoError(t, RunPreGenerateHooks(cfg, plugins))
	assert.NoError(t, RunPostGenerateHooks(cfg, plugins))

	// When SkipCompilation is true, pre/post build hooks must be skipped without error
	cfg.SkipCompilation = true
	assert.NoError(t, RunPreBuildHooks(cfg, plugins))
	assert.NoError(t, RunPostBuildHooks(cfg, plugins))
}

func TestPluginSourceConfigInstall_InvalidModule(t *testing.T) {
	outputDir := t.TempDir()
	pluginDir := filepath.Join(outputDir, "hooks")
	require.NoError(t, os.MkdirAll(pluginDir, 0750))

	cfg, err := NewDefaultConfig()
	require.NoError(t, err)
	cfg.Logger = zap.NewNop()
	cfg.Distribution.OutputPath = outputDir
	cfg.Distribution.Go = "go"

	p := PluginSourceConfig{
		GoMod: "example.com/nonexistent/module/url v1.0.0",
	}
	err = p.Install(cfg, pluginDir)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "failed to install plugin")
}

func TestInstalledPlugins_Add(t *testing.T) {
	ip := InstalledPlugins{}
	ip.Add("my-plugin", "/path/to/plugin", false)

	item, exists := ip["my-plugin"]
	require.True(t, exists)
	assert.Equal(t, "/path/to/plugin", item.path)
}

func TestGetOCBPluginDir(t *testing.T) {
	t.Run("environment variable set", func(t *testing.T) {
		customDir := filepath.Join(t.TempDir(), "custom_plugin_dir")
		t.Setenv("OCB_PLUGIN_DIR", customDir)

		dir, err := getOCBPluginDir()
		require.NoError(t, err)
		assert.Equal(t, customDir, dir)
		_, err = os.Stat(customDir)
		assert.True(t, os.IsNotExist(err), "should not automatically create directory when set via env var")
	})

	t.Run("default directory created", func(t *testing.T) {
		t.Setenv("OCB_PLUGIN_DIR", "")
		dir, err := getOCBPluginDir()
		require.NoError(t, err)
		info, err := os.Stat(dir)
		require.NoError(t, err)
		assert.True(t, info.IsDir())
	})
}
