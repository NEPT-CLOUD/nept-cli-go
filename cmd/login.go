package cmd

import (
	"errors"
	"fmt"
	"github.com/NEPT-CLOUD/nept-cli-go/internal/app"
	"github.com/NEPT-CLOUD/nept-cli-go/internal/app/utils"
	"github.com/spf13/cobra"
	"github.com/zalando/go-keyring"
)

func Login(appContainer *app.App) *cobra.Command {
	var key string

	cmd := &cobra.Command{
		Use:   "login",
		Short: "login to nept",
		Long:  "nept is a scalable command-line tool",
		RunE: func(cmd *cobra.Command, args []string) error {
			if key == "" {
				_ = cmd.Usage()
				return app.NewAppError("KEY_MISSING", errors.New("API_KEY is required"))
			}


			// Temporarily override APIKey config for validation call
			origKey := ""
			if appContainer.Config != nil {
				origKey = appContainer.Config.APIKey
				appContainer.Config.APIKey = key
			}

			type validateResp struct {
				UserID string `json:"userId"`
			}
			var resp validateResp
			_, err := utils.CallAPI(appContainer, "GET", "/api/keys/validate", nil, &resp)

			// Restore config
			if appContainer.Config != nil {
				appContainer.Config.APIKey = origKey
			}

			if err != nil {
				return app.NewAppError("INVALID_KEY", fmt.Errorf("Invalid API Key: %w", err))
			}

			err = keyring.Set("nept", "api-key", key)
			if err != nil {
				return app.NewAppError("KEYRING_ERROR", err)
			}

			if resp.UserID != "" {
				_ = keyring.Set("nept", "user-id", resp.UserID)
			}

			return appContainer.PrintResult("Login successful. API key saved to keychain.", map[string]string{"status": "success", "message": "API key saved to keychain"})
		},
	}

	cmd.Flags().StringVarP(&key, "key", "k", "", "collect the key from https://nept-cloud.io/dashboard/apikeys")

	return cmd
}
