package main

import (
	"testing"

	"github.com/auroq/cnh/pkg/sessionize"
)

func TestDocsForIndexSkipsSessionsWithoutDetails(t *testing.T) {
	sessions := []sessionize.Session{
		{ID: "described"},
		{ID: "failed"},
	}
	docs := docsForIndex(sessions, map[string]bool{"described": true}, false)

	if len(docs) != 1 {
		t.Fatalf("got %d docs, want 1", len(docs))
	}
	if docs[0].ID != "described" {
		t.Errorf("got document ID %q, want %q", docs[0].ID, "described")
	}
}

func TestDocsForIndexIncludesSessionsWhenDetailsAreSkipped(t *testing.T) {
	sessions := []sessionize.Session{{ID: "one"}, {ID: "two"}}
	docs := docsForIndex(sessions, nil, true)

	if len(docs) != len(sessions) {
		t.Fatalf("got %d docs, want %d", len(docs), len(sessions))
	}
	for i, session := range sessions {
		if docs[i].ID != session.ID {
			t.Errorf("doc %d has ID %q, want %q", i, docs[i].ID, session.ID)
		}
	}
}
