package esindex

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
)

type Client struct {
	URL      string
	Username string
	Password string
	HTTP     *http.Client
}

type Doc struct {
	ID   string
	Body any
}

func (c *Client) do(ctx context.Context, method, path string, body io.Reader, contentType string) ([]byte, int, error) {
	req, err := http.NewRequestWithContext(ctx, method, strings.TrimRight(c.URL, "/")+path, body)
	if err != nil {
		return nil, 0, err
	}
	if contentType != "" {
		req.Header.Set("Content-Type", contentType)
	}
	if c.Username != "" {
		req.SetBasicAuth(c.Username, c.Password)
	}
	hc := c.HTTP
	if hc == nil {
		hc = http.DefaultClient
	}
	res, err := hc.Do(req)
	if err != nil {
		return nil, 0, err
	}
	defer res.Body.Close()
	out, err := io.ReadAll(res.Body)
	return out, res.StatusCode, err
}

// EnsureIndex creates the index with the given body if it does not exist.
func (c *Client) EnsureIndex(ctx context.Context, name string, body any) error {
	_, code, err := c.do(ctx, http.MethodHead, "/"+name, nil, "")
	if err != nil {
		return err
	}
	if code == http.StatusOK {
		return nil
	}
	b, err := json.Marshal(body)
	if err != nil {
		return err
	}
	out, code, err := c.do(ctx, http.MethodPut, "/"+name, bytes.NewReader(b), "application/json")
	if err != nil {
		return err
	}
	if code/100 != 2 {
		return fmt.Errorf("create index %s: %d %s", name, code, out)
	}
	return nil
}

// BulkBody renders the NDJSON body for a bulk index request.
func BulkBody(index string, docs []Doc) ([]byte, error) {
	var buf bytes.Buffer
	enc := json.NewEncoder(&buf)
	for _, d := range docs {
		if err := enc.Encode(map[string]any{"index": map[string]string{"_index": index, "_id": d.ID}}); err != nil {
			return nil, err
		}
		if err := enc.Encode(d.Body); err != nil {
			return nil, err
		}
	}
	return buf.Bytes(), nil
}

type bulkResponse struct {
	Errors bool `json:"errors"`
	Items  []map[string]struct {
		Status int `json:"status"`
		Error  *struct {
			Type   string `json:"type"`
			Reason string `json:"reason"`
		} `json:"error"`
	} `json:"items"`
}

// Bulk indexes docs in batches and returns the number of documents that failed.
func (c *Client) Bulk(ctx context.Context, index string, docs []Doc, batch int) (int, error) {
	if batch <= 0 {
		batch = 100
	}
	failed := 0
	for start := 0; start < len(docs); start += batch {
		end := min(start+batch, len(docs))
		body, err := BulkBody(index, docs[start:end])
		if err != nil {
			return failed, err
		}
		out, code, err := c.do(ctx, http.MethodPost, "/_bulk?refresh=wait_for", bytes.NewReader(body), "application/x-ndjson")
		if err != nil {
			return failed, err
		}
		if code/100 != 2 {
			return failed, fmt.Errorf("bulk: %d %s", code, out)
		}
		var res bulkResponse
		if err := json.Unmarshal(out, &res); err != nil {
			return failed, err
		}
		for _, item := range res.Items {
			for _, r := range item {
				if r.Error != nil {
					failed++
				}
			}
		}
	}
	return failed, nil
}

// Count returns the document count of an index.
func (c *Client) Count(ctx context.Context, index string) (int, error) {
	out, code, err := c.do(ctx, http.MethodGet, "/"+index+"/_count", nil, "")
	if err != nil {
		return 0, err
	}
	if code != http.StatusOK {
		return 0, fmt.Errorf("count %s: %d %s", index, code, out)
	}
	var res struct {
		Count int `json:"count"`
	}
	return res.Count, json.Unmarshal(out, &res)
}

// Search runs a search body against an index and returns the raw response.
func (c *Client) Search(ctx context.Context, index string, body any) (json.RawMessage, error) {
	b, err := json.Marshal(body)
	if err != nil {
		return nil, err
	}
	out, code, err := c.do(ctx, http.MethodPost, "/"+index+"/_search", bytes.NewReader(b), "application/json")
	if err != nil {
		return nil, err
	}
	if code != http.StatusOK {
		return nil, fmt.Errorf("search %s: %d %s", index, code, out)
	}
	return out, nil
}
