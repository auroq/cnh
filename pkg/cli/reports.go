package cli

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	"github.com/auroq/cnh/pkg/config"
	"github.com/auroq/cnh/pkg/esindex"
)

// NewReports builds the reports command for volunteers: list what attendees said, hide what should not count.
func NewReports(v *viper.Viper, log *slog.Logger) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "reports",
		Short: "List and hide attendee reports on places",
		Long: `Volunteers' view of "Were you here?" reports. list prints the recent ones, newest first, with the id
to pass to hide. A hidden report stays in the index but stops counting towards a place's status.
The token column is never printed.`,
	}
	var since time.Duration
	list := &cobra.Command{
		Use:   "list",
		Short: "Print recent reports, newest first",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, _ []string) error {
			cfg, err := config.Load(v)
			if err != nil {
				return err
			}
			ctx, cancel := context.WithTimeout(cmd.Context(), cfg.Elasticsearch().Timeout())
			defer cancel()
			reps, err := newReports(cfg.Elasticsearch()).Recent(ctx, time.Now().Add(-since))
			if err != nil {
				return err
			}
			fmt.Fprintf(cmd.OutOrStdout(), "%-32s  %-20s  %-24s  %-16s  %-6s  %s\n", "ID", "AT", "PLACE", "VERDICT", "HIDDEN", "NOTE")
			for _, r := range reps {
				fmt.Fprintf(cmd.OutOrStdout(), "%-32s  %-20s  %-24s  %-16s  %-6t  %s\n", r.ID, r.At.UTC().Format(time.RFC3339), r.PlaceID, r.Verdict, r.Hidden, r.Note)
			}
			return nil
		},
	}
	list.Flags().DurationVar(&since, "since", 24*time.Hour, "how far back to list")
	hide := &cobra.Command{
		Use:   "hide ID",
		Short: "Hide a report so it stops counting",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			cfg, err := config.Load(v)
			if err != nil {
				return err
			}
			ctx, cancel := context.WithTimeout(cmd.Context(), cfg.Elasticsearch().Timeout())
			defer cancel()
			if err := newReports(cfg.Elasticsearch()).Hide(ctx, args[0]); err != nil {
				return err
			}
			log.Info("report hidden", "id", args[0])
			return nil
		},
	}
	cmd.AddCommand(list, hide)
	return cmd
}

func newReports(es config.Elasticsearch) *esindex.Reports {
	return &esindex.Reports{Client: newClient(es), ReportsIndex: es.ReportsIndex(), PlacesIndex: es.PlacesIndex()}
}
