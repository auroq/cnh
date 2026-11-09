package config

import (
	"time"

	"github.com/spf13/viper"
)

const (
	KeyElasticsearchURL           = "elasticsearch.url"
	KeyElasticsearchUsername      = "elasticsearch.username"
	KeyElasticsearchPassword      = "elasticsearch.password"
	KeyElasticsearchSessionsIndex = "elasticsearch.sessions_index"
	KeyElasticsearchPlacesIndex   = "elasticsearch.places_index"
	KeyElasticsearchNeedsIndex    = "elasticsearch.needs_index"
	KeyElasticsearchReportsIndex  = "elasticsearch.reports_index"
	KeyElasticsearchTimeout       = "elasticsearch.timeout"
)

var elasticsearchDefaults = map[string]any{
	KeyElasticsearchURL:           "http://localhost:9200",
	KeyElasticsearchUsername:      "elastic",
	KeyElasticsearchPassword:      "",
	KeyElasticsearchSessionsIndex: "sessions",
	KeyElasticsearchPlacesIndex:   "places",
	KeyElasticsearchNeedsIndex:    "needs",
	KeyElasticsearchReportsIndex:  "reports",
	KeyElasticsearchTimeout:       15 * time.Second,
}

// Elasticsearch is what the search adapter needs.
type Elasticsearch interface {
	URL() string
	Username() string
	Password() string
	SessionsIndex() string
	PlacesIndex() string
	NeedsIndex() string
	ReportsIndex() string
	Timeout() time.Duration
}

func (c *Configuration) Elasticsearch() Elasticsearch { return elasticsearch{c.v} }

type elasticsearch struct{ v *viper.Viper }

func (e elasticsearch) URL() string            { return e.v.GetString(KeyElasticsearchURL) }
func (e elasticsearch) Username() string       { return e.v.GetString(KeyElasticsearchUsername) }
func (e elasticsearch) Password() string       { return e.v.GetString(KeyElasticsearchPassword) }
func (e elasticsearch) SessionsIndex() string  { return e.v.GetString(KeyElasticsearchSessionsIndex) }
func (e elasticsearch) PlacesIndex() string    { return e.v.GetString(KeyElasticsearchPlacesIndex) }
func (e elasticsearch) NeedsIndex() string     { return e.v.GetString(KeyElasticsearchNeedsIndex) }
func (e elasticsearch) ReportsIndex() string   { return e.v.GetString(KeyElasticsearchReportsIndex) }
func (e elasticsearch) Timeout() time.Duration { return e.v.GetDuration(KeyElasticsearchTimeout) }
