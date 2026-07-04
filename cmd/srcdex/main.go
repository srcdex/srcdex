// Command srcdex is a multi-repository code intelligence engine.
package main

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"

	"darvaza.org/sidecar/pkg/service"
)

// CmdName is the name of this executable.
const CmdName = "srcdex"

var rootCmd = &cobra.Command{
	Use:   CmdName,
	Short: "srcdex indexes source workspaces for hybrid search",
	Long: `srcdex is a pure-Go multi-repository code intelligence engine.
It indexes source workspaces for hybrid keyword and semantic search,
serving an embedded web UI for developers and a Model Context Protocol
(MCP) server for AI agents.`,
	Version: buildVersion(),
	Args:    cobra.NoArgs,

	SilenceErrors: true,
	SilenceUsage:  true,
}

// printF writes a formatted message to standard error, ignoring
// write errors.
func printF(format string, args ...any) {
	_, _ = fmt.Fprintf(os.Stderr, format, args...)
}

func main() {
	// nil serve: the service layer synthesises a not-implemented
	// stub until srcdex carries its own serve command.
	svc, err := service.Build(rootCmd, nil)
	if err != nil {
		printF("%s: %s\n", CmdName, err)
		os.Exit(service.ExitStatusMajor)
	}

	code, err := service.AsExitStatus(svc.Execute())
	if err != nil {
		printF("%s: %s\n", CmdName, err)
	}

	os.Exit(code)
}
