// Package tracing implements a common tracing initialization for GD2 and CLI
package tracing

import (
	"github.com/gluster/glusterd2/glusterd2/gdctx"

	log "github.com/sirupsen/logrus"
	flag "github.com/spf13/pflag"
	config "github.com/spf13/viper"
	"go.opencensus.io/exporter/jaeger"
	"go.opencensus.io/trace"
)

// Commandline options for exporter endpoints
const (
	jaegerEndpointOpt      = "jaeger-endpoint"
	jaegerAgentEndpointOpt = "jaeger-agent-endpoint"
)

// InitFlags initializes the command line options for GD2 tracing endpoints
func InitFlags() {
	flag.String(jaegerEndpointOpt, "", "Jaeger collector endpoint that accepts spans from Jaeger agent.")
	flag.String(jaegerAgentEndpointOpt, "", "Jaeger agent endpoint that the Jaeger client sends spans to.")
}

// InitJaegerExporter initializes the jaeger exporter as the tracing endpoint
// This should be called early when a process starts.
// This creates and returns an exporter if successful.
// Otherwise, a warning is logged and a 'nil' exporter is returned.
func InitJaegerExporter() *jaeger.Exporter {
	// Get the Jaeger endpoints
	jaegerEndpoint := config.GetString(jaegerEndpointOpt)
	jaegerAgentEndpoint := config.GetString(jaegerAgentEndpointOpt)

	// Return nil exporter if either endpoints are not specified
	if jaegerEndpoint == "" || jaegerAgentEndpoint == "" {
		log.WithFields(log.Fields{
			"jaegerEndpoint":      jaegerEndpoint,
			"jaegerAgentEndpoint": jaegerAgentEndpoint,
		}).Warning("tracing: One or more Jaeger endpoints not specified")
		return nil
	}

	// Create the Opencensus Jaeger exporter
	exporter, err := jaeger.NewExporter(jaeger.Options{
		Endpoint:      jaegerEndpoint,
		AgentEndpoint: jaegerAgentEndpoint,
		ServiceName:   gdctx.HostName,
	})

	if err != nil {
		log.WithError(err).WithFields(log.Fields{
			"jaegerEndpoint":      jaegerEndpoint,
			"jaegerAgentEndpoint": jaegerAgentEndpoint,
		}).Warning("tracing: Unable to create opencensus jaeger exporter")
		return nil
	}

	// Register the Jaeger exporter
	// Register the passed exporter using opencensus API
	trace.RegisterExporter(exporter)
	// TODO: Change to probability sampler. Use "always sample" for now.
	trace.ApplyConfig(trace.Config{DefaultSampler: trace.AlwaysSample()})
	log.WithFields(log.Fields{
		"jaegerEndpoint":      jaegerEndpoint,
		"jaegerAgentEndpoint": jaegerAgentEndpoint,
	}).Info("tracing: Registered opencensus jaeger exporter for traces and stats")

	return exporter
}
