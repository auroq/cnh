package esindex

import (
	"context"
	"encoding/json"
	"strings"
	"time"

	"github.com/auroq/cnh/pkg/ask"
	"github.com/auroq/cnh/pkg/ingest"
	"github.com/auroq/cnh/pkg/model"
	"github.com/auroq/cnh/pkg/venue"
)

const DefaultNeedsIndex = "needs"

// NeedsMapping indexes each need's name and phrases as one semantic_text field, so a paraphrase of a phrase still lands.
var NeedsMapping = map[string]any{
	"mappings": map[string]any{
		"properties": map[string]any{
			"id":          map[string]any{"type": "keyword"},
			"name":        map[string]any{"type": "text"},
			"phrases":     map[string]any{"type": "text"},
			"kinds":       map[string]any{"type": "keyword"},
			"note":        map[string]any{"type": "text"},
			"content":     map[string]any{"type": "semantic_text"},
			"ingested_at": map[string]any{"type": "date"},
		},
	},
}

type needDoc struct {
	ID      string   `json:"id"`
	Name    string   `json:"name"`
	Phrases []string `json:"phrases"`
	Kinds   []string `json:"kinds"`
	Note    string   `json:"note,omitempty"`
	Content string   `json:"content,omitempty"`
}

func toNeedDoc(n venue.Need) needDoc {
	return needDoc{ID: n.ID, Name: n.Name, Phrases: n.Phrases, Kinds: n.Kinds, Note: n.Note, Content: needContent(n)}
}

func (d needDoc) model() model.Need {
	return model.Need{ID: d.ID, Name: d.Name, Phrases: d.Phrases, Kinds: d.Kinds, Note: d.Note}
}

func needContent(n venue.Need) string {
	return strings.Join(append([]string{n.Name}, n.Phrases...), ". ")
}

func (x *Indexer) needs() string {
	if x.NeedsIndex == "" {
		return DefaultNeedsIndex
	}
	return x.NeedsIndex
}

func (x *Indexer) IndexNeeds(ctx context.Context, needs []venue.Need) (ingest.Result, error) {
	now := x.now()
	docs := make([]Doc, 0, len(needs))
	for _, n := range needs {
		docs = append(docs, Doc{ID: n.ID, Body: struct {
			needDoc
			IngestedAt time.Time `json:"ingested_at"`
		}{toNeedDoc(n), now}})
	}
	return x.write(ctx, x.needs(), docs)
}

func (s *Store) needsIndex() string {
	if s.NeedsIndex == "" {
		return DefaultNeedsIndex
	}
	return s.NeedsIndex
}

// FindNeeds asks the semantic_text field which needs the question is about, best first, with the score so the use case can decide.
func (s *Store) FindNeeds(ctx context.Context, question string) ([]ask.NeedMatch, error) {
	raw, err := s.Client.Search(ctx, s.needsIndex(), needsQuery(question))
	if err != nil {
		return nil, err
	}
	var res struct {
		Hits struct {
			Hits []struct {
				Score  float64 `json:"_score"`
				Source needDoc `json:"_source"`
			} `json:"hits"`
		} `json:"hits"`
	}
	if err := json.Unmarshal(raw, &res); err != nil {
		return nil, err
	}
	out := make([]ask.NeedMatch, 0, len(res.Hits.Hits))
	for _, h := range res.Hits.Hits {
		out = append(out, ask.NeedMatch{Need: h.Source.model(), Score: h.Score})
	}
	return out, nil
}

func needsQuery(question string) map[string]any {
	return map[string]any{
		"size":    3,
		"_source": []string{"id", "name", "phrases", "kinds", "note"},
		"query":   map[string]any{"semantic": map[string]any{"field": "content", "query": question}},
	}
}
