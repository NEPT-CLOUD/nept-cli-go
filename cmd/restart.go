package cmd

import (
	"fmt"

	"github.com/NEPT-CLOUD/nept-cli-go/internal/app"
	"github.com/NEPT-CLOUD/nept-cli-go/internal/app/utls"
	"github.com/spf13/cobra"
)

type RestartPayload struct {
	ProjectName string `json:"projectName"`
	UserID      string `json:"userID"`
}

type RestartResponse struct {
	Success bool   `json:"success"`
	Message string `json:"message"`
}

func NewRestartCmd(appContainer *app.App) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "restart <projectName>",
		Short: "Gracefully restart a deployment",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			projectName := args[0]

			userId, err := appContainer.ResolveUserID()
			if err != nil {
				return err
			}

			if appContainer.Config != nil && appContainer.Config.Format != "json" {
				fmt.Fprintf(appContainer.Out, "Restarting %s...\n", projectName)
			}

			payload := RestartPayload{
				ProjectName: projectName,
				UserID:      userId,
			}

			var resp RestartResponse
			_, err = utls.CallAPI(appContainer, "POST", "/api/deploy/restart", payload, &resp)
			if err != nil {
				return err
			}

			msg := resp.Message
			if msg == "" {
				msg = fmt.Sprintf("Restarted %s", projectName)
			}

			textVal := fmt.Sprintf("%s%s%s %s", utls.ColorGreen, utls.SymbolOk, utls.ColorReset, msg)
			return appContainer.PrintResult(textVal, resp)
		},
	}
	return cmd
}
