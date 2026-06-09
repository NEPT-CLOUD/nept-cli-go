package cmd

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/NEPT-CLOUD/nept-cli-go/internal/app"
	"github.com/NEPT-CLOUD/nept-cli-go/internal/app/utils"
	"github.com/spf13/cobra"
)

type DeployPayload struct {
	ProjectName       string            `json:"projectName"`
	Branch            string            `json:"branch"`
	Commit            string            `json:"commit"`
	CommitMessage     string            `json:"commitMessage"`
	RootDirectory     string            `json:"rootDirectory"`
	BuildCommand      []string          `json:"buildCommand"`
	OutputDirectory   string            `json:"outputDirectory"`
	SelectedFramework string            `json:"selectedFramework"`
	EnvVars           map[string]string `json:"envVars"`
	UserID            string            `json:"userID"`
	RunCommand        string            `json:"runCommand"`
	FullRepoName      string            `json:"fullRepoName"`
	Port              int               `json:"port"`
	Version           string            `json:"version"`
	ZipBuffer         string            `json:"zipBuffer"`
}

type DeployResponse struct {
	Success      bool   `json:"success"`
	Message      string `json:"message"`
	LogsID       string `json:"logsID"`
	Timestamp    string `json:"timestamp"`
	ProjectId    string `json:"projectId"`
	DeploymentId string `json:"deploymentId"`
	Domain       string `json:"domain"`
}

func sanitizeName(name string) string {
	res := strings.ToLower(name)
	// Replace non-alphanumeric with -
	var sb strings.Builder
	for _, r := range res {
		if (r >= 'a' && r <= 'z') || (r >= '0' && r <= '9') || r == '-' {
			sb.WriteRune(r)
		} else {
			sb.WriteRune('-')
		}
	}
	res = sb.String()
	// Clean up duplicate dashes
	for strings.Contains(res, "--") {
		res = strings.ReplaceAll(res, "--", "-")
	}
	res = strings.Trim(res, "-")
	if len(res) > 63 {
		res = res[:63]
	}
	if res == "" {
		res = "app"
	}
	return res
}

