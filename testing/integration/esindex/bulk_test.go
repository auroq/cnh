package esindex_test

import (
	"context"
	"testing"
	"time"

	"github.com/auroq/cnh/pkg/esindex"
	"github.com/auroq/cnh/testing/helpers"
)

var mapping = map[string]any{
	"mappings": map[string]any{
		"properties": map[string]any{
			"title": map[string]any{"type": "keyword"},
		},
	},
}

func doc(id, title string) esindex.Doc {
	return esindex.Doc{ID: id, Body: map[string]string{"title": title}}
}

func bulk(t *testing.T, es *esindex.Client, index string, docs ...esindex.Doc) {
	t.Helper()
	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()
	failed, err := es.Bulk(ctx, index, docs, 50)
	if err != nil {
		t.Fatalf("bulk: %v", err)
	}
	if failed > 0 {
		t.Fatalf("bulk: %d documents failed", failed)
	}
}

func count(t *testing.T, es *esindex.Client, index string) int {
	t.Helper()
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	n, err := es.Count(ctx, index)
	if err != nil {
		t.Fatalf("count: %v", err)
	}
	return n
}

func TestBulk(t *testing.T) {
	t.Parallel()

	t.Run("when I bulk-index two docs", func(t *testing.T) {
		t.Parallel()
		es := helpers.ESClient(t)
		index := helpers.NewIndex(t, mapping)
		bulk(t, es, index, doc("1", "A"), doc("2", "B"))
		t.Run("it should count two", func(t *testing.T) {
			if got := count(t, es, index); got != 2 {
				t.Fatalf("count = %d, want 2", got)
			}
		})
	})

	t.Run("when I re-index the same id", func(t *testing.T) {
		t.Parallel()
		es := helpers.ESClient(t)
		index := helpers.NewIndex(t, mapping)
		bulk(t, es, index, doc("1", "A"))
		bulk(t, es, index, doc("1", "A, again"))
		t.Run("it should still count one", func(t *testing.T) {
			if got := count(t, es, index); got != 1 {
				t.Fatalf("count = %d, want 1", got)
			}
		})
	})
}
