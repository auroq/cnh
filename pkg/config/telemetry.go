package config

import "github.com/spf13/viper"

const (
	KeyTelemetryOTLPEndpoint = "telemetry.otlp_endpoint"
	KeyTelemetryServiceName  = "telemetry.service_name"
)

var telemetryDefaults = map[string]any{
	KeyTelemetryOTLPEndpoint: "",
	KeyTelemetryServiceName:  "widedoor",
}

// Telemetry is what the OpenTelemetry setup needs.
type Telemetry interface {
	OTLPEndpoint() string
	ServiceName() string
}

func (c *Configuration) Telemetry() Telemetry { return telemetry{c.v} }

type telemetry struct{ v *viper.Viper }

func (t telemetry) OTLPEndpoint() string { return t.v.GetString(KeyTelemetryOTLPEndpoint) }
func (t telemetry) ServiceName() string  { return t.v.GetString(KeyTelemetryServiceName) }
