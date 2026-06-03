package config

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/spf13/viper"
)

const (
	// DefaultConfigFileName is the name of the configuration file without extension.
	DefaultConfigFileName = ".nept"
	// DefaultConfigFileType is the format of the configuration file.
	DefaultConfigFileType = "yaml"
	// EnvPrefix is the prefix for environment variable overrides.
	EnvPrefix = "NEPT"
)

// Config holds all structured configurations for the CLI application.
type Config struct {
	Environment string `mapstructure:"environment"`
	APIKey      string `mapstructure:"api_key"`
	Verbose     bool   `mapstructure:"verbose"`
	Format      string `mapstructure:"format"`
}

// Load reads config from defaults, config file (if any), and environment variables.
// If configFilePath is specified, it explicitly reads from that file. Otherwise,
// it searches current directory followed by the user's home directory.
func Load(configFilePath string) (*Config, error) {
	v := viper.New()

	// 1. Establish Default values
	v.SetDefault("environment", "production")
	v.SetDefault("api_key", "")
	v.SetDefault("verbose", false)
	v.SetDefault("format", "text")

	// 2. Set up Environment Variable parsing (e.g. NEPT_API_KEY)
	v.SetEnvPrefix(EnvPrefix)
	v.AutomaticEnv()
	v.SetEnvKeyReplacer(strings.NewReplacer(".", "_", "-", "_"))

	// 3. Set up Config File paths
	if configFilePath != "" {
		v.SetConfigFile(configFilePath)
	} else {
		v.SetConfigName(DefaultConfigFileName)
		v.SetConfigType(DefaultConfigFileType)
		v.AddConfigPath(".")
		if home, err := os.UserHomeDir(); err == nil {
			v.AddConfigPath(home)
		}
	}

	// 4. Load configuration
	err := v.ReadInConfig()
	if err != nil {
		var configFileNotFoundError viper.ConfigFileNotFoundError
		if errors.As(err, &configFileNotFoundError) {
			// If file is not found and no specific file was requested, it's okay.
			// We just fallback on defaults / environment variables.
			if configFilePath != "" {
				return nil, fmt.Errorf("specified config file not found: %w", err)
			}
		} else {
			return nil, fmt.Errorf("failed to parse config file: %w", err)
		}
	}

	// 5. Decode into struct
	var cfg Config
	if err := v.Unmarshal(&cfg); err != nil {
		return nil, fmt.Errorf("unable to decode config: %w", err)
	}

	return &cfg, nil
}

// SaveDefault writes a default configuration template to the specified path.
// If the path is empty, it writes to the default home directory configuration path.
func SaveDefault(path string) error {
	v := viper.New()
	v.Set("environment", "production")
	v.Set("api_key", "your_api_key_here")
	v.Set("verbose", false)
	v.Set("format", "text")

	var destPath string
	if path != "" {
		destPath = path
	} else {
		home, err := os.UserHomeDir()
		if err != nil {
			return fmt.Errorf("unable to determine user home directory: %w", err)
		}
		destPath = filepath.Join(home, fmt.Sprintf("%s.%s", DefaultConfigFileName, DefaultConfigFileType))
	}

	if _, err := os.Stat(destPath); err == nil {
		return fmt.Errorf("configuration file already exists at: %s", destPath)
	}

	dir := filepath.Dir(destPath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("failed to create configuration directories: %w", err)
	}

	if err := v.WriteConfigAs(destPath); err != nil {
		return fmt.Errorf("failed to write default config file: %w", err)
	}

	return nil
}
