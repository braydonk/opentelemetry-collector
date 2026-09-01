// Copyright The OpenTelemetry Authors
// SPDX-License-Identifier: Apache-2.0

package main

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"os"
	"os/exec"
	"strings"

	"github.com/go-viper/mapstructure/v2"
	"go.opentelemetry.io/collector/cmd/builder/ocbplugin"
)

var _ ocbplugin.OCBPlugin = (*ScriptPlugin)(nil)

// Config defines the configuration parameters for ScriptPlugin.
type Config struct {
	Path string            `mapstructure:"path"`
	Args []string          `mapstructure:"args"`
	Env  map[string]string `mapstructure:"env"`
}

// ScriptPlugin implements ocbplugin.OCBPlugin by executing a bash script.
type ScriptPlugin struct {
	stdin  io.Reader
	stdout io.Writer
	stderr io.Writer
}

func (s *ScriptPlugin) getStdin() io.Reader {
	if s.stdin != nil {
		return s.stdin
	}
	return os.Stdin
}

func (s *ScriptPlugin) getStdout() io.Writer {
	if s.stdout != nil {
		return s.stdout
	}
	return os.Stdout
}

func (s *ScriptPlugin) getStderr() io.Writer {
	if s.stderr != nil {
		return s.stderr
	}
	return os.Stderr
}

func (s *ScriptPlugin) PreGenerate(config map[string]any) error {
	return s.run(config)
}

func (s *ScriptPlugin) PostGenerate(config map[string]any) error {
	return s.run(config)
}

func (s *ScriptPlugin) PreBuild(config map[string]any) error {
	return s.run(config)
}

func (s *ScriptPlugin) PostBuild(config map[string]any) error {
	return s.run(config)
}

func (s *ScriptPlugin) MinOCBVersion() string {
	return "0.157.0"
}

func (s *ScriptPlugin) run(config map[string]any) error {
	var cfg Config
	if err := mapstructure.WeakDecode(config, &cfg); err != nil {
		return fmt.Errorf("failed to decode script plugin configuration: %w", err)
	}

	if cfg.Path == "" {
		fmt.Fprint(s.getStdout(), "Enter script path: ")
		scanner := bufio.NewScanner(s.getStdin())
		if scanner.Scan() {
			cfg.Path = strings.TrimSpace(scanner.Text())
		}
		if err := scanner.Err(); err != nil {
			return fmt.Errorf("failed to read script path: %w", err)
		}
		if cfg.Path == "" {
			return errors.New("no script path provided")
		}
	}

	cmdArgs := append([]string{cfg.Path}, cfg.Args...)
	cmd := exec.Command("bash", cmdArgs...)
	cmd.Stdout = s.getStdout()
	cmd.Stderr = s.getStderr()
	cmd.Stdin = s.getStdin()

	if len(cfg.Env) > 0 {
		cmd.Env = os.Environ()
		for k, v := range cfg.Env {
			cmd.Env = append(cmd.Env, fmt.Sprintf("%s=%s", k, v))
		}
	}

	if err := cmd.Run(); err != nil {
		return fmt.Errorf("script %q failed: %w", cfg.Path, err)
	}

	return nil
}
