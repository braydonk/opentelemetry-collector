// Copyright The OpenTelemetry Authors
// SPDX-License-Identifier: Apache-2.0

package builder

import (
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"

	"go.yaml.in/yaml/v3"
)

const (
	ocbPluginDirEnv     = "OCB_PLUGIN_DIR"
	ocbPluginDirDefault = ".ocb"
)

func getOCBPluginDir() (string, error) {
	// Check if user supplied plugin directory via environment variable.
	pluginDir := os.Getenv(ocbPluginDirEnv)
	if pluginDir != "" {
		return pluginDir, nil
	}

	homeDir, err := os.UserHomeDir()
	if err == nil {
		// If we can determine the home directory, set to $HOME/.ocb
		pluginDir = filepath.Join(homeDir, ocbPluginDirDefault)
	} else {
		// If not, install plugins to (working dir)/.ocb
		pluginDir = ocbPluginDirDefault
	}

	if err := os.MkdirAll(pluginDir, 0750); err != nil {
		return "", fmt.Errorf("failed to create plugin directory %q: %w", pluginDir, err)
	}
	return pluginDir, nil
}

type PluginCollection []PluginSourceConfig

func (pc PluginCollection) Validate() error {
	validationErrors := []error{}
	for _, pluginConfig := range pc {
		if err := pluginConfig.Validate(); err != nil {
			validationErrors = append(validationErrors, err)
		}
	}
	return errors.Join(validationErrors...)
}

func (pc PluginCollection) InstallAll(cfg *Config) (InstalledPlugins, error) {
	// Determine where to install plugins.
	pluginDir, err := getOCBPluginDir()
	if err != nil {
		return nil, err
	}

	pluginMap := InstalledPlugins{}

	for _, pluginConfig := range pc {
		// Don't reinstall installed plugins unless asked to.
		if cfg.ReinstallHooks || !pluginConfig.IsInstalled(pluginDir) {
			if err := pluginConfig.Install(cfg, pluginDir); err != nil {
				return nil, err
			}
		}
		// If plugin installs successfully, add it to the map of installed plugins
		// used by OCB build hooks.
		pluginMap.Add(pluginConfig.PluginName(), pluginConfig.InstallPath(pluginDir), cfg.Verbose)
	}

	return pluginMap, nil
}

// PluginSourceConfig is the source for a plugin to use in OCB build hooks.
type PluginSourceConfig struct {
	// GoMod is a Go Module URL and version to install a remote plugin.
	GoMod string `mapstructure:"gomod"`
	// GoPath is the path to a local plugin. Only used if `GoMod` is not set.
	GoPath string `mapstructure:"gopath"`
}

func (p PluginSourceConfig) Validate() error {
	if p.GoMod == "" && p.GoPath == "" {
		return errors.New("plugin is missing installable plugin source, you must set `gopath` or `gomod`")
	}
	return nil
}

func parseGoMod(gomod string) (string, string) {
	name, version, ok := strings.Cut(gomod, " ")
	if !ok {
		return gomod, "latest"
	}
	return name, version
}

var badCharacter = regexp.MustCompile(`[^a-zA-Z0-9.\-]`)

func (p PluginSourceConfig) PluginName() string {
	if p.GoMod != "" {
		return fmt.Sprintf("remote:%s", p.GoMod)
	} else if p.GoPath != "" {
		return fmt.Sprintf("local:%s", p.GoPath)
	} else {
		panic(fmt.Errorf("attempting to resolve plugin name for invalid plugin source config"))
	}
}

// BinaryName returns the binary name to use for the plugin
func (p PluginSourceConfig) BinaryName() string {
	var name string
	var version string
	if p.GoMod != "" {
		name, version = parseGoMod(p.GoMod)
	} else if p.GoPath != "" {
		name = p.GoPath
		version = "local"
	} else {
		panic(fmt.Errorf("attempting to resolve binary name for invalid plugin source config"))
	}

	// Espace name for use as file name
	name = badCharacter.ReplaceAllStringFunc(name, func(s string) string {
		return fmt.Sprintf("_%x_", s)
	})

	return fmt.Sprintf("%s_%s", name, version)
}

