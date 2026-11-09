// Package cli is the command line: it reads configuration, constructs adapters and hands them to the use cases.
package cli

import (
	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	"github.com/auroq/cnh/pkg/config"
)

var groups = []struct {
	group    cobra.Group
	commands []string
}{
	{cobra.Group{ID: "run", Title: "Run:"}, []string{"serve", "ingest"}},
	{cobra.Group{ID: "print", Title: "Print:"}, []string{"qr"}},
	{cobra.Group{ID: "volunteer", Title: "Volunteer:"}, []string{"reports"}},
}

// NewRoot builds the widedoor command with its subcommands grouped for --help.
func NewRoot(v *viper.Viper, subcommands ...*cobra.Command) *cobra.Command {
	root := &cobra.Command{
		Use:   "widedoor",
		Short: "Ask the venue where the nearest step-free route, elevator, quiet room, all-gender restroom or captioned session is",
		Long: `Wide Door answers "where is the nearest X from here" for conference attendees,
in a form that works with a screen reader, a keyboard and large text.

Settings come from flags, then WIDEDOOR_* environment variables, then config.yaml
(beside the binary, in /etc/widedoor, or in the working directory, or --config), then defaults.
The Elasticsearch password is read only from WIDEDOOR_ELASTICSEARCH_PASSWORD or the config file.`,
		SilenceUsage:  true,
		SilenceErrors: true,
	}
	root.CompletionOptions.HiddenDefaultCmd = true
	config.BindFlags(v, root.PersistentFlags())
	byName := map[string]*cobra.Command{}
	for _, sub := range subcommands {
		byName[sub.Name()] = sub
	}
	for _, g := range groups {
		root.AddGroup(&g.group)
		for _, name := range g.commands {
			if sub, ok := byName[name]; ok {
				sub.GroupID = g.group.ID
			}
		}
	}
	root.AddCommand(subcommands...)
	return root
}
