// Copyright The OpenTelemetry Authors
// SPDX-License-Identifier: Apache-2.0

package ocbplugin

import (
	"encoding/gob"
	"errors"
	"fmt"
	"net/rpc"
	"os"
	"os/exec"
	"strings"

	"github.com/hashicorp/go-hclog"
	"github.com/hashicorp/go-plugin"
	"golang.org/x/mod/semver"
)

func init() {
	gob.Register(map[string]any{})
	gob.Register([]any{})
	gob.Register(map[string]string{})
	gob.Register([]string{})
}

const PluginName = "ocbplugin"

var ocbPluginMagicValue = "ocb_build_hook"

// handshake is a common handshake config for OCB plugins.
var handshake = plugin.HandshakeConfig{
	ProtocolVersion:  1,
	MagicCookieKey:   "OCB_PLUGIN",
	MagicCookieValue: ocbPluginMagicValue,
}

// Serve starts a HashiCorp go-plugin server for the given OCBPlugin implementation.
func Serve(impl OCBPlugin) {
	logLevel := hclog.Off
	if envLevel := os.Getenv("HC_LOG"); envLevel != "" {
		logLevel = hclog.LevelFromString(envLevel)
	}
	logger := hclog.New(&hclog.LoggerOptions{
		Name:   "plugin",
		Level:  logLevel,
		Output: os.Stderr,
	})

	plugin.Serve(&plugin.ServeConfig{
		HandshakeConfig: handshake,
		Plugins: map[string]plugin.Plugin{
			PluginName: &ocbPluginRPC{
				Impl: impl,
			},
		},
		Logger: logger,
	})
}

// NewClient returns a new HashiCorp go-plugin Client configured for the OCB plugin at path.
func NewClient(path string, verbose bool) *plugin.Client {
	logLevel := hclog.Off
	if verbose {
		logLevel = hclog.Debug
	}

	logger := hclog.New(&hclog.LoggerOptions{
		Name:   "plugin",
		Level:  logLevel,
		Output: os.Stderr,
	})

	cmd := exec.Command(path)
	if verbose {
		cmd.Env = append(os.Environ(), "HC_LOG=DEBUG")
	} else {
		cmd.Env = append(os.Environ(), "HC_LOG=OFF")
	}

	return plugin.NewClient(&plugin.ClientConfig{
		HandshakeConfig: handshake,
		Plugins: map[string]plugin.Plugin{
			PluginName: &ocbPluginRPC{},
		},
		Cmd: cmd,
		AllowedProtocols: []plugin.Protocol{
			plugin.ProtocolNetRPC,
		},
		SyncStdout: os.Stdout,
		SyncStderr: os.Stderr,
		Logger:     logger,
	})
}

// OCBPlugin defines the interface that plugins must implement.
type OCBPlugin interface {
	PreGenerate(config map[string]any) error
	PostGenerate(config map[string]any) error
	PreBuild(config map[string]any) error
	PostBuild(config map[string]any) error
	MinOCBVersion() string
}

// ocbPluginRPC implements plugin.Plugin / plugin.NetRPCPlugin to connect/serve OCBPlugin over NetRPC.
type ocbPluginRPC struct {
	Impl OCBPlugin
}

func (p *ocbPluginRPC) Server(*plugin.MuxBroker) (interface{}, error) {
	return &ocbPluginServer{Impl: p.Impl}, nil
}

func (p *ocbPluginRPC) Client(_ *plugin.MuxBroker, c *rpc.Client) (interface{}, error) {
	return &ocbPluginClient{client: c}, nil
}

// ocbPluginServer is the net/rpc server that RPCClient talks to.
type ocbPluginServer struct {
	Impl OCBPlugin
}

func (s *ocbPluginServer) PreGenerate(config map[string]any, _ *struct{}) error {
	return s.Impl.PreGenerate(config)
}

func (s *ocbPluginServer) PostGenerate(config map[string]any, _ *struct{}) error {
	return s.Impl.PostGenerate(config)
}

func (s *ocbPluginServer) PreBuild(config map[string]any, _ *struct{}) error {
	return s.Impl.PreBuild(config)
}

func (s *ocbPluginServer) PostBuild(config map[string]any, _ *struct{}) error {
	return s.Impl.PostBuild(config)
}

func (s *ocbPluginServer) MinOCBVersion(_ struct{}, reply *string) error {
	*reply = s.Impl.MinOCBVersion()
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

func wrapRPCError(err error) error {
	if err == nil {
		return nil
	}
	msg := err.Error()
	switch {
	case strings.Contains(msg, ErrUnsupportedActionPreGenerate.Error()):
		return ErrUnsupportedActionPreGenerate
	case strings.Contains(msg, ErrUnsupportedActionPostGenerate.Error()):
		return ErrUnsupportedActionPostGenerate
	case strings.Contains(msg, ErrUnsupportedActionPreBuild.Error()):
		return ErrUnsupportedActionPreBuild
	case strings.Contains(msg, ErrUnsupportedActionPostBuild.Error()):
		return ErrUnsupportedActionPostBuild
	default:
		return err
	}
}

// ocbPluginClient is an implementation of OCBPlugin that talks over RPC.
type ocbPluginClient struct {
	client *rpc.Client
}

func (c *ocbPluginClient) PreGenerate(config map[string]any) error {
	var reply struct{}
	return wrapRPCError(c.client.Call("Plugin.PreGenerate", config, &reply))
}

func (c *ocbPluginClient) PostGenerate(config map[string]any) error {
	var reply struct{}
	return wrapRPCError(c.client.Call("Plugin.PostGenerate", config, &reply))
}

func (c *ocbPluginClient) PreBuild(config map[string]any) error {
	var reply struct{}
	return wrapRPCError(c.client.Call("Plugin.PreBuild", config, &reply))
}

func (c *ocbPluginClient) PostBuild(config map[string]any) error {
	var reply struct{}
	return wrapRPCError(c.client.Call("Plugin.PostBuild", config, &reply))
}

func (c *ocbPluginClient) MinOCBVersion() string {
	var reply string
	err := c.client.Call("Plugin.MinOCBVersion", struct{}{}, &reply)
	if err != nil {
		return ""
	}
	return reply
}

// validateVersion checks if hostVersion satisfies the minimum version required by minVersion.
func validateVersion(minVersion, hostVersion string) error {
	minV := ensureV(minVersion)
	hostV := ensureV(hostVersion)

	if !semver.IsValid(minV) {
		return fmt.Errorf("invalid minimum version %q declared by plugin", minVersion)
	}
	if !semver.IsValid(hostV) {
		return fmt.Errorf("invalid host OCB version %q", hostVersion)
	}
	if semver.Compare(hostV, minV) < 0 {
		return fmt.Errorf("plugin requires minimum OCB version %s, but host version is %s", minVersion, hostVersion)
	}
	return nil
}

func ensureV(v string) string {
	if v != "" && !strings.HasPrefix(v, "v") {
		return "v" + v
	}
	return v
}
