package config

import (
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

func Load() (*Config, error) {
	cfg := defaultConfig

	v := viper.New()
	v.SetConfigName("config")
	v.SetConfigType("toml")
	v.AddConfigPath(filepath.Join(xdg.ConfigHome, "hdf"))

	v.SetEnvPrefix("HDF")
	v.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
	v.AutomaticEnv()

	if err := v.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); !ok {
			return &cfg, err
		}
	}

	if err := v.Unmarshal(&cfg); err != nil {
		return &cfg, err
	}

	return &cfg, nil
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
