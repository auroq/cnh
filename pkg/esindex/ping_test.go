package esindex

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestPing(t *testing.T) {
	cases := []struct {
		name    string
		status  int
		wantErr bool
	}{
		{"the cluster answers", http.StatusOK, false},
		{"the cluster is unavailable", http.StatusServiceUnavailable, true},
	}
	for _, c := range cases {
		t.Run("when "+c.name, func(t *testing.T) {
			var method, path string
			srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				method, path = r.Method, r.URL.Path
				w.WriteHeader(c.status)
			}))
			t.Cleanup(srv.Close)
			err := (&Client{URL: srv.URL, HTTP: srv.Client()}).Ping(context.Background())
			t.Run("it should ask the cluster root with a HEAD", func(t *testing.T) {
				if method != http.MethodHead || path != "/" {
					t.Fatalf("%s %s", method, path)
				}
			})
			t.Run("it should report the outcome as an error or not", func(t *testing.T) {
				if (err != nil) != c.wantErr {
					t.Fatalf("err %v, want error %v", err, c.wantErr)
				}
			})
		})
	}
}
