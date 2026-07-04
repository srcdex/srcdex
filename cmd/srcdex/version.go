package main

import (
	"runtime/debug"
)

// buildVersion reports the module version the Go toolchain recorded
// in the binary, keeping version stamping free of ldflags.
func buildVersion() string {
	if bi, ok := debug.ReadBuildInfo(); ok && bi.Main.Version != "" {
		return bi.Main.Version
	}
	return "unknown"
}
