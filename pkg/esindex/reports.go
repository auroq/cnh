package esindex

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/auroq/cnh/pkg/model"
	"github.com/auroq/cnh/pkg/reports"
)

const DefaultReportsIndex = "reports"

// ReportsMapping keeps every field exact: reports are filtered and counted, never searched by meaning.
var ReportsMapping = map[string]any{
	"mappings": map[string]any{
		"properties": map[string]any{
			"id":       map[string]any{"type": "keyword"},
			"place_id": map[string]any{"type": "keyword"},
			"verdict":  map[string]any{"type": "keyword"},
			"note":     map[string]any{"type": "text", "index": false},
			"at":       map[string]any{"type": "date"},
			"token":    map[string]any{"type": "keyword"},
			"source":   map[string]any{"type": "keyword"},
			"hidden":   map[string]any{"type": "boolean"},
		},
	},
}

const maxReportsPerQuery = 1000

// Reports stores attendee reports and looks place ids up; it is the reports use case's Store.
type Reports struct {
	Client       *Client
	ReportsIndex string
	PlacesIndex  string
}

var _ reports.Store = (*Reports)(nil)

func (r *Reports) index() string {
	if r.ReportsIndex == "" {
		return DefaultReportsIndex
	}
	return r.ReportsIndex
}

func (r *Reports) places() string {
	if r.PlacesIndex == "" {
		return DefaultPlacesIndex
	}
	return r.PlacesIndex
}

// EnsureIndex creates the reports index when it is missing; serve calls it at start so the first report has somewhere to go.
func (r *Reports) EnsureIndex(ctx context.Context) error {
	return r.Client.EnsureIndex(ctx, r.index(), ReportsMapping)
}

// Save writes the report and waits for the refresh, so the next answer already sees it.
func (r *Reports) Save(ctx context.Context, rep model.Report) error {
	return r.Client.Put(ctx, r.index(), rep.ID, rep)
}

func (r *Reports) Since(ctx context.Context, placeIDs []string, since time.Time) ([]model.Report, error) {
	raw, err := r.Client.Search(ctx, r.index(), map[string]any{
		"size": maxReportsPerQuery,
		"sort": []any{map[string]any{"at": "asc"}},
		"query": map[string]any{"bool": map[string]any{"filter": []any{
			map[string]any{"terms": map[string]any{"place_id": placeIDs}},
			map[string]any{"range": map[string]any{"at": map[string]any{"gte": since.UTC().Format(time.RFC3339)}}},
		}}},
	})
	if err != nil {
		return nil, err
	}
	return sources[model.Report](raw)
}

// Recent lists every report made at or after since, newest first, for the volunteer command.
func (r *Reports) Recent(ctx context.Context, since time.Time) ([]model.Report, error) {
	raw, err := r.Client.Search(ctx, r.index(), map[string]any{
		"size":  maxReportsPerQuery,
		"sort":  []any{map[string]any{"at": "desc"}},
		"query": map[string]any{"range": map[string]any{"at": map[string]any{"gte": since.UTC().Format(time.RFC3339)}}},
	})
	if err != nil {
		return nil, err
	}
	return sources[model.Report](raw)
}

// Hide marks a report so it stops counting; the document stays for the record.
func (r *Reports) Hide(ctx context.Context, id string) error {
	return r.Client.Update(ctx, r.index(), id, map[string]any{"hidden": true})
}

func (r *Reports) PlaceKind(ctx context.Context, id string) (string, error) {
	raw, err := r.Client.Search(ctx, r.places(), map[string]any{"size": 1, "_source": []string{"kind"}, "query": map[string]any{"term": map[string]any{"id": id}}})
	if err != nil {
		return "", err
	}
	hits, err := sources[struct {
		Kind string `json:"kind"`
	}](raw)
	if err != nil || len(hits) == 0 {
		return "", err
	}
	return hits[0].Kind, nil
}

// Put indexes one document by id and waits until a search would see it.
func (c *Client) Put(ctx context.Context, index, id string, body any) error {
	b, err := json.Marshal(body)
	if err != nil {
		return err
	}
	out, code, err := c.do(ctx, http.MethodPut, "/"+index+"/_doc/"+id+"?refresh=wait_for", bytes.NewReader(b), "application/json")
	if err != nil {
		return err
	}
	if code/100 != 2 {
		return fmt.Errorf("put %s/%s: %d %s", index, id, code, out)
	}
	return nil
}

// Update merges fields into one document by id and waits until a search would see it.
func (c *Client) Update(ctx context.Context, index, id string, fields map[string]any) error {
	b, err := json.Marshal(map[string]any{"doc": fields})
	if err != nil {
		return err
	}
	out, code, err := c.do(ctx, http.MethodPost, "/"+index+"/_update/"+id+"?refresh=wait_for", bytes.NewReader(b), "application/json")
	if err != nil {
		return err
	}
	if code/100 != 2 {
		return fmt.Errorf("update %s/%s: %d %s", index, id, code, out)
	}
	return nil
}
