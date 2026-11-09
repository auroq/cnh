package helpers

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"strings"
	"testing"
	"time"
)

func BaseURL(t *testing.T) string {
	t.Helper()
	u := strings.TrimRight(strings.TrimSpace(os.Getenv("WIDEDOOR_URL")), "/")
	if u == "" {
		t.Skip("WIDEDOOR_URL is unset; set it to the Wide Door base URL, such as https://widedoor.cnh.auroq.com, to run acceptance tests")
	}
	return u
}

func GetJSON(t *testing.T, path string, query url.Values, out any) {
	t.Helper()
	status, body := Get(t, path, query, "application/json")
	if status != http.StatusOK {
		t.Fatalf("GET %s: %d: %s", path, status, body)
	}
	if err := json.Unmarshal([]byte(body), out); err != nil {
		t.Fatalf("GET %s: decode: %v: %s", path, err, body)
	}
}

func Get(t *testing.T, path string, query url.Values, accept string) (int, string) {
	t.Helper()
	u := BaseURL(t) + path
	if len(query) > 0 {
		u += "?" + query.Encode()
	}
	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
	defer cancel()
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, u, nil)
	if err != nil {
		t.Fatalf("GET %s: %v", u, err)
	}
	req.Header.Set("Accept", accept)
	res, err := (&http.Client{Timeout: 20 * time.Second}).Do(req)
	if err != nil {
		t.Fatalf("GET %s: %v", u, err)
	}
	defer res.Body.Close()
	body, err := io.ReadAll(res.Body)
	if err != nil {
		t.Fatalf("GET %s: read body: %v", u, err)
	}
	return res.StatusCode, string(body)
}

func Query(pairs ...string) url.Values {
	if len(pairs)%2 != 0 {
		panic(fmt.Sprintf("Query: odd number of arguments: %d", len(pairs)))
	}
	q := url.Values{}
	for i := 0; i < len(pairs); i += 2 {
		if pairs[i+1] != "" {
			q.Set(pairs[i], pairs[i+1])
		}
	}
	return q
}
