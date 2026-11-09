// Package config is the one place settings are named, defaulted and read.
// Precedence: flag, then environment (WIDEDOOR_, dots become underscores), then config file, then default.
// One file per area (server, elasticsearch, ingest, venue, telemetry, qr) holds that area's keys, defaults, interface and accessors; this file holds what is common.
package config

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/spf13/pflag"
	"github.com/spf13/viper"
)

const (
	EnvPrefix  = "WIDEDOOR"
	ConfigName = "config"
	ConfigType = "yaml"

	KeyConfig = "config"
)

// Defaults is every setting's value when nothing else sets it; one map, assembled from the areas, read by --help too.
var Defaults = merge(serverDefaults, elasticsearchDefaults, ingestDefaults, venueDefaults, telemetryDefaults, qrDefaults)

// secret keys never become flags and are never printed.
var secret = map[string]bool{KeyElasticsearchPassword: true}

// Configuration hands each layer the slice it needs.
type Configuration struct {
	v *viper.Viper
}

// New prepares a viper with defaults, the environment and the config file search path, reading nothing yet.
func New() *viper.Viper {
	v := viper.New()
	for k, d := range Defaults {
		v.SetDefault(k, d)
	}
	v.SetEnvPrefix(EnvPrefix)
	v.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
	v.AutomaticEnv()
	v.SetConfigName(ConfigName)
	v.SetConfigType(ConfigType)
	v.AddConfigPath(".")
	v.AddConfigPath("/etc/widedoor")
	if exe, err := os.Executable(); err == nil {
		v.AddConfigPath(filepath.Dir(exe))
	}
	return v
}

// BindFlags declares a flag for every non-secret setting and binds it to its key.
func BindFlags(v *viper.Viper, flags *pflag.FlagSet) {
	flags.String(KeyConfig, "", "config file (default: config.yaml beside the binary, in /etc/widedoor, or in the working directory)")
	_ = v.BindPFlag(KeyConfig, flags.Lookup(KeyConfig))
	for _, key := range sortedKeys() {
		if secret[key] {
			continue
		}
		name := flagName(key)
		help := fmt.Sprintf("%s (env %s)", key, EnvName(key))
		switch d := Defaults[key].(type) {
		case string:
			flags.String(name, d, help)
		case int:
			flags.Int(name, d, help)
		case bool:
			flags.Bool(name, d, help)
		case time.Duration:
			flags.Duration(name, d, help)
		}
		_ = v.BindPFlag(key, flags.Lookup(name))
	}
}

// Load reads the config file if there is one and returns the typed configuration.
func Load(v *viper.Viper) (*Configuration, error) {
	if path := v.GetString(KeyConfig); path != "" {
		v.SetConfigFile(path)
	}
	if err := v.ReadInConfig(); err != nil {
		var notFound viper.ConfigFileNotFoundError
		if !errors.As(err, &notFound) && v.GetString(KeyConfig) != "" {
			return nil, fmt.Errorf("read config %s: %w", v.GetString(KeyConfig), err)
		}
		if !errors.As(err, &notFound) && !os.IsNotExist(err) {
			return nil, fmt.Errorf("read config: %w", err)
		}
	}
	return &Configuration{v: v}, nil
}

// EnvName is the environment variable for a key: WIDEDOOR_ELASTICSEARCH_URL for elasticsearch.url.
func EnvName(key string) string {
	return EnvPrefix + "_" + strings.ToUpper(strings.ReplaceAll(key, ".", "_"))
}

func flagName(key string) string {
	return strings.ReplaceAll(strings.ReplaceAll(key, ".", "-"), "_", "-")
}

func sortedKeys() []string {
	keys := make([]string, 0, len(Defaults))
	for k := range Defaults {
		keys = append(keys, k)
	}
	for i := 1; i < len(keys); i++ {
		for j := i; j > 0 && keys[j] < keys[j-1]; j-- {
			keys[j], keys[j-1] = keys[j-1], keys[j]
		}
	}
	return keys
}

func merge(areas ...map[string]any) map[string]any {
	out := map[string]any{}
	for _, area := range areas {
		for k, d := range area {
			if _, dup := out[k]; dup {
				panic("config: key defaulted twice: " + k)
			}
			out[k] = d
		}
	}
	return out
}
