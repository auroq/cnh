// Package app is the entry point: it builds the command line and turns its result into an exit code.
package app

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"os"

	"github.com/auroq/cnh/pkg/cli"
	"github.com/auroq/cnh/pkg/config"
)

// Main runs widedoor with os.Args and returns the exit code: 0 ok, 1 error, 2 documents rejected.
func Main() int {
	log := slog.New(slog.NewJSONHandler(os.Stdout, nil))
	v := config.New()
	root := cli.NewRoot(v, cli.NewServe(v, log), cli.NewIngest(v, log), cli.NewQR(v, log))
	err := root.ExecuteContext(context.Background())
	switch {
	case err == nil:
		return 0
	case errors.Is(err, cli.ErrDocumentsRejected):
		log.Error(err.Error())
		return 2
	default:
		fmt.Fprintln(os.Stderr, "widedoor:", err)
		return 1
	}
}
