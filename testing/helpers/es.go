package helpers

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"net/http"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/auroq/cnh/pkg/esindex"
)

func ESClient(t *testing.T) *esindex.Client {
	t.Helper()
	u := strings.TrimSpace(os.Getenv("ES_URL"))
	if u == "" {
		t.Skip("ES_URL is unset; run `make test-integration` to reach the cluster's Elasticsearch, or set ES_URL, ES_USERNAME and ES_PASSWORD")
	}
	user := strings.TrimSpace(os.Getenv("ES_USERNAME"))
	if user == "" {
		user = "elastic"
	}
	return &esindex.Client{
		URL:      u,
		Username: user,
		Password: os.Getenv("ES_PASSWORD"),
		HTTP:     &http.Client{Timeout: 60 * time.Second},
	}
}

func NewIndex(t *testing.T, mapping any) string {
	t.Helper()
	es := ESClient(t)
	name := indexName(t)
	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()
	if err := es.EnsureIndex(ctx, name, mapping); err != nil {
		t.Fatalf("create index %s: %v", name, err)
	}
	t.Cleanup(func() {
		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()
		if err := deleteIndex(ctx, es, name); err != nil {
			t.Errorf("delete index %s: %v", name, err)
		}
	})
	return name
}

func indexName(t *testing.T) string {
	t.Helper()
	var b strings.Builder
	dash := false
	for _, r := range strings.ToLower(t.Name()) {
		switch {
		case r >= 'a' && r <= 'z', r >= '0' && r <= '9':
			b.WriteRune(r)
			dash = false
		case !dash && b.Len() > 0:
			b.WriteByte('-')
			dash = true
		}
	}
	slug := strings.TrimRight(b.String(), "-")
	if len(slug) > 60 {
		slug = strings.TrimRight(slug[:60], "-")
	}
	suffix := make([]byte, 4)
	if _, err := rand.Read(suffix); err != nil {
		t.Fatalf("random index suffix: %v", err)
	}
	return "test-" + slug + "-" + hex.EncodeToString(suffix)
}

func deleteIndex(ctx context.Context, es *esindex.Client, name string) error {
	req, err := http.NewRequestWithContext(ctx, http.MethodDelete, strings.TrimRight(es.URL, "/")+"/"+name, nil)
	if err != nil {
		return err
	}
	if es.Username != "" {
		req.SetBasicAuth(es.Username, es.Password)
	}
	res, err := es.HTTP.Do(req)
	if err != nil {
		return err
	}
	defer res.Body.Close()
	if res.StatusCode/100 != 2 && res.StatusCode != http.StatusNotFound {
		return fmt.Errorf("delete index %s: %s", name, res.Status)
	}
	return nil
}
