// Copyright The OpenTelemetry Authors
// SPDX-License-Identifier: Apache-2.0

package main

import (
	"fmt"

	"go.opentelemetry.io/collector/cmd/builder/ocbplugin"
)

type dummyPlugin struct{}

func (d *dummyPlugin) PreGenerate(config map[string]any) error {
	fmt.Printf("pre-generate:%v\n", config)
	return nil
}

func (d *dummyPlugin) PostGenerate(_ map[string]any) error {
	return ocbplugin.ErrUnsupportedActionPostGenerate
}

func (d *dummyPlugin) PreBuild(_ map[string]any) error {
	return ocbplugin.ErrUnsupportedActionPreBuild
}

func (d *dummyPlugin) PostBuild(_ map[string]any) error {
	return ocbplugin.ErrUnsupportedActionPostBuild
}

func (d *dummyPlugin) MinOCBVersion() string {
	return "0.151.0"
}

func main() {
	ocbplugin.RunPlugin(&dummyPlugin{})
}
