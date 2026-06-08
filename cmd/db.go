package cmd

import (
	"fmt"
	"strings"

	"github.com/NEPT-CLOUD/nept-cli-go/internal/app"
	"github.com/NEPT-CLOUD/nept-cli-go/internal/app/utls"
	"github.com/spf13/cobra"
)

type DbDeployPayload struct {
	DbType     string  `json:"dbType"`
	Version    string  `json:"version"`
	AppName    string  `json:"appName"`
	Username   string  `json:"username"`
	UserId     string  `json:"userId"`
	VolumeSize float64 `json:"volumeSize"`
	Cpu        float64 `json:"cpu"`
	Memory     float64 `json:"memory"`
}

type DbDeployResponse struct {
	DatabaseId    string `json:"databaseId"`
	Host          string `json:"host"`
	Port          int    `json:"port"`
	Username      string `json:"username"`
	Password      string `json:"password"`
	ConnectionUrl string `json:"connectionUrl"`
}

func NewDbCmd(appContainer *app.App) *cobra.Command {
	dbCmd := &cobra.Command{
		Use:   "db",
		Short: "Manage databases",
	}

	dbCmd.AddCommand(NewDbDeployCmd(appContainer))

	return dbCmd
}

func NewDbDeployCmd(appContainer *app.App) *cobra.Command {
	var (
		typeFlag string
		nameFlag string
		verFlag  string
		userFlag string
		volFlag  float64
		cpuFlag  float64
		memFlag  float64
	)

	cmd := &cobra.Command{
		Use:   "deploy",
		Short: "Deploy a database",
		RunE: func(cmd *cobra.Command, args []string) error {
			userId, err := appContainer.ResolveUserID()
			if err != nil {
				return err
			}

			dbTypes := []string{"postgres", "mysql", "mongodb", "redis"}
			defaultVersions := map[string]string{
				"postgres": "16",
				"mysql":    "8",
				"mongodb":  "7",
				"redis":    "7-alpine",
			}

			// Interactive Prompts if in Human/Interactive mode and flags are not specified
			dbType := typeFlag
			if dbType == "" && appContainer.Config != nil && appContainer.Config.Format != "json" {
				dbType = utls.AskChoice(appContainer.In, appContainer.Out, true, "Database type", dbTypes, "postgres")
			}
			if dbType == "" {
				dbType = "postgres"
			}

			validType := false
			for _, t := range dbTypes {
				if t == dbType {
					validType = true
					break
				}
			}
			if !validType {
				return fmt.Errorf("invalid database type: %s. Allowed: %s", dbType, strings.Join(dbTypes, ", "))
			}

			appName := nameFlag
			if appName == "" && appContainer.Config != nil && appContainer.Config.Format != "json" {
				appName = utls.Ask(appContainer.In, appContainer.Out, true, "Database app name", "my-"+dbType)
			}
			appName = sanitizeName(appName)

			version := verFlag
			if version == "" {
				version = defaultVersions[dbType]
			}

			username := userFlag
			if username == "" && appContainer.Config != nil && appContainer.Config.Format != "json" {
				username = utls.Ask(appContainer.In, appContainer.Out, true, "Database username", "admin")
			}
			if username == "" {
				username = "admin"
			}

			if appContainer.Config != nil && appContainer.Config.Format != "json" {
				fmt.Fprintf(appContainer.Out, "Deploying %s database...\n", dbType)
			}

			payload := DbDeployPayload{
				DbType:     dbType,
				Version:    version,
				AppName:    appName,
				Username:   username,
				UserId:     userId,
				VolumeSize: volFlag,
				Cpu:        cpuFlag,
				Memory:     memFlag,
			}

			var resp DbDeployResponse
			_, err = utls.CallAPI(appContainer, "POST", "/api/deploy-db", payload, &resp)
			if err != nil {
				return err
			}

			var textVal strings.Builder
			textVal.WriteString(fmt.Sprintf("%s%s%s Database '%s' deployed\n\n", utls.ColorGreen, utls.SymbolOk, utls.ColorReset, appName))
			textVal.WriteString(fmt.Sprintf("  %s%-12s%s %s\n", utls.ColorDim, "id", utls.ColorReset, resp.DatabaseId))
			textVal.WriteString(fmt.Sprintf("  %s%-12s%s %s\n", utls.ColorDim, "host", utls.ColorReset, resp.Host))
			textVal.WriteString(fmt.Sprintf("  %s%-12s%s %d\n", utls.ColorDim, "port", utls.ColorReset, resp.Port))
			textVal.WriteString(fmt.Sprintf("  %s%-12s%s %s\n", utls.ColorDim, "username", utls.ColorReset, resp.Username))
			textVal.WriteString(fmt.Sprintf("  %s%-12s%s %s%s%s\n", utls.ColorDim, "password", utls.ColorReset, utls.ColorYellow, resp.Password, utls.ColorReset))
			textVal.WriteString(fmt.Sprintf("  %s%-12s%s %s%s%s\n\n", utls.ColorDim, "url", utls.ColorReset, utls.ColorCyan, resp.ConnectionUrl, utls.ColorReset))
			textVal.WriteString(fmt.Sprintf("%s%s Store the password now — it cannot be retrieved later.%s", utls.ColorYellow, utls.SymbolWarn, utls.ColorReset))

			return appContainer.PrintResult(textVal.String(), resp)
		},
	}

	cmd.Flags().StringVar(&typeFlag, "type", "", "Database type (postgres | mysql | mongodb | redis)")
	cmd.Flags().StringVar(&nameFlag, "name", "", "Database app name")
	cmd.Flags().StringVar(&verFlag, "version", "", "DB version")
	cmd.Flags().StringVar(&userFlag, "username", "", "DB username")
	cmd.Flags().Float64Var(&volFlag, "volume", 10.0, "Storage size in Gi")
	cmd.Flags().Float64Var(&cpuFlag, "cpu", 1.0, "CPU cores")
	cmd.Flags().Float64Var(&memFlag, "memory", 512.0, "Memory in Mi")

	return cmd
}
