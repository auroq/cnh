package esindex

import (
	"context"
	"time"

	"github.com/auroq/cnh/pkg/sessionize"
	"github.com/auroq/cnh/pkg/venue"
)

// Indexer owns the index mappings and the document shapes; the ingest only hands it sessions and places.
type Indexer struct {
	Client        *Client
	SessionsIndex string
	PlacesIndex   string
	Batch         int
	Now           func() time.Time
}

// Result is what one bulk write did.
type Result struct {
	Sent   int
	Failed int
	Count  int
}

func (x *Indexer) sessions() string {
	if x.SessionsIndex == "" {
		return DefaultSessionsIndex
	}
	return x.SessionsIndex
}

func (x *Indexer) places() string {
	if x.PlacesIndex == "" {
		return DefaultPlacesIndex
	}
	return x.PlacesIndex
}

func (x *Indexer) now() time.Time {
	if x.Now == nil {
		return time.Now().UTC()
	}
	return x.Now()
}

func (x *Indexer) EnsureIndexes(ctx context.Context) error {
	if err := x.Client.EnsureIndex(ctx, x.sessions(), SessionsMapping); err != nil {
		return err
	}
	return x.Client.EnsureIndex(ctx, x.places(), PlacesMapping)
}

func (x *Indexer) IndexSessions(ctx context.Context, sessions []sessionize.Session) (Result, error) {
	now := x.now()
	docs := make([]Doc, 0, len(sessions))
	for _, s := range sessions {
		docs = append(docs, Doc{ID: s.ID, Body: sessionDoc(s, now)})
	}
	return x.write(ctx, x.sessions(), docs)
}

func (x *Indexer) IndexPlaces(ctx context.Context, places []venue.Place) (Result, error) {
	now := x.now()
	docs := make([]Doc, 0, len(places))
	for _, p := range places {
		docs = append(docs, Doc{ID: p.ID, Body: placeDoc(p, now)})
	}
	return x.write(ctx, x.places(), docs)
}

func (x *Indexer) write(ctx context.Context, index string, docs []Doc) (Result, error) {
	failed, err := x.Client.Bulk(ctx, index, docs, x.Batch)
	res := Result{Sent: len(docs), Failed: failed}
	if err != nil {
		return res, err
	}
	res.Count, _ = x.Client.Count(ctx, index)
	return res, nil
}

func sessionDoc(s sessionize.Session, now time.Time) map[string]any {
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
		"ingested_at": now,
	}
}

func placeDoc(p venue.Place, now time.Time) map[string]any {
	return map[string]any{
		"id": p.ID, "kind": p.Kind, "name": p.Name, "venue": p.Venue, "level": p.Level,
		"near": p.Near, "notes": p.Notes, "for": p.For, "verified": p.Verified, "source": p.Source,
		"content": p.Content(), "ingested_at": now,
	}
}

var SessionsMapping = map[string]any{
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

var PlacesMapping = map[string]any{
	"mappings": map[string]any{
		"properties": map[string]any{
			"id":          map[string]any{"type": "keyword"},
			"kind":        map[string]any{"type": "keyword"},
			"name":        map[string]any{"type": "text", "fields": map[string]any{"keyword": map[string]any{"type": "keyword"}}},
			"venue":       map[string]any{"type": "keyword"},
			"level":       map[string]any{"type": "keyword"},
			"near":        map[string]any{"type": "text"},
			"notes":       map[string]any{"type": "text"},
			"for":         map[string]any{"type": "keyword"},
			"verified":    map[string]any{"type": "boolean"},
			"source":      map[string]any{"type": "keyword", "index": false},
			"content":     map[string]any{"type": "semantic_text"},
			"ingested_at": map[string]any{"type": "date"},
		},
	},
}
