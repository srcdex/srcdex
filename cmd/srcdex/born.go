package main

import (
	"github.com/spf13/cobra"

	"srcdex.dev/pkg/embeddings/born"
)

// bornCmd groups exploration tools for the in-process Born engine;
// it is not part of the stable surface.
var bornCmd = &cobra.Command{
	Use:   "born",
	Short: "Explores the in-process Born embeddings engine",
}

var bornInfoCmd = &cobra.Command{
	Use:   "info",
	Short: "Reports the compute backend Born selects on this host",
	Args:  cobra.NoArgs,
	RunE:  runBornInfo,

	SilenceUsage: true,
}

func runBornInfo(cmd *cobra.Command, _ []string) error {
	eng, err := born.New()
	if err != nil {
		return err
	}
	defer func() { _ = eng.Close() }()

	cmd.Printf("Engine: %s\n", eng.Name())
	return nil
}

func init() {
	bornCmd.AddCommand(bornInfoCmd)
	rootCmd.AddCommand(bornCmd)
}
