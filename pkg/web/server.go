// Package web serves the Wide Door page and JSON API. It is the controller; ask is the use case.
package web

import (
	"context"
	"embed"
	"encoding/json"
	"fmt"
	"html/template"
	"log/slog"
	"net/http"
	"strings"
	"sync"
	"time"

	"go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/metric"
	"go.opentelemetry.io/otel/trace"

	v1 "github.com/auroq/cnh/pkg/api/v1"
	"github.com/auroq/cnh/pkg/ask"
)

//go:embed templates/*.html static/*
var assets embed.FS

// Asker is what the web layer needs from the use case.
type Asker interface {
	Ask(ctx context.Context, question, from string) (ask.Answer, error)
	Locate(ctx context.Context, asked string) (ask.Here, error)
	Ready(ctx context.Context) error
}

type Server struct {
	ask    Asker
	tmpl   *template.Template
	log    *slog.Logger
	tracer trace.TracerProvider
	meter  metric.MeterProvider

	once               sync.Once
	chipCounter        metric.Int64Counter
	chipLearnedCounter metric.Int64Counter
}

// Option adjusts a Server; nil tracing options mean the OpenTelemetry globals.
type Option func(*Server)

// WithTracerProvider is where request spans go; tests pass an in-memory one.
func WithTracerProvider(tp trace.TracerProvider) Option {
	return func(s *Server) { s.tracer = tp }
}

type page struct {
	Title       string
	Question    string
	From        string
	Answer      *ask.Answer
	Scanned     *ask.Here
	Error       string
	Kinds       []kind
	Prefs       string
	Chips       []chip
	Count       string
	ProfileOpen bool
	ProfileURL  string
	Status      string
	Forgot      bool
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

const (
	askTimeout   = 10 * time.Second
	readyTimeout = 3 * time.Second
	searchDown   = "The search is not answering right now. Ask staff at registration, and try again in a minute."
)

func New(a Asker, log *slog.Logger, opts ...Option) *Server {
	s := &Server{
		ask:    a,
		tmpl:   template.Must(template.New("").Funcs(template.FuncMap{"clock": clock, "location": readableLocation, "when": when, "distance": distance}).ParseFS(assets, "templates/*.html")),
		log:    log,
		tracer: otel.GetTracerProvider(),
	}
	for _, opt := range opts {
		opt(s)
	}
	return s
}

// Handler is the mux behind one span per request, named for the route pattern that matched;
// the probes are served without a span so they do not drown the real requests.
func (s *Server) Handler() http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("GET /{$}", s.home)
	mux.HandleFunc("GET /ask", s.home)
	mux.HandleFunc("GET /profile", s.home)
	mux.HandleFunc("POST /profile", s.profile)
	mux.HandleFunc("GET /api/ask", s.apiAsk)
	mux.HandleFunc("GET /healthz", func(w http.ResponseWriter, _ *http.Request) { fmt.Fprintln(w, "ok") })
	mux.HandleFunc("GET /readyz", s.ready)
	mux.Handle("GET /static/", http.FileServerFS(assets))
	return otelhttp.NewHandler(s.logged(mux), "widedoor", otelhttp.WithFilter(notProbe), otelhttp.WithTracerProvider(s.tracer))
}

func notProbe(r *http.Request) bool {
	return r.URL.Path != "/healthz" && r.URL.Path != "/readyz"
}

func (s *Server) home(w http.ResponseWriter, r *http.Request) {
	q := strings.TrimSpace(r.URL.Query().Get("q"))
	from := strings.TrimSpace(r.URL.Query().Get("from"))
	prefs := s.preferences(w, r)
	p := page{Title: "Wide Door", Question: q, From: from, Kinds: quickKinds}
	if q == "" && from != "" {
		ctx, cancel := context.WithTimeout(r.Context(), askTimeout)
		defer cancel()
		if here, err := s.ask.Locate(ctx, from); err != nil {
			s.log.ErrorContext(ctx, "locate", "err", err.Error())
		} else if here.Room != "" {
			p.Scanned = &here
			p.From = here.Room
		}
	}
	if q != "" {
		ctx, cancel := context.WithTimeout(r.Context(), askTimeout)
		defer cancel()
		s.countChips(ctx, prefs)
		a, err := s.askFor(ctx, q, from, prefs)
		if err != nil {
			s.log.ErrorContext(ctx, "ask", "err", err.Error())
			p.Error = searchDown
			p.Title = "Search not answering | Wide Door"
		} else {
			p.Answer = &a
			p.Title = a.Summary + " | Wide Door"
			if r.URL.Path != profilePath {
				prefs = s.learn(ctx, w, r, &p, prefs, a.Learned)
			}
		}
	}
	s.describeProfile(&p, r, prefs)
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	if err := s.tmpl.ExecuteTemplate(w, "home.html", p); err != nil {
		s.log.ErrorContext(r.Context(), "render", "err", err.Error())
	}
}

func (s *Server) apiAsk(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(r.Context(), askTimeout)
	defer cancel()
	prefs := s.preferences(w, r)
	if strings.TrimSpace(r.URL.Query().Get("q")) != "" {
		s.countChips(ctx, prefs)
	}
	a, err := s.askFor(ctx, r.URL.Query().Get("q"), r.URL.Query().Get("from"), prefs)
	w.Header().Set("Content-Type", "application/json")
	if err != nil {
		s.log.ErrorContext(ctx, "api ask", "err", err.Error())
		w.WriteHeader(http.StatusBadGateway)
		_ = json.NewEncoder(w).Encode(v1.Error{Error: "search unavailable"})
		return
	}
	_ = json.NewEncoder(w).Encode(toV1(a))
}

func (s *Server) ready(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(r.Context(), readyTimeout)
	defer cancel()
	if err := s.ask.Ready(ctx); err != nil {
		http.Error(w, "search unavailable", http.StatusServiceUnavailable)
		return
	}
	fmt.Fprintln(w, "ready")
}

func (s *Server) logged(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		next.ServeHTTP(w, r)
		if !notProbe(r) {
			return
		}
		route := routeOf(r.Pattern)
		trace.SpanFromContext(r.Context()).SetAttributes(attribute.String("http.route", route))
		s.log.InfoContext(r.Context(), "request", "method", r.Method, "route", route, "path", r.URL.Path, "q", r.URL.Query().Get("q"), "ms", time.Since(start).Milliseconds())
	})
}

func routeOf(pattern string) string {
	if i := strings.IndexByte(pattern, '/'); i >= 0 {
		return pattern[i:]
	}
	return pattern
}

var mountain = time.FixedZone("MST", -7*3600)

func clock(t time.Time) string {
	return t.In(mountain).Format("Mon 3:04 pm")
}

func hour(t time.Time) string {
	return t.In(mountain).Format("3:04 pm")
}

func readableLocation(location string) string {
	return strings.ReplaceAll(location, " | ", ", ")
}
