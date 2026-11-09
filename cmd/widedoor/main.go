package main

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/auroq/cnh/pkg/ask"
	"github.com/auroq/cnh/pkg/esindex"
	"github.com/auroq/cnh/pkg/web"
)

func main() {
	fs := flag.NewFlagSet("widedoor", flag.ExitOnError)
	fs.Usage = func() {
		fmt.Fprint(fs.Output(), `usage: widedoor [flags]

Serves the Wide Door web app and JSON API.

Environment:
  ES_URL          Elasticsearch base URL (default http://localhost:9200)
  ES_USERNAME     basic auth user (default elastic)
  ES_PASSWORD     basic auth password
  PLACES_INDEX    places index (default places)
  ES_INDEX        sessions index (default sessions)
  LISTEN          address to listen on (default :8080)

Flags:
`)
		fs.PrintDefaults()
	}
	_ = fs.Parse(os.Args[1:])

	log := slog.New(slog.NewJSONHandler(os.Stdout, nil))
	store := &esindex.Store{
		Client: &esindex.Client{
			URL:      env("ES_URL", "http://localhost:9200"),
			Username: env("ES_USERNAME", "elastic"),
			Password: os.Getenv("ES_PASSWORD"),
			HTTP:     &http.Client{Timeout: 15 * time.Second},
		},
		PlacesIndex:   env("PLACES_INDEX", esindex.DefaultPlacesIndex),
		SessionsIndex: env("ES_INDEX", esindex.DefaultSessionsIndex),
	}
	srv := &http.Server{
		Addr:              env("LISTEN", ":8080"),
		Handler:           web.New(&ask.Service{Store: store}, log).Handler(),
		ReadHeaderTimeout: 5 * time.Second,
	}

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()
	go func() {
		<-ctx.Done()
		shutdown, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		_ = srv.Shutdown(shutdown)
	}()
	log.Info("listening", "addr", srv.Addr)
	if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
		log.Error("serve", "err", err.Error())
		os.Exit(1)
	}
}

func env(k, def string) string {
	if v := strings.TrimSpace(os.Getenv(k)); v != "" {
		return v
	}
	return def
}
