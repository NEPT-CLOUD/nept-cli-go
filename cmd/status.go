package cmd

import (
	"fmt"
	"strings"

	"github.com/NEPT-CLOUD/nept-cli-go/internal/app"
	"github.com/NEPT-CLOUD/nept-cli-go/internal/app/utils"
	"github.com/spf13/cobra"
)

type HealthResponse struct {
	Connected bool    `json:"connected"`
	Status    string  `json:"status"`
	Uptime    float64 `json:"uptime"`
	Timestamp string  `json:"timestamp"`
}

func NewStatusCmd(appContainer *app.App) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "status",
		Short: "Check engine health",
		RunE: func(cmd *cobra.Command, args []string) error {
			var resp HealthResponse
			apiURL := "https://server.nept.cloud"
			if appContainer.Config != nil && appContainer.Config.APIURL != "" {
				apiURL = appContainer.Config.APIURL
			} else if utils.BackendUrl != "" {
				apiURL = utils.BackendUrl
			}

			_, err := utils.CallAPI(appContainer, "GET", "/api/health", nil, &resp)
			if err != nil {
				return err
			}
			resp.Connected = true

			var textVal strings.Builder
			textVal.WriteString(fmt.Sprintf("%s%s%s %sEngine online%s  %s\n", utils.ColorGreen, utils.SymbolOk, utils.ColorReset, utils.ColorBold, utils.ColorReset, apiURL))
			if resp.Status != "" {
				textVal.WriteString(fmt.Sprintf("  status    %s\n", resp.Status))
			}
			if resp.Uptime > 0 {
				textVal.WriteString(fmt.Sprintf("  uptime    %.0fs\n", resp.Uptime))
			}
			if resp.Timestamp != "" {
				textVal.WriteString(fmt.Sprintf("  time      %s\n", resp.Timestamp))
			}

			return appContainer.PrintResult(textVal.String(), resp)
		},
	}
	return cmd
}
