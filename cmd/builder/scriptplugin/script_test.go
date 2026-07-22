// Copyright The OpenTelemetry Authors
// SPDX-License-Identifier: Apache-2.0

package main

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestScriptPlugin_Run(t *testing.T) {
	tempDir := t.TempDir()
	outputFile := filepath.Join(tempDir, "output.txt")

	scriptPath := filepath.Join(tempDir, "test.sh")
	scriptContent := "#!/bin/bash\necho \"arg: $1, env: $TEST_VAR\" > " + outputFile + "\n"
	require.NoError(t, os.WriteFile(scriptPath, []byte(scriptContent), 0755))

	plugin := &ScriptPlugin{}

	// Test PreGenerate with path, args, and env
	err := plugin.PreGenerate(map[string]any{
		"path": scriptPath,
		"args": []string{"myarg"},
		"env": map[string]string{
			"TEST_VAR": "myenv",
		},
	})
	require.NoError(t, err)

	content, err := os.ReadFile(outputFile)
	require.NoError(t, err)
	assert.Equal(t, "arg: myarg, env: myenv\n", string(content))

	// Test missing path
	err = plugin.PreBuild(map[string]any{})
	require.Error(t, err)
	assert.Contains(t, err.Error(), "no script path specified")

	// Test non-existent script
	err = plugin.PostBuild(map[string]any{"path": filepath.Join(tempDir, "nonexistent.sh")})
	require.Error(t, err)
	assert.Contains(t, err.Error(), "failed")
}
