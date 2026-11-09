package esindex

import (
	"context"

	"github.com/auroq/cnh/pkg/ask"
)

var _ ask.SessionFinder = (*Store)(nil)

// MatchSession finds the session a pasted title names: the exact title first, then the best match that has most of the words.
func (s *Store) MatchSession(ctx context.Context, title string) (ask.Session, bool, error) {
	exact := map[string]any{"term": map[string]any{"title.keyword": title}}
	fuzzy := map[string]any{"match": map[string]any{"title": map[string]any{"query": title, "fuzziness": "AUTO", "minimum_should_match": "75%"}}}
	for _, query := range []map[string]any{exact, fuzzy} {
		raw, err := s.Client.Search(ctx, s.sessions(), map[string]any{"size": 1, "query": query})
		if err != nil {
			return ask.Session{}, false, err
		}
		hits, err := sources[ask.Session](raw)
		if err != nil {
			return ask.Session{}, false, err
		}
		if len(hits) > 0 {
			return hits[0], true, nil
		}
	}
	return ask.Session{}, false, nil
}

// SessionsByID loads the sessions an attendee kept, in no particular order.
func (s *Store) SessionsByID(ctx context.Context, ids []string) ([]ask.Session, error) {
	if len(ids) == 0 {
		return nil, nil
	}
	raw, err := s.Client.Search(ctx, s.sessions(), map[string]any{"size": len(ids), "query": map[string]any{"terms": map[string]any{"id": ids}}})
	if err != nil {
		return nil, err
	}
	return sources[ask.Session](raw)
}
