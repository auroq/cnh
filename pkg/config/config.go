// Package config is the one place settings are named, defaulted and read.
// Precedence: flag, then environment (WIDEDOOR_, dots become underscores), then config file, then default.
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

	KeyServerListen = "server.listen"

	KeyElasticsearchURL           = "elasticsearch.url"
	KeyElasticsearchUsername      = "elasticsearch.username"
	KeyElasticsearchPassword      = "elasticsearch.password"
	KeyElasticsearchSessionsIndex = "elasticsearch.sessions_index"
	KeyElasticsearchPlacesIndex   = "elasticsearch.places_index"
	KeyElasticsearchNeedsIndex    = "elasticsearch.needs_index"
	KeyElasticsearchTimeout       = "elasticsearch.timeout"

	KeyIngestPlacesFile    = "ingest.places_file"
	KeyIngestSessionizeURL = "ingest.sessionize_url"
	KeyIngestConcurrency   = "ingest.concurrency"
	KeyIngestTimeout       = "ingest.timeout"
	KeyIngestFetchTimeout  = "ingest.fetch_timeout"
	KeyIngestSkipDetails   = "ingest.skip_details"
	KeyIngestBatch         = "ingest.batch"
	KeyIngestESTimeout     = "ingest.es_timeout"

	KeyVenueNeedsFile = "venue.needs_file"

	KeyTelemetryOTLPEndpoint = "telemetry.otlp_endpoint"
	KeyTelemetryServiceName  = "telemetry.service_name"
)

// Defaults is every setting's value when nothing else sets it; one place, read by --help too.
var Defaults = map[string]any{
	KeyServerListen:               ":8080",
	KeyElasticsearchURL:           "http://localhost:9200",
	KeyElasticsearchUsername:      "elastic",
	KeyElasticsearchPassword:      "",
	KeyElasticsearchSessionsIndex: "sessions",
	KeyElasticsearchPlacesIndex:   "places",
	KeyElasticsearchNeedsIndex:    "needs",
	KeyElasticsearchTimeout:       15 * time.Second,
	KeyIngestPlacesFile:           "data/venue/places.yaml",
	KeyIngestSessionizeURL:        "https://sessionize.com/api/v2/uo948i5p/view/markdown",
	KeyIngestConcurrency:          8,
	KeyIngestTimeout:              10 * time.Minute,
	KeyIngestFetchTimeout:         30 * time.Second,
	KeyIngestSkipDetails:          false,
	KeyIngestBatch:                50,
	KeyIngestESTimeout:            5 * time.Minute,
	KeyVenueNeedsFile:             "data/venue/needs.yaml",
	KeyTelemetryOTLPEndpoint:      "",
	KeyTelemetryServiceName:       "widedoor",
}

// secret keys never become flags and are never printed.
var secret = map[string]bool{KeyElasticsearchPassword: true}

// Server is what the HTTP server needs.
type Server interface {
	Listen() string
}

// Elasticsearch is what the search adapter needs.
type Elasticsearch interface {
	URL() string
	Username() string
	Password() string
	SessionsIndex() string
	PlacesIndex() string
	NeedsIndex() string
	Timeout() time.Duration
}

// Venue is what both serve and ingest need to know about the curated vocabulary.
type Venue interface {
	NeedsFile() string
}

// Ingest is what the ingest run needs.
type Ingest interface {
	PlacesFile() string
	SessionizeURL() string
	Concurrency() int
	Timeout() time.Duration
	FetchTimeout() time.Duration
	SkipDetails() bool
	Batch() int
	ESTimeout() time.Duration
}

// Telemetry is what the OpenTelemetry setup needs.
type Telemetry interface {
	OTLPEndpoint() string
	ServiceName() string
}

// Configuration hands each layer the slice it needs.
type Configuration struct {
	v *viper.Viper
}

func (c *Configuration) Server() Server               { return server{c.v} }
func (c *Configuration) Elasticsearch() Elasticsearch { return elasticsearch{c.v} }
func (c *Configuration) Ingest() Ingest               { return ingest{c.v} }
func (c *Configuration) Telemetry() Telemetry         { return telemetry{c.v} }
func (c *Configuration) Venue() Venue                 { return venueCfg{c.v} }

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

type server struct{ v *viper.Viper }

func (s server) Listen() string { return s.v.GetString(KeyServerListen) }

type elasticsearch struct{ v *viper.Viper }

func (e elasticsearch) URL() string            { return e.v.GetString(KeyElasticsearchURL) }
func (e elasticsearch) Username() string       { return e.v.GetString(KeyElasticsearchUsername) }
func (e elasticsearch) Password() string       { return e.v.GetString(KeyElasticsearchPassword) }
func (e elasticsearch) SessionsIndex() string  { return e.v.GetString(KeyElasticsearchSessionsIndex) }
func (e elasticsearch) PlacesIndex() string    { return e.v.GetString(KeyElasticsearchPlacesIndex) }
func (e elasticsearch) NeedsIndex() string     { return e.v.GetString(KeyElasticsearchNeedsIndex) }
func (e elasticsearch) Timeout() time.Duration { return e.v.GetDuration(KeyElasticsearchTimeout) }

type ingest struct{ v *viper.Viper }

func (i ingest) PlacesFile() string          { return i.v.GetString(KeyIngestPlacesFile) }
func (i ingest) SessionizeURL() string       { return i.v.GetString(KeyIngestSessionizeURL) }
func (i ingest) Concurrency() int            { return i.v.GetInt(KeyIngestConcurrency) }
func (i ingest) Timeout() time.Duration      { return i.v.GetDuration(KeyIngestTimeout) }
func (i ingest) FetchTimeout() time.Duration { return i.v.GetDuration(KeyIngestFetchTimeout) }
func (i ingest) SkipDetails() bool           { return i.v.GetBool(KeyIngestSkipDetails) }
func (i ingest) Batch() int                  { return i.v.GetInt(KeyIngestBatch) }
func (i ingest) ESTimeout() time.Duration    { return i.v.GetDuration(KeyIngestESTimeout) }

type telemetry struct{ v *viper.Viper }

func (t telemetry) OTLPEndpoint() string { return t.v.GetString(KeyTelemetryOTLPEndpoint) }
func (t telemetry) ServiceName() string  { return t.v.GetString(KeyTelemetryServiceName) }

type venueCfg struct{ v *viper.Viper }

func (c venueCfg) NeedsFile() string { return c.v.GetString(KeyVenueNeedsFile) }
