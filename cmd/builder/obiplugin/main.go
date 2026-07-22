// Copyright The OpenTelemetry Authors
// SPDX-License-Identifier: Apache-2.0

package main

import (
	"go.opentelemetry.io/collector/cmd/builder/ocbplugin"
)

func main() {
	ocbplugin.Serve(&OBIPlugin{})
}
