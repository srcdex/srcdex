package main

import (
	"time"

	"darvaza.org/core"

	"github.com/spf13/cobra"
	"github.com/spf13/pflag"

	"srcdex.dev/pkg/embeddings/openai"
	"srcdex.dev/pkg/embeddings/vector"
)

// openaiCmd groups exploration tools for OpenAI-compatible
// embeddings services such as Ollama; it is not part of the stable
// surface.
var openaiCmd = &cobra.Command{
	Use:   "openai",
	Short: "Explores OpenAI-compatible embeddings services",
}

var openaiInfoCmd = &cobra.Command{
	Use:   "info",
	Short: "Reports the engine the flags select and the models it serves",
	Args:  cobra.NoArgs,
	RunE:  runOpenAIInfo,

	SilenceUsage: true,
}

var openaiEmbedCmd = &cobra.Command{
	Use:   "embed <text>...",
	Short: "Embeds the given texts and reports the vector shapes",
	Args:  cobra.MinimumNArgs(1),
	RunE:  runOpenAIEmbed,

	SilenceUsage: true,
}

var openaiModelsCmd = &cobra.Command{
	Use:   "models",
	Short: "Lists the models the service can serve",
	Args:  cobra.NoArgs,
	RunE:  runOpenAIModels,

	SilenceUsage: true,
}

const (
	baseURLFlag = "base-url"
	modelFlag   = "model"
	apiKeyFlag  = "api-key"
)

func init() {
	flags := openaiCmd.PersistentFlags()
	flags.String(baseURLFlag, "http://localhost:11434",
		"service base URL, without /v1")
	flags.String(modelFlag, "", "embedding model to run")
	flags.String(apiKeyFlag, "", "bearer token, when the service needs one")

	openaiCmd.AddCommand(openaiInfoCmd, openaiEmbedCmd, openaiModelsCmd)
	rootCmd.AddCommand(openaiCmd)
}

func newOpenAIClient(flags *pflag.FlagSet) (*openai.Client, error) {
	cfg := newConfig(flags)

	baseURL := core.Maybe(flags.GetString(baseURLFlag))
	apiKey := core.Maybe(flags.GetString(apiKeyFlag))

	return openai.New(openai.Config{
		Logger:  cfg.Logger,
		BaseURL: baseURL,
		APIKey:  apiKey,
	})
}

func runOpenAIInfo(cmd *cobra.Command, _ []string) error {
	client, err := newOpenAIClient(cmd.Flags())
	if err != nil {
		return err
	}
	defer func() { _ = client.Close() }()

	cmd.Printf("Engine: %s\n", client.Name())

	models, err := client.Models(cmd.Context())
	if err != nil {
		return err
	}

	cmd.Println("Models:")
	for _, m := range models {
		cmd.Printf("  %s\n", m.Name())
	}
	return nil
}

func runOpenAIEmbed(cmd *cobra.Command, args []string) error {
	client, err := newOpenAIClient(cmd.Flags())
	if err != nil {
		return err
	}
	defer func() { _ = client.Close() }()

	emb, err := client.New(core.Maybe(cmd.Flags().GetString(modelFlag)))
	if err != nil {
		return err
	}

	start := time.Now()
	vectors, err := emb.Embed(cmd.Context(), args)
	if err != nil {
		return err
	}
	elapsed := time.Since(start)

	for i, v := range vectors {
		cmd.Printf("%q: %v dimensions, L2 norm %.4f\n",
			args[i], len(v), vector.L2Norm(v))
	}
	cmd.Printf("%v texts in %s\n", len(vectors), elapsed)
	return nil
}

func runOpenAIModels(cmd *cobra.Command, _ []string) error {
	client, err := newOpenAIClient(cmd.Flags())
	if err != nil {
		return err
	}
	defer func() { _ = client.Close() }()

	models, err := client.Models(cmd.Context())
	if err != nil {
		return err
	}

	for _, m := range models {
		cmd.Println(m.Name())
	}
	return nil
}
