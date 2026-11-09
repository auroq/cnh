package esindex

import (
	"context"
	"encoding/json"
	"strings"
	"time"
	"unicode"

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

func (s *Store) Ping(ctx context.Context) error {
	return s.Client.Ping(ctx)
}

func (s *Store) LocateRoom(ctx context.Context, name string) (ask.Here, error) {
	return s.firstPlace(ctx, roomQuery(map[string]any{"match_phrase": map[string]any{"name": name}}))
}

func (s *Store) LocateRoomFuzzy(ctx context.Context, asked string) (ask.Here, error) {
	return s.firstPlace(ctx, roomQuery(map[string]any{"bool": map[string]any{
		"should": []any{
			map[string]any{"match": map[string]any{"name": map[string]any{"query": asked, "fuzziness": "AUTO", "operator": "and"}}},
			map[string]any{"match": map[string]any{"near": asked}},
		},
		"minimum_should_match": 1,
	}}))
}

// LocateAlias matches the meaningful words of what was said against the phrases people use for a place;
// filler such as "the", "I'm at" and "just left" is dropped first, so "the 100 S side" finds the alias "100 S side".
func (s *Store) LocateAlias(ctx context.Context, phrase string) (ask.Here, error) {
	words := meaningful(phrase)
	if words == "" {
		return ask.Here{}, nil
	}
	return s.firstPlace(ctx, map[string]any{"size": 1, "query": map[string]any{
		"match": map[string]any{"aliases": map[string]any{"query": words, "operator": "and"}},
	}})
}

// LocateSession finds the session whose title has every meaningful word said, preferring the one that ended closest to now.
func (s *Store) LocateSession(ctx context.Context, title string, now time.Time) (ask.Here, error) {
	words := meaningful(title)
	if words == "" {
		return ask.Here{}, nil
	}
	raw, err := s.Client.Search(ctx, s.sessions(), map[string]any{"size": 1, "query": map[string]any{"bool": map[string]any{
		"must":   []any{map[string]any{"match": map[string]any{"title": map[string]any{"query": words, "operator": "and"}}}},
		"should": []any{map[string]any{"distance_feature": map[string]any{"field": "end", "origin": now.UTC().Format(time.RFC3339), "pivot": "30m"}}},
	}}})
	if err != nil {
		return ask.Here{}, err
	}
	hits, err := sources[ask.Session](raw)
	if err != nil || len(hits) == 0 {
		return ask.Here{}, err
	}
	return ask.Here{Room: hits[0].Room, Venue: hits[0].Venue, Level: hits[0].Level}, nil
}

// LocateNearby asks the semantic field which place the words describe, and answers only when the match is a confident one.
func (s *Store) LocateNearby(ctx context.Context, phrase string) (ask.Here, error) {
	return s.firstPlace(ctx, map[string]any{"size": 1, "min_score": nearbyMinScore, "query": map[string]any{
		"semantic": map[string]any{"field": "content", "query": phrase},
	}})
}

func (s *Store) firstPlace(ctx context.Context, body any) (ask.Here, error) {
	hits, err := s.searchPlaces(ctx, body)
	if err != nil || len(hits) == 0 {
		return ask.Here{}, err
	}
	return ask.Here{Room: hits[0].Name, Venue: hits[0].Venue, Level: hits[0].Level, Side: hits[0].Side}, nil
}

// LocatePlace resolves a place id, as a door code carries it, to where that place is; any kind, not only rooms.
func (s *Store) LocatePlace(ctx context.Context, id string) (ask.Here, error) {
	hits, err := s.searchPlaces(ctx, map[string]any{"size": 1, "query": map[string]any{"term": map[string]any{"id": id}}})
	if err != nil || len(hits) == 0 {
		return ask.Here{}, err
	}
	return ask.Here{Room: hits[0].Name, Venue: hits[0].Venue, Level: hits[0].Level, Side: hits[0].Side}, nil
}

func (s *Store) FindPlaces(ctx context.Context, question, kind string) ([]ask.Place, error) {
	var onlyKind any
	if kind != "" {
		onlyKind = map[string]any{"term": map[string]any{"kind": kind}}
	}
	return s.searchPlaces(ctx, hybrid(question, placeFields, 10, onlyKind, roomsOnly))
}

func (s *Store) FindSessions(ctx context.Context, question string) ([]ask.Session, error) {
	raw, err := s.Client.Search(ctx, s.sessions(), hybrid(question, sessionFields, 5, nil, nil))
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

func roomQuery(match any) map[string]any {
	return map[string]any{"size": 1, "query": map[string]any{"bool": map[string]any{"filter": []any{roomsOnly}, "must": []any{match}}}}
}

const nearbyMinScore = 8.0

var filler = map[string]bool{}

func init() {
	for _, w := range strings.Fields("a an and are as at be but by for if in into is it no not of on or such that the their then there these they this to was will with i im m am my me we you just got left came coming out from near here where outside inside now right") {
		filler[w] = true
	}
}

// meaningful lowercases, splits on anything that is not a letter or digit, and drops the filler words.
func meaningful(s string) string {
	var out []string
	for _, w := range strings.FieldsFunc(strings.ToLower(s), func(r rune) bool { return !unicode.IsLetter(r) && !unicode.IsDigit(r) }) {
		if !filler[w] {
			out = append(out, w)
		}
	}
	return strings.Join(out, " ")
}

func hybrid(q string, fields []string, size int, filter, mustNot any) map[string]any {
	lexical := map[string]any{"multi_match": map[string]any{"query": q, "fields": fields, "fuzziness": "AUTO"}}
	semantic := map[string]any{"semantic": map[string]any{"field": "content", "query": q}}
	if filter != nil || mustNot != nil {
		lexical = narrowed(lexical, filter, mustNot)
		semantic = narrowed(semantic, filter, mustNot)
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

func narrowed(must, filter, mustNot any) map[string]any {
	b := map[string]any{"must": must}
	if filter != nil {
		b["filter"] = []any{filter}
	}
	if mustNot != nil {
		b["must_not"] = mustNot
	}
	return map[string]any{"bool": b}
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