func NewDeployCmd(appContainer *app.App) *cobra.Command {
	var (
		nameFlag      string
		fwFlag        string
		portFlag      int
		verFlag       string
		rootFlag      string
		noFollowFlag  bool
		envSlice      []string
	)

	cmd := &cobra.Command{
		Use:   "deploy [dir]",
		Short: "Package & deploy a directory",
		Args:  cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			yesFlag, _ := cmd.Flags().GetBool("yes")
			dir := "."
			if len(args) > 0 {
				dir = args[0]
			}
			absDir, err := filepath.Abs(dir)
			if err != nil {
				return err
			}

			dirStat, err := os.Stat(absDir)
			if err != nil || !dirStat.IsDir() {
				return fmt.Errorf("not a directory: %s", absDir)
			}

			// 1. Resolve UserID
			userID, err := appContainer.ResolveUserID()
			if err != nil {
				return err
			}

			// 2. Resolve framework & git context
			preset := utils.DetectFramework(absDir)
			if fwFlag != "" {
				preset = utils.PresetFor(fwFlag)
			} else if appContainer.Config != nil && appContainer.Config.Format != "json" && !yesFlag {
				// We can implement interactive confirm here if needed
				confirmMsg := fmt.Sprintf("Detected framework: %s%s%s. Use it?", utils.ColorBold, preset.Framework, utils.ColorReset)
				if !utils.Confirm(appContainer.In, appContainer.Out, true, confirmMsg, true) {
					selected := utils.AskChoice(appContainer.In, appContainer.Out, true, "Select framework", utils.FrameworkNames, preset.Framework)
					preset = utils.PresetFor(selected)
				}
			}

			git := utils.ResolveGitInfo(absDir)

			projectName := sanitizeName(filepath.Base(absDir))
			if nameFlag != "" {
				projectName = sanitizeName(nameFlag)
			} else if appContainer.Config != nil && appContainer.Config.Format != "json" && !yesFlag {
				projectName = sanitizeName(utils.Ask(appContainer.In, appContainer.Out, true, "Project name", projectName))
			}

			port := preset.Port
			if portFlag != 0 {
				port = portFlag
			}

			version := preset.Version
			if verFlag != "" {
				version = verFlag
			}

			rootDirectory := "."
			if rootFlag != "" {
				rootDirectory = rootFlag
			}

			// Parse envVars
			envVars := make(map[string]string)
			for _, pair := range envSlice {
				eq := strings.Index(pair, "=")
				if eq > 0 {
					envVars[pair[:eq]] = pair[eq+1:]
				}
			}

			// 3. Packaging
			if appContainer.Config != nil && appContainer.Config.Format != "json" {
				fmt.Fprintf(appContainer.Out, "Packaging project...\n")
			}
			zipRes, err := utils.ZipDirectory(absDir)
			if err != nil {
				return err
			}
			if appContainer.Config != nil && appContainer.Config.Format != "json" {
				fmt.Fprintf(appContainer.Out, "%s%s%s Packaged %d files (%s)\n", utils.ColorGreen, utils.SymbolOk, utils.ColorReset, zipRes.FileCount, utils.HumanBytes(zipRes.Bytes))
			}

			// 4. Upload & build
			payload := DeployPayload{
				ProjectName:       projectName,
				Branch:            git.Branch,
				Commit:            git.Commit,
				CommitMessage:     git.CommitMessage,
				RootDirectory:     rootDirectory,
				BuildCommand:      preset.BuildCommand,
				OutputDirectory:   preset.OutputDirectory,
				SelectedFramework: preset.Framework,
				EnvVars:           envVars,
				UserID:            userID,
				RunCommand:        preset.RunCommand,
				FullRepoName:      git.FullRepoName,
				Port:              port,
				Version:           version,
				ZipBuffer:         zipRes.Base64,
			}

			if payload.Branch == "" {
				payload.Branch = "main"
			}
			if payload.CommitMessage == "" {
				payload.CommitMessage = "Deployed via Nept CLI"
			}

			if appContainer.Config != nil && appContainer.Config.Format != "json" {
				fmt.Fprintf(appContainer.Out, "Uploading & starting build...\n")
			}

			var resp DeployResponse
			_, err = utils.CallAPI(appContainer, "POST", "/api/deploy", payload, &resp)
			if err != nil {
				return err
			}

			if appContainer.Config != nil && appContainer.Config.Format != "json" {
				fmt.Fprintf(appContainer.Out, "%s%s%s Build started\n", utils.ColorGreen, utils.SymbolOk, utils.ColorReset)
				fmt.Fprintln(appContainer.Out)
				fmt.Fprintf(appContainer.Out, "  %s%-12s%s %s\n", utils.ColorDim, "project", utils.ColorReset, projectName)
				if resp.DeploymentId != "" {
					fmt.Fprintf(appContainer.Out, "  %s%-12s%s %s\n", utils.ColorDim, "deployment", utils.ColorReset, resp.DeploymentId)
				}
				if resp.LogsID != "" {
					fmt.Fprintf(appContainer.Out, "  %s%-12s%s %s\n", utils.ColorDim, "logs", utils.ColorReset, resp.LogsID)
				}
				fmt.Fprintln(appContainer.Out)
			}

			if appContainer.Config != nil && appContainer.Config.Format == "json" {
				return appContainer.PrintResult("", resp)
			}

			if resp.LogsID == "" || noFollowFlag {
				if resp.Domain != "" {
					fmt.Fprintf(appContainer.Out, "%s%s%s https://%s\n", utils.ColorGreen, utils.SymbolOk, utils.ColorReset, resp.Domain)
				}
				return nil
			}

			fmt.Fprintf(appContainer.Out, "%sBuild logs:%s\n", utils.ColorBold, utils.ColorReset)
			failed, _, err := utils.StreamBuildLogs(appContainer, resp.LogsID, false)
			if err != nil {
				return err
			}
			fmt.Fprintln(appContainer.Out)

			if failed {
				return fmt.Errorf("Build failed — see logs above")
			}

			fmt.Fprintf(appContainer.Out, "%s%s%s Deployed\n", utils.ColorGreen, utils.SymbolOk, utils.ColorReset)
			if resp.Domain != "" {
				fmt.Fprintf(appContainer.Out, "  %s%s%s https://%s\n", utils.ColorCyan, utils.SymbolArrow, utils.ColorReset, resp.Domain)
			}

			return nil
		},
	}

	cmd.Flags().StringVar(&nameFlag, "name", "", "Project name")
	cmd.Flags().StringVar(&fwFlag, "framework", "", "Override auto-detected framework")
	cmd.Flags().IntVar(&portFlag, "port", 0, "Container port")
	cmd.Flags().StringVar(&verFlag, "version", "", "Runtime version")
	cmd.Flags().StringVar(&rootFlag, "root", "", "Root directory")
	cmd.Flags().BoolVar(&noFollowFlag, "no-follow", false, "Don't stream logs after starting build")
	cmd.Flags().StringSliceVarP(&envSlice, "env", "e", nil, "Environment variable, repeatable (K=V)")

	return cmd
}
