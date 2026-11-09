// Package ask answers "where is the nearest X from here" and free-text questions from the places and sessions indexes.
package ask

import (
	"context"
	"encoding/json"
	"fmt"
	"sort"
	"strings"
	"time"
)

type Searcher interface {
	Search(ctx context.Context, index string, body any) (json.RawMessage, error)
}

type Place struct {
	ID       string   `json:"id"`
	Kind     string   `json:"kind"`
	Name     string   `json:"name"`
	Venue    string   `json:"venue"`
	Level    string   `json:"level"`
	Near     string   `json:"near"`
	Notes    string   `json:"notes,omitempty"`
	For      []string `json:"for,omitempty"`
	Verified bool     `json:"verified"`
	Source   string   `json:"source"`
	Score    float64  `json:"-"`
}

type Session struct {
	ID       string    `json:"id"`
	Title    string    `json:"title"`
	URL      string    `json:"url"`
	Speakers []string  `json:"speakers"`
	Start    time.Time `json:"start"`
	End      time.Time `json:"end"`
	Location string    `json:"location"`
	Venue    string    `json:"venue"`
	Level    string    `json:"level"`
	Room     string    `json:"room"`
}

// Here is where the attendee is: a room they named, resolved to a venue and level.
type Here struct {
	Asked string `json:"asked"`
	Room  string `json:"room,omitempty"`
	Venue string `json:"venue,omitempty"`
	Level string `json:"level,omitempty"`
}

type Answer struct {
	Question string    `json:"question"`
	Here     Here      `json:"here"`
	Summary  string    `json:"summary"`
	Places   []Place   `json:"places"`
	Sessions []Session `json:"sessions"`
}

type Service struct {
	ES            Searcher
	PlacesIndex   string
	SessionsIndex string
}

func (s *Service) places() string {
	if s.PlacesIndex == "" {
		return "places"
	}
	return s.PlacesIndex
}

func (s *Service) sessions() string {
	if s.SessionsIndex == "" {
		return "sessions"
	}
	return s.SessionsIndex
}

// Locate resolves what the attendee typed as their location to a known room.
func (s *Service) Locate(ctx context.Context, asked string) (Here, error) {
	here := Here{Asked: strings.TrimSpace(asked)}
	if here.Asked == "" {
		return here, nil
	}
	body := map[string]any{
		"size": 1,
		"query": map[string]any{
			"bool": map[string]any{
				"filter": []any{map[string]any{"term": map[string]any{"kind": "room"}}},
				"should": []any{
					map[string]any{"match_phrase": map[string]any{"name": map[string]any{"query": here.Asked, "boost": 3}}},
					map[string]any{"match": map[string]any{"name": map[string]any{"query": here.Asked, "fuzziness": "AUTO", "operator": "and"}}},
				},
				"minimum_should_match": 1,
			},
		},
	}
	hits, err := s.searchPlaces(ctx, body)
	if err != nil || len(hits) == 0 {
		return here, err
	}
	here.Room, here.Venue, here.Level = hits[0].Name, hits[0].Venue, hits[0].Level
	return here, nil
}