func (p PluginSourceConfig) InstallPath(pluginDir string) string {
	return filepath.Join(pluginDir, p.BinaryName())
}

func (p PluginSourceConfig) IsInstalled(pluginDir string) bool {
	_, err := os.Stat(p.InstallPath(pluginDir))
	// If we can successfully stat the plugin path then it's already installed.
	return err == nil
}

// Install installs the plugin binary into the specified pluginDir directory.
func (p PluginSourceConfig) Install(cfg *Config, pluginDir string) error {
	if p.GoPath == "" && p.GoMod == "" {
		return errors.New("either GoPath or GoMod must be specified to install plugin")
	}

	// Ensure we've been passed a valid plugin directory.
	info, err := os.Stat(pluginDir)
	if err != nil {
		if os.IsNotExist(err) {
			return fmt.Errorf("plugin directory %q does not exist", pluginDir)
		}
		return fmt.Errorf("failed to stat plugin directory %q: %w", pluginDir, err)
	}
	if !info.IsDir() {
		return fmt.Errorf("plugin directory path %q is not a directory", pluginDir)
	}

	absPluginDir, err := filepath.Abs(pluginDir)
	if err != nil {
		return fmt.Errorf("failed to get absolute path for plugin directory %q: %w", pluginDir, err)
	}

	var target string
	var workDir string
	if p.GoPath != "" {
		target = "."
		workDir = p.GoPath
	} else {
		mod, version := parseGoMod(p.GoMod)
		target = fmt.Sprintf("%s@%s", mod, version)
	}

	tempGOBIN, err := os.MkdirTemp(absPluginDir, ".tmp-install-*")
	if err != nil {
		return fmt.Errorf("failed to create temporary installation directory: %w", err)
	}
	defer os.RemoveAll(tempGOBIN)

	_, err = goCommand{
		dir: workDir,
		env: []string{"GOBIN=" + tempGOBIN},
	}.run(cfg, "install", target)
	if err != nil {
		return fmt.Errorf("failed to install plugin %q: %w", p.PluginName(), err)
	}

	entries, err := os.ReadDir(tempGOBIN)
	if err != nil {
		return fmt.Errorf("failed to read temporary installation directory: %w", err)
	}

	var installedBinary string
	for _, entry := range entries {
		if !entry.IsDir() {
			installedBinary = entry.Name()
			break
		}
	}

	if installedBinary == "" {
		return fmt.Errorf("no binary installed for plugin %q", p.PluginName())
	}

	src := filepath.Join(tempGOBIN, installedBinary)
	dst := p.InstallPath(absPluginDir)

	if err := os.Rename(src, dst); err != nil {
		return fmt.Errorf("failed to move installed plugin binary to %q: %w", dst, err)
	}

	return nil
}

type InstalledPlugin struct {
	path    string
	verbose bool
}

type inputData struct {
	Action string         `yaml:"action"`
	Config map[string]any `yaml:"config"`
}

func (ip *InstalledPlugin) run(action string, config map[string]any) error {
	input := inputData{Action: action, Config: config}
	inputBytes, err := yaml.Marshal(&input)
	if err != nil {
		return err
	}
	f, err := os.CreateTemp("", "ocb-plugin-input-*.yaml")
	defer f.Close()
	if err == nil {
		_, err = f.Write(inputBytes)
		if err == nil {
			err = f.Sync()
		}
	}
	if err != nil {
		return err
	}
	cmd := exec.Command(ip.path, f.Name())
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("error running hook: %v", err)
	}
	return nil
}

func (ip *InstalledPlugin) RunPreGenerate(data map[string]any) error {
	return ip.run("pre-generate", data)
}

func (ip *InstalledPlugin) RunPostGenerate(data map[string]any) error {
	return ip.run("post-generate", data)
}

func (ip *InstalledPlugin) RunPreBuild(data map[string]any) error {
	return ip.run("pre-build", data)
}

func (ip *InstalledPlugin) RunPostBuild(data map[string]any) error {
	return ip.run("post-build", data)
}

type InstalledPlugins map[string]*InstalledPlugin

func (ip InstalledPlugins) Add(name string, installPath string, verbose bool) {
	ip[name] = &InstalledPlugin{
		path:    installPath,
		verbose: verbose,
	}
}
