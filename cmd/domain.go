package cmd

import (
	"fmt"
	"strings"

	"github.com/NEPT-CLOUD/nept-cli-go/internal/app"
	"github.com/NEPT-CLOUD/nept-cli-go/internal/app/utils"
	"github.com/spf13/cobra"
)

type DNSInfo struct {
	Type  string `json:"type"`
	Name  string `json:"name"`
	Value string `json:"value"`
}

type SSLInfo struct {
	Status string `json:"status"`
	Method string `json:"method"`
}

type DomainResponse struct {
	Success               bool     `json:"success"`
	Message               string   `json:"message"`
	OwnershipVerification *DNSInfo `json:"ownership_verification"`
	OwnershipTxt          *DNSInfo `json:"ownership_txt"`
	SSLValidation         *DNSInfo `json:"ssl_validation"`
	SSL                   *SSLInfo `json:"ssl"`
}

type DomainPayload struct {
	ProjectId string `json:"projectId"`
	Domain    string `json:"domain"`
}

func NewDomainCmd(appContainer *app.App) *cobra.Command {
	domainCmd := &cobra.Command{
		Use:   "domain",
		Short: "Manage custom domains",
	}

	domainCmd.AddCommand(NewDomainAddCmd(appContainer))

	return domainCmd
}

func NewDomainAddCmd(appContainer *app.App) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "add <projectId> <domain>",
		Short: "Attach a custom domain",
		Args:  cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			projectId := args[0]
			domain := args[1]

			if appContainer.Config != nil && appContainer.Config.Format != "json" {
				fmt.Fprintf(appContainer.Out, "Attaching %s to project %s...\n", domain, projectId)
			}

			payload := DomainPayload{
				ProjectId: projectId,
				Domain:    domain,
			}

			var resp DomainResponse
			_, err := utils.CallAPI(appContainer, "POST", "/api/domains/domains", payload, &resp)
			if err != nil {
				return err
			}

			var textVal strings.Builder
			textVal.WriteString(fmt.Sprintf("%s%s%s Domain %s registered\n\n", utils.ColorGreen, utils.SymbolOk, utils.ColorReset, domain))
			textVal.WriteString(fmt.Sprintf("%sAdd these DNS records:%s\n", utils.ColorBold, utils.ColorReset))

			if resp.OwnershipVerification != nil {
				o := resp.OwnershipVerification
				textVal.WriteString(fmt.Sprintf("  %s%-6s%s %s %s %s%s%s\n", utils.ColorDim, o.Type, utils.ColorReset, o.Name, utils.SymbolArrow, utils.ColorCyan, o.Value, utils.ColorReset))
			}
			if resp.OwnershipTxt != nil {
				o := resp.OwnershipTxt
				textVal.WriteString(fmt.Sprintf("  %s%-6s%s %s %s %s%s%s\n", utils.ColorDim, o.Type, utils.ColorReset, o.Name, utils.SymbolArrow, utils.ColorCyan, o.Value, utils.ColorReset))
			}
			if resp.SSLValidation != nil {
				o := resp.SSLValidation
				textVal.WriteString(fmt.Sprintf("  %s%-6s%s %s %s %s%s%s\n", utils.ColorDim, o.Type, utils.ColorReset, o.Name, utils.SymbolArrow, utils.ColorCyan, o.Value, utils.ColorReset))
			}

			textVal.WriteString("\n")
			sslStatus := "pending"
			sslMethod := "txt"
			if resp.SSL != nil {
				if resp.SSL.Status != "" {
					sslStatus = resp.SSL.Status
				}
				if resp.SSL.Method != "" {
					sslMethod = resp.SSL.Method
				}
			}
			textVal.WriteString(fmt.Sprintf("%sSSL: %s (%s)%s", utils.ColorDim, sslStatus, sslMethod, utils.ColorReset))

			return appContainer.PrintResult(textVal.String(), resp)
		},
	}
	return cmd
}
