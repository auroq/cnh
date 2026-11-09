package esindex

import (
	"context"
	"encoding/json"

	"github.com/auroq/cnh/pkg/ask"
)

// Store answers ask's questions with Elasticsearch: hybrid BM25 plus semantic_text retrievers fused with RRF.
type Store struct {
	Client        *Client
	PlacesIndex   string
	SessionsIndex string
}

var _ ask.Store = (*Store)(nil)

const (
	DefaultPlacesIndex   = "places"
	DefaultSessionsIndex = "sessions"
)

var (
	placeFields   = []string{"name^3", "near^2", "notes", "kind"}
	sessionFields = []string{"title^3", "description", "room.text", "speakers"}
	roomsOnly     = map[string]any{"term": map[string]any{"kind": "room"}}
)

func (s *Store) places() string {
	if s.PlacesIndex == "" {
		return DefaultPlacesIndex
	}
	return s.PlacesIndex
}

func (s *Store) sessions() string {
	if s.SessionsIndex == "" {
		return DefaultSessionsIndex
	}
	return s.SessionsIndex
}

func (s *Store) LocateRoom(ctx context.Context, asked string) (ask.Here, error) {
	hits, err := s.searchPlaces(ctx, locateQuery(asked))
	if err != nil || len(hits) == 0 {
		return ask.Here{}, err
	}
	return ask.Here{Room: hits[0].Name, Venue: hits[0].Venue, Level: hits[0].Level}, nil
}

func (s *Store) FindPlaces(ctx context.Context, question string) ([]ask.Place, error) {
	return s.searchPlaces(ctx, hybrid(question, placeFields, 10, roomsOnly))
}

func (s *Store) FindSessions(ctx context.Context, question string) ([]ask.Session, error) {
	raw, err := s.Client.Search(ctx, s.sessions(), hybrid(question, sessionFields, 5, nil))
	if err != nil {
		return nil, err
	}
	return sources[ask.Session](raw)
}

func (s *Store) searchPlaces(ctx context.Context, body any) ([]ask.Place, error) {
	raw, err := s.Client.Search(ctx, s.places(), body)
	if err != nil {
		return nil, err
	}
	return sources[ask.Place](raw)
}

func locateQuery(asked string) map[string]any {
	return map[string]any{
		"size": 1,
		"query": map[string]any{
			"bool": map[string]any{
				"filter": []any{roomsOnly},
				"should": []any{
					map[string]any{"match_phrase": map[string]any{"name": map[string]any{"query": asked, "boost": 3}}},
					map[string]any{"match": map[string]any{"name": map[string]any{"query": asked, "fuzziness": "AUTO", "operator": "and"}}},
					map[string]any{"match": map[string]any{"near": asked}},
				},
				"minimum_should_match": 1,
			},
		},
	}
}

func hybrid(q string, fields []string, size int, mustNot any) map[string]any {
	lexical := map[string]any{"multi_match": map[string]any{"query": q, "fields": fields, "fuzziness": "AUTO"}}
	semantic := map[string]any{"semantic": map[string]any{"field": "content", "query": q}}
	if mustNot != nil {
		lexical = map[string]any{"bool": map[string]any{"must": lexical, "must_not": mustNot}}
		semantic = map[string]any{"bool": map[string]any{"must": semantic, "must_not": mustNot}}
	}
	return map[string]any{
		"size": size,
		"retriever": map[string]any{
			"rrf": map[string]any{
				"rank_window_size": 20,
				"retrievers": []any{
					map[string]any{"standard": map[string]any{"query": lexical}},
					map[string]any{"standard": map[string]any{"query": semantic}},
				},
			},
		},
	}
}

func sources[T any](raw json.RawMessage) ([]T, error) {
	var res struct {
		Hits struct {
			Hits []struct {
				Source T `json:"_source"`
			} `json:"hits"`
		} `json:"hits"`
	}
	if err := json.Unmarshal(raw, &res); err != nil {
		return nil, err
	}
	out := make([]T, 0, len(res.Hits.Hits))
	for _, h := range res.Hits.Hits {
		out = append(out, h.Source)
	}
	return out, nil
}
