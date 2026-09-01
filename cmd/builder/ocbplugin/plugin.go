// Copyright The OpenTelemetry Authors
// SPDX-License-Identifier: Apache-2.0

package ocbplugin

import (
	"errors"
	"flag"
	"fmt"
	"io"
	"os"

	"go.yaml.in/yaml/v3"
)

// OCBPlugin defines the interface that plugins must implement.
type OCBPlugin interface {
	PreGenerate(config map[string]any) error
	PostGenerate(config map[string]any) error
	PreBuild(config map[string]any) error
	PostBuild(config map[string]any) error
	MinOCBVersion() string
}

type inputData struct {
	Action string         `yaml:"action"`
	Config map[string]any `yaml:"config"`
}

// RunPlugin runs an OCBPlugin implementation. This should be called from main.
func RunPlugin(impl OCBPlugin) {
	flag.Parse()
	inputPath := flag.Arg(0)
	if err := runPlugin(impl, inputPath, os.Stdout); err != nil {
		_, _ = fmt.Fprintf(os.Stderr, "%s\n", err)
		os.Exit(1)
	}
	os.Exit(0)
}

func runPlugin(impl OCBPlugin, inputPath string, stdout io.Writer) error {
	inputBytes, err := os.ReadFile(inputPath)
	if err != nil {
		return fmt.Errorf("error reading plugin input: %w", err)
	}
	var input inputData
	if err := yaml.Unmarshal(inputBytes, &input); err != nil {
		return fmt.Errorf("error decoding plugin input: %w", err)
	}
	switch input.Action {
	case "min-ocb-version":
		fmt.Fprintln(stdout, impl.MinOCBVersion())
		return nil
	case "pre-generate":
		err = impl.PreGenerate(input.Config)
	case "post-generate":
		err = impl.PostGenerate(input.Config)
	case "pre-build":
		err = impl.PreBuild(input.Config)
	case "post-build":
		err = impl.PostBuild(input.Config)
	default:
		err = fmt.Errorf("unknown plugin action %q", input.Action)
	}
	if err != nil {
		return fmt.Errorf("error running '%s' plugin action: %w", input.Action, err)
	}
	return nil
}

var (
	// ErrUnsupportedActionPreGenerate is returned when a plugin does not support the PreGenerate lifecycle hook action.
	ErrUnsupportedActionPreGenerate = errors.New("pre_generate action not supported")

	// ErrUnsupportedActionPostGenerate is returned when a plugin does not support the PostGenerate lifecycle hook action.
	ErrUnsupportedActionPostGenerate = errors.New("post_generate action not supported")

	// ErrUnsupportedActionPreBuild is returned when a plugin does not support the PreBuild lifecycle hook action.
	ErrUnsupportedActionPreBuild = errors.New("pre_build action not supported")

	// ErrUnsupportedActionPostBuild is returned when a plugin does not support the PostBuild lifecycle hook action.
	ErrUnsupportedActionPostBuild = errors.New("post_build action not supported")
)
