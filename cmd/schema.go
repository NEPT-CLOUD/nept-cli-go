package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
	"github.com/spf13/pflag"

	"github.com/NEPT-CLOUD/nept-cli-go/internal/app"
)

// FlagSchema represents the machine-readable metadata of a command-line flag.
type FlagSchema struct {
	Name        string `json:"name"`
	Shorthand   string `json:"shorthand,omitempty"`
	Usage       string `json:"usage"`
	ValueType   string `json:"type"`
	Default     string `json:"default,omitempty"`
	Required    bool   `json:"required"`
}

// CommandSchema represents the machine-readable metadata of a CLI command and its children.
type CommandSchema struct {
	Name        string          `json:"name"`
	Use         string          `json:"use"`
	Short       string          `json:"short"`
	Long        string          `json:"long"`
	Usage       string          `json:"usage"`
	Example     string          `json:"example,omitempty"`
	Flags       []FlagSchema    `json:"flags,omitempty"`
	Subcommands []CommandSchema `json:"subcommands,omitempty"`
}

// NewSchemaCmd constructs the schema subcommand.
// It takes a pointer to the rootCmd to recursively introspect the entire CLI command structure.
func NewSchemaCmd(appContainer *app.App, rootCmd *cobra.Command) *cobra.Command {
	return &cobra.Command{
		Use:    "schema",
		Short:  "Print machine-readable schema of all CLI commands",
		Long:   `Generates a structured representation of all available commands, subcommands, flags, and arguments. Useful for AI agent tool integration.`,
		Hidden: false,
		RunE: func(cmd *cobra.Command, args []string) error {
			schema := getCommandSchema(rootCmd)

			// Generate text-based output for text format
			textTree := formatTextSchema(schema, "")
			return appContainer.PrintResult(textTree, schema)
		},
	}
}

// getCommandSchema recursively parses a *cobra.Command to build a CommandSchema tree.
func getCommandSchema(cmd *cobra.Command) CommandSchema {
	schema := CommandSchema{
		Name:    cmd.Name(),
		Use:     cmd.Use,
		Short:   cmd.Short,
		Long:    cmd.Long,
		Usage:   cmd.UseLine(),
		Example: cmd.Example,
	}

	cmd.LocalFlags().VisitAll(func(f *pflag.Flag) {
		schema.Flags = append(schema.Flags, FlagSchema{
			Name:      f.Name,
			Shorthand: f.Shorthand,
			Usage:     f.Usage,
			ValueType: f.Value.Type(),
			Default:   f.DefValue,
			Required:  f.Annotations != nil && len(f.Annotations[cobra.BashCompOneRequiredFlag]) > 0,
		})
	})

	for _, sub := range cmd.Commands() {
		// Skip hidden or built-in help commands to keep the schema clean
		if sub.Hidden || sub.Name() == "help" {
			continue
		}
		schema.Subcommands = append(schema.Subcommands, getCommandSchema(sub))
	}

	return schema
}

// formatTextSchema formats the schema into a clean, human-friendly hierarchical tree layout.
func formatTextSchema(schema CommandSchema, indent string) string {
	res := fmt.Sprintf("%s%s - %s\n", indent, schema.Name, schema.Short)
	
	// Print flags if they exist
	if len(schema.Flags) > 0 {
		flagIndent := indent + "    * "
		for _, f := range schema.Flags {
			shorthand := ""
			if f.Shorthand != "" {
				shorthand = fmt.Sprintf("-%s, ", f.Shorthand)
			}
			res += fmt.Sprintf("%s--%s (%s%s) - %s (default: %q)\n", flagIndent, f.Name, shorthand, f.ValueType, f.Usage, f.Default)
		}
	}

	for _, sub := range schema.Subcommands {
		res += formatTextSchema(sub, indent+"  ")
	}
	return res
}