// Ask runs the hybrid search over places and sessions and ranks places by nearness to here.
func (s *Service) Ask(ctx context.Context, question, from string) (Answer, error) {
	question = strings.TrimSpace(question)
	here, err := s.Locate(ctx, from)
	if err != nil {
		return Answer{}, err
	}
	ans := Answer{Question: question, Here: here}
	if question == "" {
		ans.Summary = "Ask where something is, such as the nearest all-gender restroom or quiet room, or search for a session."
		return ans, nil
	}

	places, err := s.searchPlaces(ctx, hybrid(question, []string{"name^3", "near^2", "notes", "kind"}, 10, map[string]any{"term": map[string]any{"kind": "room"}}))
	if err != nil {
		return ans, err
	}
	ans.Places = RankByNearness(places, here)
	if len(ans.Places) > 5 {
		ans.Places = ans.Places[:5]
	}

	raw, err := s.ES.Search(ctx, s.sessions(), hybrid(question, []string{"title^3", "description", "room.text", "speakers"}, 5, nil))
	if err != nil {
		return ans, err
	}
	var sr struct {
		Hits struct {
			Hits []struct {
				Source Session `json:"_source"`
			} `json:"hits"`
		} `json:"hits"`
	}
	if err := json.Unmarshal(raw, &sr); err != nil {
		return ans, err
	}
	for _, h := range sr.Hits.Hits {
		ans.Sessions = append(ans.Sessions, h.Source)
	}
	ans.Summary = Summarise(ans)
	return ans, nil
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

func (s *Service) searchPlaces(ctx context.Context, body any) ([]Place, error) {
	raw, err := s.ES.Search(ctx, s.places(), body)
	if err != nil {
		return nil, err
	}
	var pr struct {
		Hits struct {
			Hits []struct {
				Score  float64 `json:"_score"`
				Rank   int     `json:"_rank"`
				Source Place   `json:"_source"`
			} `json:"hits"`
		} `json:"hits"`
	}
	if err := json.Unmarshal(raw, &pr); err != nil {
		return nil, err
	}
	out := make([]Place, 0, len(pr.Hits.Hits))
	for i, h := range pr.Hits.Hits {
		p := h.Source
		p.Score = float64(len(pr.Hits.Hits) - i)
		out = append(out, p)
	}
	return out, nil
}

// RankByNearness keeps the search order but moves places on the attendee's level, then venue, to the front.
// Without coordinates this is the honest ordering: same level beats same building beats a different building.
func RankByNearness(places []Place, here Here) []Place {
	out := append([]Place(nil), places...)
	if here.Venue == "" {
		return out
	}
	tier := func(p Place) int {
		switch {
		case p.Venue == here.Venue && p.Level != "" && p.Level == here.Level:
			return 0
		case p.Venue == here.Venue:
			return 1
		default:
			return 2
		}
	}
	sort.SliceStable(out, func(i, j int) bool { return tier(out[i]) < tier(out[j]) })
	return out
}

// Summarise writes the one-sentence answer a screen reader hears first.
func Summarise(a Answer) string {
	if len(a.Places) == 0 && len(a.Sessions) == 0 {
		return "Nothing matched. Try other words, or ask staff at registration."
	}
	if len(a.Places) == 0 {
		return fmt.Sprintf("No place matched, but %d sessions did.", len(a.Sessions))
	}
	if len(a.Sessions) > 0 && (looksLikeSessionQuestion(a.Question) || len(a.Places) == 0) {
		s := a.Sessions[0]
		return fmt.Sprintf("%d sessions match. First: %s, %s, %s.", len(a.Sessions), s.Title, s.Start.In(mst).Format("Monday 3:04 pm"), strings.ReplaceAll(s.Location, " | ", ", "))
	}
	p := a.Places[0]
	where := p.Venue
	if p.Level != "" {
		where += ", " + p.Level
	}
	if p.Near != "" {
		where += ", " + strings.TrimRight(p.Near, ".")
	}
	sentence := fmt.Sprintf("%s: %s.", p.Name, where)
	switch {
	case a.Here.Venue == "" || a.Here.Level == "":
	case p.Venue == a.Here.Venue && p.Level == a.Here.Level:
		sentence += " Same level as you, no stairs or elevator."
	case p.Venue == a.Here.Venue:
		sentence += fmt.Sprintf(" Same building, you are on %s and it is on %s.", a.Here.Level, orUnknown(p.Level))
	default:
		sentence += fmt.Sprintf(" It is in the %s, you are in the %s.", p.Venue, a.Here.Venue)
	}
	if p.Verified {
		sentence += " Published by the organiser."
	} else {
		sentence += " Unverified: read off a plan, nobody has checked it yet."
	}
	return sentence
}

var mst = time.FixedZone("MST", -7*3600)

var sessionWords = []string{"talk", "session", "keynote", "speaker", "workshop", "panel", "lightning", "about", "presentation"}

func looksLikeSessionQuestion(q string) bool {
	q = strings.ToLower(q)
	for _, w := range sessionWords {
		if strings.Contains(q, w) {
			return true
		}
	}
	return false
}

func orUnknown(s string) string {
	if s == "" {
		return "an unlisted level"
	}
	return s
}
