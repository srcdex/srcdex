package main

// cspell:words pflag zerolog slogzerolog

import (
	"os"

	"github.com/rs/zerolog"
	"github.com/spf13/pflag"

	"darvaza.org/core"
	"darvaza.org/slog"
	slogzerolog "darvaza.org/slog/handlers/zerolog"
)

const (
	verbosityFlag      = "verbose"
	verbosityShortFlag = "v"
)

func init() {
	pFlags := rootCmd.PersistentFlags()
	pFlags.CountP(verbosityFlag, verbosityShortFlag, "increase verbosity")
}

// logLevel maps the -v count onto the zerolog scale, from errors
// only up to debug.
func logLevel(flags *pflag.FlagSet) zerolog.Level {
	switch core.Maybe(flags.GetCount(verbosityFlag)) {
	case 0:
		return zerolog.ErrorLevel
	case 1:
		return zerolog.WarnLevel
	case 2:
		return zerolog.InfoLevel
	default:
		return zerolog.DebugLevel
	}
}

// newLogger assembles the interactive logger, writing to standard
// error and honouring the -v verbosity flags.
func newLogger(flags *pflag.FlagSet) slog.Logger {
	w := zerolog.NewConsoleWriter(func(cw *zerolog.ConsoleWriter) {
		cw.Out = os.Stderr
		cw.PartsExclude = []string{zerolog.TimestampFieldName}
	})
	zl := zerolog.New(w).Level(logLevel(flags))
	return slogzerolog.New(&zl)
}
