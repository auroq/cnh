package main

import (
	"context"
	"flag"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/auroq/cnh/pkg/esindex"
	"github.com/auroq/cnh/pkg/sessionize"
	"github.com/auroq/cnh/pkg/venue"
)

func main() {
	fs := flag.NewFlagSet("ingest", flag.ExitOnError)
	fs.Usage = func() {
		fmt.Fprintf(fs.Output(), `usage: ingest [flags]

Fetches the KubeCon schedule from Sessionize and indexes it into Elasticsearch,
then indexes the curated venue places and the rooms the schedule names.
Safe to re-run: documents are keyed by id and overwritten.

Environment:
  ES_URL          Elasticsearch base URL (default http://localhost:9200)
  ES_USERNAME     basic auth user (default elastic)
  ES_PASSWORD     basic auth password
  ES_INDEX        sessions index (default sessions)
  PLACES_INDEX    places index (default places)
  PLACES_FILE     venue access YAML (default data/venue/places.yaml)
  SESSIONIZE_URL  markdown index URL (default %s)

Flags:
`, sessionize.DefaultIndexURL)
		fs.PrintDefaults()
	}
	skipDetails := fs.Bool("skip-details", false, "index only what the index page has, without per-session descriptions")
	timeout := fs.Duration("timeout", 10*time.Minute, "overall timeout")
	_ = fs.Parse(os.Args[1:])

	log := slog.New(slog.NewJSONHandler(os.Stdout, nil))
	ctx, cancel := context.WithTimeout(context.Background(), *timeout)
	defer cancel()

	indexer := &esindex.Indexer{
		Client: &esindex.Client{
			URL:      env("ES_URL", "http://localhost:9200"),
			Username: env("ES_USERNAME", "elastic"),
			Password: os.Getenv("ES_PASSWORD"),
			HTTP:     &http.Client{Timeout: 60 * time.Second},
		},
		SessionsIndex: env("ES_INDEX", esindex.DefaultSessionsIndex),
		PlacesIndex:   env("PLACES_INDEX", esindex.DefaultPlacesIndex),
		Batch:         50,
	}
	src := &sessionize.Client{HTTP: &http.Client{Timeout: 30 * time.Second}, IndexURL: os.Getenv("SESSIONIZE_URL")}

	if err := indexer.EnsureIndexes(ctx); err != nil {
		fatal(log, "ensure indexes", err)
	}
	sessions, err := src.Index(ctx)
	if err != nil {
		fatal(log, "fetch index", err)
	}
	log.Info("fetched index", "sessions", len(sessions))
	if !*skipDetails {
		failed := src.Describe(ctx, sessions, func(id string, err error) {
			if err != nil {
				log.Warn("describe failed", "session", id, "err", err.Error())
			}
		})
		log.Info("fetched details", "failed", failed)
	}

	sessionsResult, err := indexer.IndexSessions(ctx, sessions)
	if err != nil {
		fatal(log, "index sessions", err)
	}
	log.Info("indexed", "index", indexer.SessionsIndex, "sent", sessionsResult.Sent, "failed", sessionsResult.Failed, "count", sessionsResult.Count)

	places, err := loadPlaces(env("PLACES_FILE", "data/venue/places.yaml"))
	if err != nil {
		fatal(log, "load places", err)
	}
	rooms := venue.RoomsFromLocations(locations(sessions))
	placesResult, err := indexer.IndexPlaces(ctx, append(places, rooms...))
	if err != nil {
		fatal(log, "index places", err)
	}
	log.Info("indexed", "index", indexer.PlacesIndex, "curated", len(places), "rooms", len(rooms), "failed", placesResult.Failed, "count", placesResult.Count)

	if sessionsResult.Failed+placesResult.Failed > 0 {
		os.Exit(2)
	}
}

func loadPlaces(path string) ([]venue.Place, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	return venue.Load(f)
}

func locations(sessions []sessionize.Session) []string {
	out := make([]string, 0, len(sessions))
	for _, s := range sessions {
		out = append(out, s.Location)
	}
	return out
}

func env(k, def string) string {
	if v := strings.TrimSpace(os.Getenv(k)); v != "" {
		return v
	}
	return def
}

func fatal(log *slog.Logger, msg string, err error) {
	log.Error(msg, "err", err.Error())
	os.Exit(1)
}
