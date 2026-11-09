package config

import (
	"time"

	"github.com/spf13/viper"
)

const (
	KeyIngestPlacesFile    = "ingest.places_file"
	KeyIngestSessionizeURL = "ingest.sessionize_url"
	KeyIngestConcurrency   = "ingest.concurrency"
	KeyIngestTimeout       = "ingest.timeout"
	KeyIngestFetchTimeout  = "ingest.fetch_timeout"
	KeyIngestSkipDetails   = "ingest.skip_details"
	KeyIngestBatch         = "ingest.batch"
	KeyIngestESTimeout     = "ingest.es_timeout"
)

var ingestDefaults = map[string]any{
	KeyIngestPlacesFile:    "data/venue/places.yaml",
	KeyIngestSessionizeURL: "https://sessionize.com/api/v2/uo948i5p/view/markdown",
	KeyIngestConcurrency:   8,
	KeyIngestTimeout:       10 * time.Minute,
	KeyIngestFetchTimeout:  30 * time.Second,
	KeyIngestSkipDetails:   false,
	KeyIngestBatch:         50,
	KeyIngestESTimeout:     5 * time.Minute,
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

func (c *Configuration) Ingest() Ingest { return ingest{c.v} }

type ingest struct{ v *viper.Viper }

func (i ingest) PlacesFile() string          { return i.v.GetString(KeyIngestPlacesFile) }
func (i ingest) SessionizeURL() string       { return i.v.GetString(KeyIngestSessionizeURL) }
func (i ingest) Concurrency() int            { return i.v.GetInt(KeyIngestConcurrency) }
func (i ingest) Timeout() time.Duration      { return i.v.GetDuration(KeyIngestTimeout) }
func (i ingest) FetchTimeout() time.Duration { return i.v.GetDuration(KeyIngestFetchTimeout) }
func (i ingest) SkipDetails() bool           { return i.v.GetBool(KeyIngestSkipDetails) }
func (i ingest) Batch() int                  { return i.v.GetInt(KeyIngestBatch) }
func (i ingest) ESTimeout() time.Duration    { return i.v.GetDuration(KeyIngestESTimeout) }
