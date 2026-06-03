package cmd

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/NEPT-CLOUD/nept-cli-go/internal/app"
)

// These variables are designed to be overridden during build time using linker flags:
// e.g. go build -ldflags "-X github.com/NEPT-CLOUD/nept-cli-go/cmd.Version=1.0.0"
var (
	Version   = "dev"
	GitCommit = "unknown"
	BuildDate = "unknown"
)

// VersionInfo models the structure of build metadata.
type VersionInfo struct {
	Version   string `json:"version"`
	GitCommit string `json:"git_commit"`
	BuildDate string `json:"build_date"`
}

// NewVersionCmd constructs the version subcommand.
func NewVersionCmd(appContainer *app.App) *cobra.Command {
	return &cobra.Command{
		Use:   "version",
		Short: "Print CLI build and version information",
		Long:  `Displays details about the application's compiled version, git commit hash, and build timestamp.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			info := VersionInfo{
				Version:   Version,
				GitCommit: GitCommit,
				BuildDate: BuildDate,
			}

			textOut := fmt.Sprintf("nept version:    %s\ngit commit:      %s\nbuild date:      %s", info.Version, info.GitCommit, info.BuildDate)
			return appContainer.PrintResult(textOut, info)
		},
	}
}
