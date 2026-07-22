// Copyright The OpenTelemetry Authors
// SPDX-License-Identifier: Apache-2.0

package main

import (
	"errors"
	"fmt"
	"os"
	"os/exec"

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
type ScriptPlugin struct{}

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
	if err := mapstructure.Decode(config, &cfg); err != nil {
		return fmt.Errorf("failed to decode script plugin configuration: %w", err)
	}

	if cfg.Path == "" {
		return errors.New("no script path specified in configuration (`path` field required)")
	}

	cmdArgs := append([]string{cfg.Path}, cfg.Args...)
	cmd := exec.Command("bash", cmdArgs...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Stdin = os.Stdin

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
