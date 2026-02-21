package config

import (
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/adrg/xdg"
	"github.com/spf13/viper"
)

type Config struct {
	GracefulTimeout time.Duration     `mapstructure:"graceful_timeout"`
	Protected       []string          `mapstructure:"protected"`
	Aliases         map[string]string `mapstructure:"aliases"`
	DefaultForce    bool              `mapstructure:"default_force"`
	DefaultVerbose  bool              `mapstructure:"default_verbose"`
}

func configDir() string {
	return filepath.Join(xdg.ConfigHome, "hdf")
}

func configPath() string {
	return filepath.Join(configDir(), "config.toml")
}

func Load() (*Config, error) {
	cfg := defaultConfig

	v := viper.New()
	v.SetConfigName("config")
	v.SetConfigType("toml")
	v.AddConfigPath(configDir())

	v.SetEnvPrefix("HDF")
	v.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
	v.AutomaticEnv()

	if err := v.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); ok {
			if writeErr := writeDefault(); writeErr != nil {
				return &cfg, nil
			}
			return &cfg, nil
		}
		return &cfg, err
	}

	if err := v.Unmarshal(&cfg); err != nil {
		return &cfg, err
	}

	return &cfg, nil
}

func writeDefault() error {
	dir := configDir()
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return err
	}

	v := viper.New()
	v.SetConfigType("toml")
	v.Set("graceful_timeout", defaultConfig.GracefulTimeout.String())
	v.Set("protected", defaultConfig.Protected)
	v.Set("aliases", defaultConfig.Aliases)
	v.Set("default_force", defaultConfig.DefaultForce)
	v.Set("default_verbose", defaultConfig.DefaultVerbose)

	return v.WriteConfigAs(configPath())
}

func (c *Config) IsProtected(name string) bool {
	lower := strings.ToLower(name)
	for _, p := range c.Protected {
		if strings.ToLower(p) == lower {
			return true
		}
	}
	return false
}

func (c *Config) ResolveAlias(input string) string {
	if resolved, ok := c.Aliases[input]; ok {
		return resolved
	}
	return input
}
