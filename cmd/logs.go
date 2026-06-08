package cmd

import (
	"fmt"
	"os"
	"strings"

	"github.com/NEPT-CLOUD/nept-cli-go/internal/app"
	"github.com/NEPT-CLOUD/nept-cli-go/internal/app/utls"
	"github.com/spf13/cobra"
)

func NewLogsCmd(appContainer *app.App) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "logs <logsId>",
		Short: "Stream build logs",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			logsID := args[0]

			jsonMode := false
			if appContainer.Config != nil && appContainer.Config.Format == "json" {
				jsonMode = true
			}

			if !jsonMode {
				fmt.Fprintf(appContainer.Out, "%sStreaming logs for %s%s\n", utls.ColorBold, logsID, utls.ColorReset)
			}

			failed, entries, err := utls.StreamBuildLogs(appContainer, logsID, jsonMode)
			if err != nil {
				return err
			}

			if jsonMode {
				type logsResponse struct {
					LogsID    string          `json:"logsId"`
					Completed bool            `json:"completed"`
					Failed    bool            `json:"failed"`
					Entries   []utls.LogEntry `json:"entries"`
				}
				resp := logsResponse{
					LogsID:    logsID,
					Completed: true,
					Failed:    failed,
					Entries:   entries,
				}
				_ = appContainer.PrintResult("", resp)
				if failed {
					os.Exit(1)
				}
				return nil
			}

			fmt.Fprintln(appContainer.Out)
			if failed {
				fmt.Fprintf(appContainer.Out, "%s%s%s Build failed\n", utls.ColorRed, utls.SymbolErr, utls.ColorReset)
				os.Exit(1)
			} else {
				fmt.Fprintf(appContainer.Out, "%s%s%s Stream ended\n", utls.ColorGreen, utls.SymbolOk, utls.ColorReset)
			}

			return nil
		},
	}
	return cmd
}

type AppLogEntry struct {
	Level     string `json:"level"`
	Message   string `json:"message"`
	Timestamp string `json:"timestamp"`
}

func NewAppCmd(appContainer *app.App) *cobra.Command {
	appCmd := &cobra.Command{
		Use:   "app",
		Short: "Manage deployed applications",
	}

	appCmd.AddCommand(NewAppLogsCmd(appContainer))

	return appCmd
}

func NewAppLogsCmd(appContainer *app.App) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "logs <deploymentId>",
		Short: "Fetch runtime/deployment logs",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			deploymentID := args[0]

			var logs []AppLogEntry
			_, err := utls.CallAPI(appContainer, "GET", "/api/deploymentLogs/"+deploymentID, nil, &logs)
			if err != nil {
				return err
			}

			type appLogsResponse struct {
				DeploymentID string        `json:"deploymentId"`
				Count        int           `json:"count"`
				Logs         []AppLogEntry `json:"logs"`
			}
			resp := appLogsResponse{
				DeploymentID: deploymentID,
				Count:        len(logs),
				Logs:         logs,
			}

			var textVal strings.Builder
			if len(logs) == 0 {
				textVal.WriteString(utls.ColorDim + "No logs found." + utls.ColorReset)
			} else {
				for i, entry := range logs {
					if i > 0 {
						textVal.WriteString("\n")
					}
					timeStr := ""
					if entry.Timestamp != "" {
						timeStr = utls.ColorDim + "[" + entry.Timestamp + "] " + utls.ColorReset
					}
					color := utls.LevelColor(entry.Level)
					textVal.WriteString(fmt.Sprintf("%s%s%s", timeStr, color, entry.Message))
				}
			}

			return appContainer.PrintResult(textVal.String(), resp)
		},
	}
	return cmd
}
