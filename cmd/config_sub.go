package cmd

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/NEPT-CLOUD/nept-cli-go/internal/app"
	"github.com/NEPT-CLOUD/nept-cli-go/internal/app/utils"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

func NewConfigCmd(appContainer *app.App) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "config",
		Short: "Manage CLI configuration",
	}

	cmd.AddCommand(NewConfigSetCmd(appContainer))
	cmd.AddCommand(NewConfigGetCmd(appContainer))
	cmd.AddCommand(NewConfigListCmd(appContainer))

	return cmd
}

func setConfigKey(key string, val interface{}) error {
	v := viper.New()
	v.SetConfigName(".nept")
	v.SetConfigType("yaml")
	v.AddConfigPath(".")
	home, err := os.UserHomeDir()
	if err == nil {
		v.AddConfigPath(home)
	}

	_ = v.ReadInConfig()
	v.Set(key, val)

	dest := v.ConfigFileUsed()
	if dest == "" {
		if home == "" {
			return fmt.Errorf("unable to determine home directory for configuration")
		}
		dest = filepath.Join(home, ".nept.yaml")
		return v.WriteConfigAs(dest)
	}

	return v.WriteConfig()
}

func getSource(key string, fileViper *viper.Viper) string {
	envKey := "NEPT_" + strings.ToUpper(strings.ReplaceAll(key, ".", "_"))
	if os.Getenv(envKey) != "" {
		return "env"
	}
	if fileViper != nil && fileViper.IsSet(key) {
		return "file"
	}
	return "default"
}

func loadFileViper() *viper.Viper {
	v := viper.New()
	v.SetConfigName(".nept")
	v.SetConfigType("yaml")
	v.AddConfigPath(".")
	home, err := os.UserHomeDir()
	if err == nil {
		v.AddConfigPath(home)
	}
	if err := v.ReadInConfig(); err == nil {
		return v
	}
	return nil
}

func NewConfigSetCmd(appContainer *app.App) *cobra.Command {
	return &cobra.Command{
		Use:   "set <key> <value>",
		Short: "Set a configuration value",
		Args:  cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			key := args[0]
			val := args[1]

			if key != "api_url" && key != "user_id" && key != "api_key" {
				return fmt.Errorf("invalid config key. Allowed: api_url, user_id, api_key")
			}

			err := setConfigKey(key, val)
			if err != nil {
				return err
			}

			type setResponse struct {
				Success bool   `json:"success"`
				Key     string `json:"key"`
				Value   string `json:"value"`
			}
			resp := setResponse{Success: true, Key: key, Value: val}

			textVal := fmt.Sprintf("%s%s%s Set %s%s%s = %s", utils.ColorGreen, utils.SymbolOk, utils.ColorReset, utils.ColorBold, key, utils.ColorReset, val)
			return appContainer.PrintResult(textVal, resp)
		},
	}
}

func NewConfigGetCmd(appContainer *app.App) *cobra.Command {
	return &cobra.Command{
		Use:   "get <key>",
		Short: "Get a configuration value",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			key := args[0]

			if key != "api_url" && key != "user_id" && key != "api_key" {
				return fmt.Errorf("invalid config key. Allowed: api_url, user_id, api_key")
			}

			val := ""
			if appContainer.Config != nil {
				switch key {
				case "api_url":
					val = appContainer.Config.APIURL
				case "user_id":
					val = appContainer.Config.UserID
					if val == "" {
						val, _ = appContainer.ResolveUserID()
					}
				case "api_key":
					val = appContainer.Config.APIKey
					if val == "" {
						val, _ = appContainer.ResolveAPIKey()
					}
				}
			}

			type getResponse struct {
				Key   string `json:"key"`
				Value string `json:"value"`
			}
			resp := getResponse{Key: key, Value: val}

			textVal := val
			if val == "" {
				textVal = utils.ColorDim + "(not set)" + utils.ColorReset
			}

			return appContainer.PrintResult(textVal, resp)
		},
	}
}

func NewConfigListCmd(appContainer *app.App) *cobra.Command {
	return &cobra.Command{
		Use:   "list",
		Short: "List active configuration settings",
		RunE: func(cmd *cobra.Command, args []string) error {
			fileViper := loadFileViper()

			apiURL := "https://server.nept.cloud"
			userID := ""
			apiKey := ""
			if appContainer.Config != nil {
				apiURL = appContainer.Config.APIURL
				userID = appContainer.Config.UserID
				apiKey = appContainer.Config.APIKey
			}

			urlSrc := getSource("api_url", fileViper)
			uidSrc := getSource("user_id", fileViper)
			keySrc := getSource("api_key", fileViper)

			if apiKey == "" {
				if key, err := appContainer.ResolveAPIKey(); err == nil {
					apiKey = key
					keySrc = "keyring"
				}
			}
			if userID == "" {
				if uid, err := appContainer.ResolveUserID(); err == nil {
					userID = uid
					uidSrc = "keyring"
				}
			}

			type listResponse struct {
				APIURL     string            `json:"api_url"`
				UserID     string            `json:"user_id"`
				APIKey     string            `json:"api_key"`
				Sources    map[string]string `json:"sources"`
				ConfigFile string            `json:"config_file"`
			}

			configFile := ""
			if fileViper != nil {
				configFile = fileViper.ConfigFileUsed()
			} else {
				home, _ := os.UserHomeDir()
				configFile = filepath.Join(home, ".nept.yaml")
			}

			resp := listResponse{
				APIURL:  apiURL,
				UserID:  userID,
				APIKey:  apiKey,
				ConfigFile: configFile,
				Sources: map[string]string{
					"api_url": urlSrc,
					"user_id": uidSrc,
					"api_key": keySrc,
				},
			}

			var textVal strings.Builder
			textVal.WriteString(fmt.Sprintf("%sEffective configuration%s\n", utils.ColorBold, utils.ColorReset))
			textVal.WriteString(fmt.Sprintf("  api_url   %s %s(%s)%s\n", apiURL, utils.ColorDim, urlSrc, utils.ColorReset))
			
			uidStr := userID
			if uidStr == "" {
				uidStr = utils.ColorDim + "(not set)" + utils.ColorReset
			}
			textVal.WriteString(fmt.Sprintf("  user_id   %s %s(%s)%s\n", uidStr, utils.ColorDim, uidSrc, utils.ColorReset))

			keyStr := apiKey
			if keyStr == "" {
				keyStr = utils.ColorDim + "(not set)" + utils.ColorReset
			} else {
				// Mask key for safety in text output
				if len(keyStr) > 12 {
					keyStr = keyStr[:8] + "..." + keyStr[len(keyStr)-4:]
				}
			}
			textVal.WriteString(fmt.Sprintf("  api_key   %s %s(%s)%s\n", keyStr, utils.ColorDim, keySrc, utils.ColorReset))
			textVal.WriteString(fmt.Sprintf("%s  file      %s%s", utils.ColorDim, configFile, utils.ColorReset))

			return appContainer.PrintResult(textVal.String(), resp)
		},
	}
}
