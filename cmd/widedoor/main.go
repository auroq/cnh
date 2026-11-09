package main

import (
	"context"
	"embed"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"html/template"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/auroq/cnh/pkg/ask"
	"github.com/auroq/cnh/pkg/esindex"
)

//go:embed templates/*.html static/*
var assets embed.FS

type server struct {
	ask  *ask.Service
	tmpl *template.Template
	log  *slog.Logger
}

type page struct {
	Title    string
	Question string
	From     string
	Answer   *ask.Answer
	Error    string
	Kinds    []kind
}

type kind struct{ Query, Label string }

var quickKinds = []kind{
	{"nearest all-gender restroom", "All-gender restroom"},
	{"nearest quiet room", "Quiet room"},
	{"nearest elevator", "Elevator"},
	{"step-free entrance", "Step-free entrance"},
	{"prayer room", "Prayer room"},
	{"nursing room", "Nursing room"},
	{"coat and bag check", "Coat check"},
	{"registration and help desk", "Help desk"},
}

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
	es := &esindex.Client{
		URL:      env("ES_URL", "http://localhost:9200"),
		Username: env("ES_USERNAME", "elastic"),
		Password: os.Getenv("ES_PASSWORD"),
		HTTP:     &http.Client{Timeout: 15 * time.Second},
	}
	s := &server{
		ask:  &ask.Service{ES: es, PlacesIndex: env("PLACES_INDEX", "places"), SessionsIndex: env("ES_INDEX", "sessions")},
		tmpl: template.Must(template.New("").Funcs(template.FuncMap{"clock": clock, "location": readableLocation}).ParseFS(assets, "templates/*.html")),
		log:  log,
	}

	mux := http.NewServeMux()
	mux.HandleFunc("GET /{$}", s.home)
	mux.HandleFunc("GET /ask", s.home)
	mux.HandleFunc("GET /api/ask", s.apiAsk)
	mux.HandleFunc("GET /healthz", func(w http.ResponseWriter, _ *http.Request) { fmt.Fprintln(w, "ok") })
	mux.HandleFunc("GET /readyz", s.ready)
	mux.Handle("GET /static/", http.FileServerFS(assets))

	srv := &http.Server{
		Addr:              env("LISTEN", ":8080"),
		Handler:           s.logged(mux),
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

func (s *server) home(w http.ResponseWriter, r *http.Request) {
	q := strings.TrimSpace(r.URL.Query().Get("q"))
	from := strings.TrimSpace(r.URL.Query().Get("from"))
	p := page{Title: "Wide Door", Question: q, From: from, Kinds: quickKinds}
	if q != "" {
		ctx, cancel := context.WithTimeout(r.Context(), 10*time.Second)
		defer cancel()
		a, err := s.ask.Ask(ctx, q, from)
		if err != nil {
			s.log.Error("ask", "err", err.Error())
			p.Error = "The search is not answering right now. Ask staff at registration, and try again in a minute."
			p.Title = "Search not answering | Wide Door"
		} else {
			p.Answer = &a
			p.Title = a.Summary + " | Wide Door"
		}
	}
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	if err := s.tmpl.ExecuteTemplate(w, "home.html", p); err != nil {
		s.log.Error("render", "err", err.Error())
	}
}

func (s *server) apiAsk(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(r.Context(), 10*time.Second)
	defer cancel()
	a, err := s.ask.Ask(ctx, r.URL.Query().Get("q"), r.URL.Query().Get("from"))
	w.Header().Set("Content-Type", "application/json")
	if err != nil {
		s.log.Error("api ask", "err", err.Error())
		w.WriteHeader(http.StatusBadGateway)
		_ = json.NewEncoder(w).Encode(map[string]string{"error": "search unavailable"})
		return
	}
	_ = json.NewEncoder(w).Encode(a)
}

func (s *server) ready(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(r.Context(), 3*time.Second)
	defer cancel()
	if _, err := s.ask.Locate(ctx, "Grand Ballroom J"); err != nil {
		http.Error(w, "elasticsearch unavailable", http.StatusServiceUnavailable)
		return
	}
	fmt.Fprintln(w, "ready")
}

func (s *server) logged(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		next.ServeHTTP(w, r)
		if r.URL.Path == "/healthz" || r.URL.Path == "/readyz" {
			return
		}
		s.log.Info("request", "method", r.Method, "path", r.URL.Path, "q", r.URL.Query().Get("q"), "ms", time.Since(start).Milliseconds())
	})
}

func clock(t time.Time) string {
	return t.In(time.FixedZone("MST", -7*3600)).Format("Mon 3:04 pm")
}

func readableLocation(location string) string {
	return strings.ReplaceAll(location, " | ", ", ")
}

func env(k, def string) string {
	if v := strings.TrimSpace(os.Getenv(k)); v != "" {
		return v
	}
	return def
}
