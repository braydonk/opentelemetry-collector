package ocbplugin

import "runtime/debug"

var moduleVersion = ""

func init() {
	if moduleVersion != "" {
		return
	}
	// Read the version of the ocbplugin module, which is versioned
	// in sync with the OCB binary. A plugin pulling in a particular
	// version of the `ocbplugin` package is expected to reflect
	// its support of the given version via its `MinOCBVersion` method.
	info, ok := debug.ReadBuildInfo()
	if ok {
		moduleVersion = info.Main.Version
	}
}
