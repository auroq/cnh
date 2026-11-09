package cli

import (
	"context"
	"errors"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	"github.com/auroq/cnh/pkg/ask"
	"github.com/auroq/cnh/pkg/config"
	"github.com/auroq/cnh/pkg/esindex"
	"github.com/auroq/cnh/pkg/telemetry"
	"github.com/auroq/cnh/pkg/web"
)

const flushTimeout = 10 * time.Second

// NewServe builds the serve command: the web app and JSON API in front of the search store.
func NewServe(v *viper.Viper, log *slog.Logger) *cobra.Command {
	return &cobra.Command{
		Use:   "serve",
		Short: "Serve the Wide Door web app and JSON API",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, _ []string) error {
			cfg, err := config.Load(v)
			if err != nil {
				return err
			}
			log, flush, err := observe(cmd.Context(), cfg.Telemetry(), log)
			if err != nil {
				return err
			}
			defer flush()
			store := newStore(cfg.Elasticsearch())
			handler := web.New(&ask.Service{Store: store}, log).Handler()
			return serve(cmd.Context(), cfg.Server(), handler, log)
		},
	}
}

// observe installs the OpenTelemetry providers for cfg and returns the logger that reaches them,
// with a flush to call before the process exits.
func observe(ctx context.Context, cfg config.Telemetry, log *slog.Logger) (*slog.Logger, func(), error) {
	providers, err := telemetry.Setup(ctx, cfg)
	if err != nil {
		return nil, nil, err
	}
	providers.Install()
	log = slog.New(telemetry.Handler(log.Handler()))
	endpoint := cfg.OTLPEndpoint()
	if endpoint == "" {
		endpoint = "disabled"
	}
	log.Info("telemetry", "otlp_endpoint", endpoint, "service_name", cfg.ServiceName(), "service_version", telemetry.Version)
	flush := func() {
		ctx, cancel := context.WithTimeout(context.Background(), flushTimeout)
		defer cancel()
		if err := providers.Shutdown(ctx); err != nil {
			log.Error("telemetry shutdown", "err", err.Error())
		}
	}
	return log, flush, nil
}

func newClient(es config.Elasticsearch) *esindex.Client {
	return newClientWithTimeout(es, es.Timeout())
}

func newClientWithTimeout(es config.Elasticsearch, timeout time.Duration) *esindex.Client {
	return &esindex.Client{
		URL:      es.URL(),
		Username: es.Username(),
		Password: es.Password(),
		HTTP:     &http.Client{Timeout: timeout, Transport: telemetry.Transport("elasticsearch", http.DefaultTransport)},
	}
}

func newStore(es config.Elasticsearch) *esindex.Store {
	return &esindex.Store{
		Client:        newClient(es),
		PlacesIndex:   es.PlacesIndex(),
		SessionsIndex: es.SessionsIndex(),
	}
}

func serve(ctx context.Context, cfg config.Server, handler http.Handler, log *slog.Logger) error {
	srv := &http.Server{
		Addr:              cfg.Listen(),
		Handler:           handler,
		ReadHeaderTimeout: 5 * time.Second,
	}
	ctx, stop := signal.NotifyContext(ctx, os.Interrupt, syscall.SIGTERM)
	defer stop()
	go func() {
		<-ctx.Done()
		shutdown, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		_ = srv.Shutdown(shutdown)
	}()
	log.Info("listening", "addr", srv.Addr)
	if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
		return err
	}
	return nil
}
