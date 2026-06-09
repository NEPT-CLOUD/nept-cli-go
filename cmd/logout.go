package cmd

import (
	"errors"
	"fmt"

	"github.com/spf13/cobra"
	"github.com/zalando/go-keyring"

	"github.com/NEPT-CLOUD/nept-cli-go/internal/app"
)

// NewLogoutCmd constructs the logout subcommand.
func NewLogoutCmd(appContainer *app.App) *cobra.Command {
	return &cobra.Command{
		Use:   "logout",
		Short: "logout from nept and clear saved credentials",
		Long:  `Clears the saved API key and User ID from the system keychain.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			errKey := keyring.Delete("nept", "api-key")
			errUser := keyring.Delete("nept", "user-id")

			// If both keys were already missing, inform the user they were not logged in.
			if errors.Is(errKey, keyring.ErrNotFound) && errors.Is(errUser, keyring.ErrNotFound) {
				return appContainer.PrintResult("You are not currently logged in.", map[string]string{
					"status":  "info",
					"message": "No active session found.",
				})
			}

			// Handle other unexpected keyring errors
			if errKey != nil && !errors.Is(errKey, keyring.ErrNotFound) {
				return app.NewAppError("LOGOUT_ERROR", fmt.Errorf("failed to delete API key: %w", errKey))
			}
			if errUser != nil && !errors.Is(errUser, keyring.ErrNotFound) {
				return app.NewAppError("LOGOUT_ERROR", fmt.Errorf("failed to delete User ID: %w", errUser))
			}

			return appContainer.PrintResult("Logout successful. Credentials cleared from keychain.", map[string]string{
				"status":  "success",
				"message": "Credentials cleared from keychain.",
			})
		},
	}
}
