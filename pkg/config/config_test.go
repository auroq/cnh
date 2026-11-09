package config

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/spf13/pflag"
)

func load(t *testing.T, args []string, env map[string]string, file string) *Configuration {
	t.Helper()
	for k, val := range env {
		t.Setenv(k, val)
	}
	v := New()
	flags := pflag.NewFlagSet("test", pflag.ContinueOnError)
	BindFlags(v, flags)
	if file != "" {
		path := filepath.Join(t.TempDir(), "config.yaml")
		if err := os.WriteFile(path, []byte(file), 0o600); err != nil {
			t.Fatal(err)
		}
		args = append(args, "--config", path)
	}
	if err := flags.Parse(args); err != nil {
		t.Fatal(err)
	}
	cfg, err := Load(v)
	if err != nil {
		t.Fatal(err)
	}
	return cfg
}

func TestPrecedence(t *testing.T) {
	file := "elasticsearch:\n  url: http://from-file:9200\n"
	cases := []struct {
		name string
		args []string
		env  map[string]string
		file string
		want string
	}{
		{"nothing is set", nil, nil, "", "http://localhost:9200"},
		{"only the file sets it", nil, nil, file, "http://from-file:9200"},
		{"the environment and the file set it", nil, map[string]string{"WIDEDOOR_ELASTICSEARCH_URL": "http://from-env:9200"}, file, "http://from-env:9200"},
		{"a flag, the environment and the file set it", []string{"--elasticsearch-url", "http://from-flag:9200"}, map[string]string{"WIDEDOOR_ELASTICSEARCH_URL": "http://from-env:9200"}, file, "http://from-flag:9200"},
	}
	for _, c := range cases {
		t.Run("when "+c.name, func(t *testing.T) {
			cfg := load(t, c.args, c.env, c.file)
			t.Run("it should take flag over env over file over default", func(t *testing.T) {
				if got := cfg.Elasticsearch().URL(); got != c.want {
					t.Fatalf("got %q, want %q", got, c.want)
				}
			})
		})
	}
}

func TestSecrets(t *testing.T) {
	t.Run("when flags are bound", func(t *testing.T) {
		flags := pflag.NewFlagSet("test", pflag.ContinueOnError)
		BindFlags(New(), flags)
		t.Run("it should not offer the Elasticsearch password as a flag", func(t *testing.T) {
			if flags.Lookup("elasticsearch-password") != nil {
				t.Fatal("password is a flag")
			}
		})
	})

	t.Run("when the password is in the environment", func(t *testing.T) {
		cfg := load(t, nil, map[string]string{"WIDEDOOR_ELASTICSEARCH_PASSWORD": "hunter2"}, "")
		t.Run("it should be read", func(t *testing.T) {
			if cfg.Elasticsearch().Password() != "hunter2" {
				t.Fatal("password not read from env")
			}
		})
	})
}

func TestTypedAccessors(t *testing.T) {
	file := "server:\n  listen: \":9090\"\ningest:\n  concurrency: 3\n  skip_details: true\n  timeout: 2m\ntelemetry:\n  otlp_endpoint: http://collector:4318\n"
	cfg := load(t, nil, nil, file)
	t.Run("when a file sets several areas", func(t *testing.T) {
		checks := []struct {
			name string
			got  any
			want any
		}{
			{"server listen", cfg.Server().Listen(), ":9090"},
			{"ingest concurrency", cfg.Ingest().Concurrency(), 3},
			{"ingest skip details", cfg.Ingest().SkipDetails(), true},
			{"ingest timeout", cfg.Ingest().Timeout().String(), "2m0s"},
			{"ingest elasticsearch timeout default", cfg.Ingest().ESTimeout().String(), "5m0s"},
			{"telemetry endpoint", cfg.Telemetry().OTLPEndpoint(), "http://collector:4318"},
			{"telemetry service name default", cfg.Telemetry().ServiceName(), "widedoor"},
		}
		for _, c := range checks {
			t.Run("it should expose "+c.name+" through its interface", func(t *testing.T) {
				if c.got != c.want {
					t.Fatalf("got %v, want %v", c.got, c.want)
				}
			})
		}
	})
}

func TestEnvName(t *testing.T) {
	cases := map[string]string{
		KeyElasticsearchURL:  "WIDEDOOR_ELASTICSEARCH_URL",
		KeyIngestSkipDetails: "WIDEDOOR_INGEST_SKIP_DETAILS",
	}
	for key, want := range cases {
		t.Run("when the key is "+key, func(t *testing.T) {
			t.Run("it should prefix and upper-case with underscores", func(t *testing.T) {
				if got := EnvName(key); got != want {
					t.Fatalf("got %s", got)
				}
			})
		})
	}
}
