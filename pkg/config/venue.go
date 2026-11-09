package config

import "github.com/spf13/viper"

const (
	KeyVenueNeedsFile        = "venue.needs_file"
	KeyVenueBlockedWordsFile = "venue.blocked_words_file"
)

var venueDefaults = map[string]any{
	KeyVenueNeedsFile:        "data/venue/needs.yaml",
	KeyVenueBlockedWordsFile: "data/venue/blocked-words.txt",
}

// Venue is what both serve and ingest need to know about the curated vocabulary.
type Venue interface {
	NeedsFile() string
	BlockedWordsFile() string
}

func (c *Configuration) Venue() Venue { return venueCfg{c.v} }

type venueCfg struct{ v *viper.Viper }

func (c venueCfg) NeedsFile() string        { return c.v.GetString(KeyVenueNeedsFile) }
func (c venueCfg) BlockedWordsFile() string { return c.v.GetString(KeyVenueBlockedWordsFile) }
