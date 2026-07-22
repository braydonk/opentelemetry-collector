// Copyright The OpenTelemetry Authors
// SPDX-License-Identifier: Apache-2.0

package builder

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/hashicorp/go-plugin"
	"go.opentelemetry.io/collector/cmd/builder/ocbplugin"
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
		pluginMap.Add(pluginConfig.Name, pluginConfig.InstallPath(pluginDir), cfg.Verbose)
	}

	return pluginMap, nil
}

// PluginSourceConfig is the source for a plugin to use in OCB build hooks.
type PluginSourceConfig struct {
	// Name is the user-decided name of the plugin to use within the config.
	Name string `mapstructure:"name"`

	// Module is a Go Module URL to install a remote plugin.
	Module string `mapstructure:"module"`
	// Version is the Go Module version to use for a remote plugin.
	Version string `mapstructure:"version"`

	// Path is the path to a local plugin. Only used if `Module` is not set.
	Path string `mapstructure:"path"`
}

func (p PluginSourceConfig) Validate() error {
	if p.Name == "" {
		return fmt.Errorf("missing plugin name")
	}

	if p.Module == "" && p.Path == "" {
		return fmt.Errorf("plugin %s is missing installable plugin source, you must set `path` or `module`", p.Name)
	}

	return nil
}

// BinaryName returns the binary name to use for the plugin, using fallback if Name is not specified.
func (p PluginSourceConfig) BinaryName() string {
	var version string
	if p.Module != "" {
		// When a version is not specified, the installation is done with `latest`.
		version = "latest"
		if p.Version != "" {
			version = p.Version
		}
	} else if p.Path != "" {
		// Locally installed plugins don't get a version discriminator.
		// We still need to disambiguate between a local copy of a plugin
		// and a potential to remotely install the same plugin.
		version = "local"
	}

	if version != "" {
		return fmt.Sprintf("%s_%s", p.Name, version)
	}
	if p.Name == "" {
		// Should be an impossible state.
		panic(fmt.Errorf("attempting to resolve binary name for invalid plugin source config"))
	}
	return p.Name
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
	if p.Path == "" && p.Module == "" {
		return errors.New("either Path or Module must be specified to install plugin")
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
	if p.Path != "" {
		target = "."
		workDir = p.Path
	} else {
		version := p.Version
		if version == "" {
			version = "latest"
		}
		if strings.Contains(p.Module, "@") {
			target = p.Module
		} else {
			target = fmt.Sprintf("%s@%s", p.Module, version)
		}
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
		return fmt.Errorf("failed to install plugin %q: %w", p.Name, err)
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
		return fmt.Errorf("no binary installed for plugin %q", p.Name)
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

	client         *plugin.Client
	pluginInstance ocbplugin.OCBPlugin
}

func (ip *InstalledPlugin) Start() error {
	client := ocbplugin.NewClient(ip.path, ip.verbose)

	rpcClient, err := client.Client()
	if err != nil {
		return fmt.Errorf("failed to create RPC client for plugin %q: %w", ip.path, err)
	}

	raw, err := rpcClient.Dispense(ocbplugin.PluginName)
	if err != nil {
		client.Kill()
		return fmt.Errorf("failed to dispense plugin %q: %w", ip.path, err)
	}

	p, ok := raw.(ocbplugin.OCBPlugin)
	if !ok {
		client.Kill()
		return fmt.Errorf("plugin %q does not implement OCBPlugin interface", ip.path)
	}

	ip.client = client
	ip.pluginInstance = p

	return nil
}

func (ip *InstalledPlugin) Stop() {
	if ip.client != nil {
		ip.client.Kill()
	}
}

func (ip *InstalledPlugin) RunPreGenerate(data map[string]any) error {
	if ip.pluginInstance == nil {
		if err := ip.Start(); err != nil {
			return err
		}
	}
	return ip.pluginInstance.PreGenerate(data)
}

func (ip *InstalledPlugin) RunPostGenerate(data map[string]any) error {
	if ip.pluginInstance == nil {
		if err := ip.Start(); err != nil {
			return err
		}
	}
	return ip.pluginInstance.PostGenerate(data)
}

func (ip *InstalledPlugin) RunPreBuild(data map[string]any) error {
	if ip.pluginInstance == nil {
		if err := ip.Start(); err != nil {
			return err
		}
	}
	return ip.pluginInstance.PreBuild(data)
}

func (ip *InstalledPlugin) RunPostBuild(data map[string]any) error {
	if ip.pluginInstance == nil {
		if err := ip.Start(); err != nil {
			return err
		}
	}
	return ip.pluginInstance.PostBuild(data)
}

type InstalledPlugins map[string]*InstalledPlugin

func (ip InstalledPlugins) Add(name, installPath string, verbose bool) {
	ip[name] = &InstalledPlugin{
		path:    installPath,
		verbose: verbose,
	}
}

func (ip InstalledPlugins) StopAll() {
	for _, p := range ip {
		p.Stop()
	}
}
