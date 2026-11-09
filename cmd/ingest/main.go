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
)

func main() {
	fs := flag.NewFlagSet("ingest", flag.ExitOnError)
	fs.Usage = func() {
		fmt.Fprintf(fs.Output(), `usage: ingest [flags]

Fetches the KubeCon schedule from Sessionize and indexes it into Elasticsearch.
Safe to re-run: documents are keyed by session id and overwritten.

Environment:
  ES_URL          Elasticsearch base URL (default http://localhost:9200)
  ES_USERNAME     basic auth user (default elastic)
  ES_PASSWORD     basic auth password
  ES_INDEX        target index (default sessions)
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

	es := &esindex.Client{
		URL:      env("ES_URL", "http://localhost:9200"),
		Username: env("ES_USERNAME", "elastic"),
		Password: os.Getenv("ES_PASSWORD"),
		HTTP:     &http.Client{Timeout: 60 * time.Second},
	}
	index := env("ES_INDEX", "sessions")
	src := &sessionize.Client{HTTP: &http.Client{Timeout: 30 * time.Second}, IndexURL: os.Getenv("SESSIONIZE_URL")}

	if err := es.EnsureIndex(ctx, index, sessionsMapping); err != nil {
		fatal(log, "ensure index", err)
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

	docs := make([]esindex.Doc, 0, len(sessions))
	for _, s := range sessions {
		docs = append(docs, esindex.Doc{ID: s.ID, Body: toDoc(s)})
	}
	failed, err := es.Bulk(ctx, index, docs, 50)
	if err != nil {
		fatal(log, "bulk index", err)
	}
	count, _ := es.Count(ctx, index)
	log.Info("indexed", "index", index, "sent", len(docs), "failed", failed, "count", count)
	if failed > 0 {
		os.Exit(2)
	}
}

func toDoc(s sessionize.Session) map[string]any {
	content := s.Title
	if s.Description != "" {
		content += "\n\n" + s.Description
	}
	return map[string]any{
		"id":          s.ID,
		"title":       s.Title,
		"url":         s.URL,
		"speakers":    s.Speakers,
		"start":       s.Start,
		"end":         s.End,
		"location":    s.Location,
		"venue":       s.Venue,
		"level":       s.Level,
		"room":        s.Room,
		"description": s.Description,
		"content":     content,
		"ingested_at": time.Now().UTC(),
	}
}

var sessionsMapping = map[string]any{
	"mappings": map[string]any{
		"properties": map[string]any{
			"id":          map[string]any{"type": "keyword"},
			"title":       map[string]any{"type": "text", "fields": map[string]any{"keyword": map[string]any{"type": "keyword"}}},
			"url":         map[string]any{"type": "keyword", "index": false},
			"speakers":    map[string]any{"type": "keyword"},
			"start":       map[string]any{"type": "date"},
			"end":         map[string]any{"type": "date"},
			"location":    map[string]any{"type": "keyword"},
			"venue":       map[string]any{"type": "keyword"},
			"level":       map[string]any{"type": "keyword"},
			"room":        map[string]any{"type": "keyword", "fields": map[string]any{"text": map[string]any{"type": "text"}}},
			"description": map[string]any{"type": "text"},
			"content":     map[string]any{"type": "semantic_text"},
			"ingested_at": map[string]any{"type": "date"},
		},
	},
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
