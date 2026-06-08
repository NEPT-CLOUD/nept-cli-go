package cmd

import (
	"errors"
	"fmt"
	"os"

	"github.com/spf13/cobra"

	"github.com/NEPT-CLOUD/nept-cli-go/internal/app"
	"github.com/NEPT-CLOUD/nept-cli-go/internal/config"
	"github.com/NEPT-CLOUD/nept-cli-go/internal/logger"
)

// NewRootCmd creates the root command for the CLI.
// It accepts the application container `appContainer` where initialized config, logger, and streams will be injected.
func NewRootCmd(appContainer *app.App) *cobra.Command {
	var (
		cfgFile string
		verbose bool
		format  string
	)

	rootCmd := &cobra.Command{
		Use:   "nept",
		Short: "nept is a scalable command-line tool",
		Long: `A highly scalable, modern CLI boilerplate built in Go.
Supports configuration files, environment variables, subcommands, and structured logging.`,
		PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
			// 1. Load config
			cfg, err := config.Load(cfgFile)
			if err != nil {
				return fmt.Errorf("failed to load configuration: %w", err)
			}

			// Override config values if command-line flags are explicitly set
			if cmd.Flags().Changed("verbose") {
				cfg.Verbose = verbose
			}
			if cmd.Flags().Changed("format") {
				cfg.Format = format
			}

			// 2. Initialize Logger using the command's error writer (supports test redirection)
			log := logger.New(cmd.ErrOrStderr(), cfg.Verbose, cfg.Format)

			// 3. Inject into the App container
			appContainer.Config = cfg
			appContainer.Logger = log
			appContainer.Out = cmd.OutOrStdout()
			appContainer.ErrOut = cmd.ErrOrStderr()
			appContainer.In = cmd.InOrStdin()

			return nil
		},
	}

	var yesGlobal bool

	// Define global persistent flags available to all subcommands
	rootCmd.PersistentFlags().StringVarP(&cfgFile, "config", "c", "", "config file (default is $HOME/.nept.yaml)")
	rootCmd.PersistentFlags().BoolVarP(&verbose, "verbose", "v", false, "enable verbose debug logging")
	rootCmd.PersistentFlags().StringVarP(&format, "format", "f", "text", "output format (text, json)")
	rootCmd.PersistentFlags().BoolVarP(&yesGlobal, "yes", "y", false, "Skip confirmation prompts")

	// Register subcommands
	rootCmd.AddCommand(NewVersionCmd(appContainer))
	rootCmd.AddCommand(NewHelloCmd(appContainer))
	rootCmd.AddCommand(NewSchemaCmd(appContainer, rootCmd))
	rootCmd.AddCommand(Auther(appContainer))
	rootCmd.AddCommand(Login(appContainer))
	rootCmd.AddCommand(NewStatusCmd(appContainer))
	rootCmd.AddCommand(NewConfigCmd(appContainer))
	rootCmd.AddCommand(NewDeployCmd(appContainer))
	rootCmd.AddCommand(NewLogsCmd(appContainer))
	rootCmd.AddCommand(NewAppCmd(appContainer))
	rootCmd.AddCommand(NewDbCmd(appContainer))
	rootCmd.AddCommand(NewRestartCmd(appContainer))
	rootCmd.AddCommand(NewDeleteCmd(appContainer))
	rootCmd.AddCommand(NewDomainCmd(appContainer))

	return rootCmd
}

// Execute builds and executes the root command. It returns the exit status code.
func Execute() int {
	appContainer := &app.App{
		Out:    os.Stdout,
		ErrOut: os.Stderr,
		In:     os.Stdin,
	}
	rootCmd := NewRootCmd(appContainer)

	// Suppress Cobra's built-in error/usage outputs since we format and print
	// errors structured as JSON/text inside the App container.
	rootCmd.SilenceErrors = true
	rootCmd.SilenceUsage = true

	if err := rootCmd.Execute(); err != nil {
		var appErr *app.AppError
		if errors.As(err, &appErr) {
			appContainer.PrintErr(appErr.Code, appErr.Err)
		} else {
			appContainer.PrintErr("EXECUTION_ERROR", err)
		}
		return 1
	}
	return 0
}
