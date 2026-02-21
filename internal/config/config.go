package config

import (
	"fmt"
	"os"
	"path/filepath"
	"reflect"
	"strconv"
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

func Path() string {
	return filepath.Join(xdg.ConfigHome, "hdf", "config.toml")
}

func dir() string {
	return filepath.Dir(Path())
}

func structFieldForKey(cfg *Config, key string) (reflect.Value, bool) {
	v := reflect.ValueOf(cfg).Elem()
	t := v.Type()
	for i := range t.NumField() {
		if t.Field(i).Tag.Get("mapstructure") == key {
			return v.Field(i), true
		}
	}
	return reflect.Value{}, false
}

func newViper() *viper.Viper {
	v := viper.New()
	v.SetConfigName("config")
	v.SetConfigType("toml")
	v.AddConfigPath(dir())
	v.SetEnvPrefix("HDF")
	v.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
	v.AutomaticEnv()

	for _, f := range Schema {
		if f.Kind == Duration {
			if d, ok := f.Default.(time.Duration); ok {
				v.SetDefault(f.Key, d.String())
				continue
			}
		}
		v.SetDefault(f.Key, f.Default)
	}
	return v
}

func Load() (*Config, error) {
	v := newViper()

	if err := v.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); ok {
			_ = Reset()
		} else {
			return configFromDefaults(), err
		}
	}

	var cfg Config
	if err := v.Unmarshal(&cfg); err != nil {
		return configFromDefaults(), err
	}

	if cfg.Aliases == nil {
		cfg.Aliases = map[string]string{}
	}

	return &cfg, nil
}

func Save(cfg *Config) error {
	if err := os.MkdirAll(dir(), 0o755); err != nil {
		return err
	}

	v := viper.New()
	v.SetConfigType("toml")

	for _, f := range Schema {
		val, _ := GetValue(cfg, f.Key)
		if f.Kind == Duration {
			v.Set(f.Key, val.(time.Duration).String())
			continue
		}
		v.Set(f.Key, val)
	}

	return v.WriteConfigAs(Path())
}

func Reset() error {
	return Save(configFromDefaults())
}

func configFromDefaults() *Config {
	cfg := &Config{}
	for _, f := range Schema {
		_ = SetValue(cfg, f.Key, f.Default)
	}
	return cfg
}

func GetValue(cfg *Config, key string) (any, error) {
	fv, ok := structFieldForKey(cfg, key)
	if !ok {
		return nil, fmt.Errorf("unknown config key: %s", key)
	}
	return fv.Interface(), nil
}

func SetValue(cfg *Config, key string, val any) error {
	fv, ok := structFieldForKey(cfg, key)
	if !ok {
		return fmt.Errorf("unknown config key: %s", key)
	}
	rv := reflect.ValueOf(val)
	if !rv.Type().AssignableTo(fv.Type()) {
		return fmt.Errorf("type mismatch for %s: got %T, want %s", key, val, fv.Type())
	}
	fv.Set(rv)
	return nil
}

func ParseValue(f *Field, raw string) (any, error) {
	switch f.Kind {
	case Bool:
		return strconv.ParseBool(raw)
	case String:
		return raw, nil
	case Duration:
		return time.ParseDuration(raw)
	case StringSlice:
		parts := strings.Split(raw, ",")
		for i := range parts {
			parts[i] = strings.TrimSpace(parts[i])
		}
		return parts, nil
	case StringMap:
		return nil, fmt.Errorf("%q is a collection type and cannot be set directly", f.Key)
	default:
		return nil, fmt.Errorf("unknown kind %d", f.Kind)
	}
}

func FormatValue(f *Field, val any) string {
	switch f.Kind {
	case Bool:
		return strconv.FormatBool(val.(bool))
	case String:
		return fmt.Sprintf("%q", val.(string))
	case Duration:
		return fmt.Sprintf("%q", val.(time.Duration).String())
	case StringSlice:
		s := val.([]string)
		quoted := make([]string, len(s))
		for i, v := range s {
			quoted[i] = fmt.Sprintf("%q", v)
		}
		return "[" + strings.Join(quoted, ", ") + "]"
	case StringMap:
		m := val.(map[string]string)
		if len(m) == 0 {
			return "(none)"
		}
		var lines []string
		for k, v := range m {
			lines = append(lines, fmt.Sprintf("  %s = %q", k, v))
		}
		return strings.Join(lines, "\n")
	default:
		return fmt.Sprintf("%v", val)
	}
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
