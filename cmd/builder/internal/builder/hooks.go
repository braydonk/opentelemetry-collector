package builder

import (
	"errors"
	"fmt"
)

type HookAction int

const (
	HookActionPreGenerate HookAction = iota
	HookActionPostGenerate
	HookActionPreBuild
	HookActionPostBuild
)

type HookCollection []HookConfig

func (hc HookCollection) Validate() error {
	errs := []error{}
	for _, h := range hc {
		errs = append(errs, h.Validate())
	}
	return errors.Join(errs...)
}

func (hc HookCollection) RunAll(action HookAction, pluginMap InstalledPlugins) error {
	for _, hook := range hc {
		plugin, ok := pluginMap[hook.Plugin]
		if !ok {
			return fmt.Errorf("build hook requested unrecognized plugin \"%s\"", hook.Plugin)
		}

		var err error
		switch action {
		case HookActionPreGenerate:
			err = plugin.RunPreGenerate(hook.Config)
		case HookActionPostGenerate:
			err = plugin.RunPostGenerate(hook.Config)
		case HookActionPreBuild:
			err = plugin.RunPreBuild(hook.Config)
		case HookActionPostBuild:
			err = plugin.RunPostBuild(hook.Config)
		default:
			// Should be an impossible state
			panic("requested an unsupported hook action")
		}
		if err != nil {
			return err
		}
	}

	return nil
}

type HookConfig struct {
	Plugin string         `mapstructure:"plugin"`
	Config map[string]any `mapstructure:",remain"`
}

func (h HookConfig) Validate() error {
	if h.Plugin == "" {
		return errors.New("hook config missing required field `plugin`")
	}

	return nil
}
