package cli

import (
	"context"
	"errors"
	"log/slog"
	"net/http"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	"github.com/auroq/cnh/pkg/config"
	"github.com/auroq/cnh/pkg/esindex"
	"github.com/auroq/cnh/pkg/ingest"
	"github.com/auroq/cnh/pkg/sessionize"
	"github.com/auroq/cnh/pkg/venue"
)

// ErrDocumentsRejected means the run finished but the store rejected some documents; the job exits 2.
var ErrDocumentsRejected = errors.New("some documents were rejected by the store")

// NewIngest builds the ingest command: schedule and venue places into the search store.
func NewIngest(v *viper.Viper, log *slog.Logger) *cobra.Command {
	return &cobra.Command{
		Use:   "ingest",
		Short: "Fetch the KubeCon schedule and the venue places and index them",
		Long: `Fetches the KubeCon schedule from Sessionize and indexes it, then indexes the curated
venue places and the rooms the schedule names. Safe to re-run: documents are keyed by id.`,
		Args: cobra.NoArgs,
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
			in := cfg.Ingest()
			es := cfg.Elasticsearch()
			ctx, cancel := context.WithTimeout(cmd.Context(), in.Timeout())
			defer cancel()
			runner := &ingest.Runner{
				Source:      &sessionize.Client{HTTP: &http.Client{Timeout: in.FetchTimeout()}, IndexURL: in.SessionizeURL(), Concurrency: in.Concurrency()},
				Places:      venue.File(in.PlacesFile()),
				Indexer:     &esindex.Indexer{Client: newClient(es), SessionsIndex: es.SessionsIndex(), PlacesIndex: es.PlacesIndex(), Batch: in.Batch()},
				SkipDetails: in.SkipDetails(),
				Log:         log,
			}
			report, err := runner.Run(ctx)
			if err != nil {
				return err
			}
			if report.Failed() {
				return ErrDocumentsRejected
			}
			return nil
		},
	}
}
