package sessionize

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"sync"
)

const DefaultIndexURL = "https://sessionize.com/api/v2/uo948i5p/view/markdown"

type Client struct {
	HTTP        *http.Client
	IndexURL    string
	Concurrency int
}

func (c *Client) client() *http.Client {
	if c.HTTP != nil {
		return c.HTTP
	}
	return http.DefaultClient
}

func (c *Client) get(ctx context.Context, url string) (io.ReadCloser, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("User-Agent", "wide-door-ingest/0.1 (+https://github.com/auroq/cnh)")
	res, err := c.client().Do(req)
	if err != nil {
		return nil, err
	}
	if res.StatusCode != http.StatusOK {
		res.Body.Close()
		return nil, fmt.Errorf("GET %s: %s", url, res.Status)
	}
	return res.Body, nil
}

// Index fetches and parses the whole-event markdown.
func (c *Client) Index(ctx context.Context) ([]Session, error) {
	url := c.IndexURL
	if url == "" {
		url = DefaultIndexURL
	}
	body, err := c.get(ctx, url)
	if err != nil {
		return nil, err
	}
	defer body.Close()
	return ParseIndex(body)
}

// Describe fetches each session's page and fills in Description, in place.
// It returns the number of sessions that could not be fetched.
func (c *Client) Describe(ctx context.Context, sessions []Session, report func(id string, err error)) int {
	n := c.Concurrency
	if n <= 0 {
		n = 8
	}
	var (
		wg     sync.WaitGroup
		sem    = make(chan struct{}, n)
		mu     sync.Mutex
		failed int
	)
	for i := range sessions {
		wg.Add(1)
		sem <- struct{}{}
		go func(s *Session) {
			defer wg.Done()
			defer func() { <-sem }()
			err := c.describe(ctx, s)
			if err != nil {
				mu.Lock()
				failed++
				mu.Unlock()
			}
			if report != nil {
				report(s.ID, err)
			}
		}(&sessions[i])
	}
	wg.Wait()
	return failed
}

func (c *Client) describe(ctx context.Context, s *Session) error {
	body, err := c.get(ctx, s.URL)
	if err != nil {
		return err
	}
	defer body.Close()
	d, err := ParseDetail(body)
	if err != nil {
		return err
	}
	s.Description = d.Description
	return nil
}
