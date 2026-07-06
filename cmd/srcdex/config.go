package main

import (
	"github.com/spf13/pflag"

	"darvaza.org/slog"
)

// Config carries the application-level facts every command shares:
// the logger they report through.
type Config struct {
	// Logger is the interactive logger, honouring the -v flags.
	Logger slog.Logger
}

// newConfig assembles the shared facts from the command flags.
func newConfig(flags *pflag.FlagSet) *Config {
	return &Config{
		Logger: newLogger(flags),
	}
}
