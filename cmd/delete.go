package cmd

import (
	"fmt"

	"github.com/NEPT-CLOUD/nept-cli-go/internal/app"
	"github.com/NEPT-CLOUD/nept-cli-go/internal/app/utls"
	"github.com/spf13/cobra"
)

type DeletePayload struct {
	ProjectName string `json:"projectName"`
	UserID      string `json:"userID"`
}

type DeleteResponse struct {
	Success bool   `json:"success"`
	Message string `json:"message"`
}

func NewDeleteCmd(appContainer *app.App) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "delete <projectName>",
		Short: "Delete a deployment",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			projectName := args[0]
			yesFlag, _ := cmd.Flags().GetBool("yes")

			userId, err := appContainer.ResolveUserID()
			if err != nil {
				return err
			}

			// Interactive confirm if TTY, unless yesFlag is set
			if appContainer.Config != nil && appContainer.Config.Format != "json" && !yesFlag {
				confirmMsg := fmt.Sprintf("Delete %s%s%s and all its resources?", utls.ColorBold, projectName, utls.ColorReset)
				if !utls.Confirm(appContainer.In, appContainer.Out, true, confirmMsg, false) {
					fmt.Fprintln(appContainer.Out, utls.ColorDim+"Aborted."+utls.ColorReset)
					return nil
				}
			}

			if appContainer.Config != nil && appContainer.Config.Format != "json" {
				fmt.Fprintf(appContainer.Out, "Deleting %s...\n", projectName)
			}

			payload := DeletePayload{
				ProjectName: projectName,
				UserID:      userId,
			}

			var resp DeleteResponse
			_, err = utls.CallAPI(appContainer, "DELETE", "/api/deploy", payload, &resp)
			if err != nil {
				return err
			}

			msg := resp.Message
			if msg == "" {
				msg = fmt.Sprintf("Deleted %s", projectName)
			}

			textVal := fmt.Sprintf("%s%s%s %s", utls.ColorGreen, utls.SymbolOk, utls.ColorReset, msg)
			return appContainer.PrintResult(textVal, resp)
		},
	}

	return cmd
}
