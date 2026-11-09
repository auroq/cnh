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
	"github.com/auroq/cnh/pkg/web"
)

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
			store := newStore(cfg.Elasticsearch())
			handler := web.New(&ask.Service{Store: store}, log).Handler()
			return serve(cmd.Context(), cfg.Server(), handler, log)
		},
	}
}

func newClient(es config.Elasticsearch) *esindex.Client {
	return &esindex.Client{
		URL:      es.URL(),
		Username: es.Username(),
		Password: es.Password(),
		HTTP:     &http.Client{Timeout: es.Timeout()},
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
