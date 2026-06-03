package cmd

import (
	"fmt"
	"strings"

	"github.com/spf13/cobra"

	"github.com/NEPT-CLOUD/nept-cli-go/internal/app"
)

// HelloResponse is the structured JSON model returned when using json format.
type HelloResponse struct {
	Greeting string `json:"greeting"`
	Name     string `json:"name"`
}

// NewHelloCmd constructs the hello subcommand.
func NewHelloCmd(appContainer *app.App) *cobra.Command {
	var (
		greetName string
		uppercase bool
	)

	cmd := &cobra.Command{
		Use:   "hello",
		Short: "Say hello to someone",
		Long:  `A sample command that showcases flag parsing, configuration usage, structured logging, and JSON outputs.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			logger := appContainer.Logger
			cfg := appContainer.Config

			logger.Debug("Hello command called", "name", greetName, "uppercase", uppercase, "env", cfg.Environment)

			greeting := fmt.Sprintf("Hello, %s!", greetName)
			if uppercase {
				greeting = strings.ToUpper(greeting)
			}

			response := HelloResponse{
				Greeting: greeting,
				Name:     greetName,
			}

			return appContainer.PrintResult(greeting, response)
		},
	}

	// Define command-specific flags
	cmd.Flags().StringVarP(&greetName, "name", "n", "world", "The name of the entity to greet")
	cmd.Flags().BoolVarP(&uppercase, "uppercase", "u", false, "Convert greeting to uppercase")

	return cmd
}
