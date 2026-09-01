// Copyright The OpenTelemetry Authors
// SPDX-License-Identifier: Apache-2.0

package main

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
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

	// Test empty path with user prompt providing a valid path
	promptOutputFile := filepath.Join(tempDir, "prompt_output.txt")
	promptScriptPath := filepath.Join(tempDir, "prompt_test.sh")
	promptScriptContent := "#!/bin/bash\necho \"prompted: $1\" > " + promptOutputFile + "\n"
	require.NoError(t, os.WriteFile(promptScriptPath, []byte(promptScriptContent), 0755))

	var stdout bytes.Buffer
	promptPlugin := &ScriptPlugin{
		stdin:  strings.NewReader(promptScriptPath + "\n"),
		stdout: &stdout,
	}
	err = promptPlugin.PostGenerate(map[string]any{
		"args": []string{"promptarg"},
	})
	require.NoError(t, err)
	assert.Equal(t, "Enter script path: ", stdout.String())

	promptContent, err := os.ReadFile(promptOutputFile)
	require.NoError(t, err)
	assert.Equal(t, "prompted: promptarg\n", string(promptContent))

	// Test empty path with empty user input
	stdout.Reset()
	emptyInputPlugin := &ScriptPlugin{
		stdin:  strings.NewReader("\n"),
		stdout: &stdout,
	}
	err = emptyInputPlugin.PreBuild(map[string]any{})
	require.Error(t, err)
	assert.Equal(t, "Enter script path: ", stdout.String())
	assert.Contains(t, err.Error(), "no script path provided")

	// Test non-existent script
	err = plugin.PostBuild(map[string]any{"path": filepath.Join(tempDir, "nonexistent.sh")})
	require.Error(t, err)
	assert.Contains(t, err.Error(), "failed")

	// Test MinOCBVersion
	assert.Equal(t, "0.157.0", plugin.MinOCBVersion())
}

func TestScriptPlugin_EmptyPathWithEnvConfig(t *testing.T) {
	tempDir := t.TempDir()
	outputFile := filepath.Join(tempDir, "output.txt")

	scriptPath := filepath.Join(tempDir, "test.sh")
	scriptContent := "#!/bin/bash\necho \"TEST=$TEST\" > " + outputFile + "\n"
	require.NoError(t, os.WriteFile(scriptPath, []byte(scriptContent), 0755))

	var stdout bytes.Buffer
	plugin := &ScriptPlugin{
		stdin:  strings.NewReader(scriptPath + "\n"),
		stdout: &stdout,
	}

	err := plugin.PreGenerate(map[string]any{
		"env": map[string]any{
			"TEST": 0,
		},
	})
	require.NoError(t, err)
	assert.Equal(t, "Enter script path: ", stdout.String())

	content, err := os.ReadFile(outputFile)
	require.NoError(t, err)
	assert.Equal(t, "TEST=0\n", string(content))
}
