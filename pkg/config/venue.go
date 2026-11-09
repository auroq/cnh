package config

import "github.com/spf13/viper"

const KeyVenueNeedsFile = "venue.needs_file"

var venueDefaults = map[string]any{
	KeyVenueNeedsFile: "data/venue/needs.yaml",
}

// Venue is what both serve and ingest need to know about the curated vocabulary.
type Venue interface {
	NeedsFile() string
}

func (c *Configuration) Venue() Venue { return venueCfg{c.v} }

type venueCfg struct{ v *viper.Viper }

func (c venueCfg) NeedsFile() string { return c.v.GetString(KeyVenueNeedsFile) }
