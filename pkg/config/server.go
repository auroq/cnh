package config

import "github.com/spf13/viper"

const KeyServerListen = "server.listen"

var serverDefaults = map[string]any{
	KeyServerListen: ":8080",
}

// Server is what the HTTP server needs.
type Server interface {
	Listen() string
}

func (c *Configuration) Server() Server { return server{c.v} }

type server struct{ v *viper.Viper }

func (s server) Listen() string { return s.v.GetString(KeyServerListen) }
