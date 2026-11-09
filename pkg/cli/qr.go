package cli

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	"github.com/auroq/cnh/pkg/config"
	"github.com/auroq/cnh/pkg/qr"
	"github.com/auroq/cnh/pkg/sessionize"
	"github.com/auroq/cnh/pkg/venue"
)

// NewQR builds the qr command: one code per room and landmark, and a sheet to print and tape beside the door signs.
func NewQR(v *viper.Viper, log *slog.Logger) *cobra.Command {
	var out string
	var sheet, rooms bool
	cmd := &cobra.Command{
		Use:   "qr",
		Short: "Render a QR code per room and landmark that opens Wide Door with the location filled in",
		Long: `Writes <place id>.svg and <place id>.png for every room, entrance, landmark, quiet room,
prayer room, nursing room, first-aid point, info desk and coat check, each encoding
<qr.base_url>/?from=<place id>. With --sheet it also writes sheet.html, one printable page per place
with the name in large type, the building and level, and the URL under the code.

Curated places come from ingest.places_file; the rooms come from the Sessionize schedule unless --rooms=false.`,
		Args: cobra.NoArgs,
		RunE: func(cmd *cobra.Command, _ []string) error {
			cfg, err := config.Load(v)
			if err != nil {
				return err
			}
			in := cfg.Ingest()
			places, err := venue.File(in.PlacesFile()).Places()
			if err != nil {
				return fmt.Errorf("load places: %w", err)
			}
			if rooms {
				ctx, cancel := context.WithTimeout(cmd.Context(), in.FetchTimeout())
				defer cancel()
				sessions, err := (&sessionize.Client{HTTP: &http.Client{Timeout: in.FetchTimeout()}, IndexURL: in.SessionizeURL()}).Index(ctx)
				if err != nil {
					return fmt.Errorf("fetch schedule for rooms: %w", err)
				}
				locations := make([]string, 0, len(sessions))
				for _, s := range sessions {
					locations = append(locations, s.Location)
				}
				places = append(places, venue.RoomsFromLocations(locations)...)
			}
			written, err := qr.Codes{BaseURL: cfg.QR().BaseURL()}.Write(out, places, sheet)
			if err != nil {
				return err
			}
			log.Info("qr codes written", "dir", out, "codes", len(written), "sheet", sheet)
			return nil
		},
	}
	cmd.Flags().StringVar(&out, "out", "qr", "directory to write the codes into")
	cmd.Flags().BoolVar(&sheet, "sheet", false, "also write sheet.html, one printable page per place")
	cmd.Flags().BoolVar(&rooms, "rooms", true, "fetch the schedule and render a code for every room it names")
	return cmd
}
