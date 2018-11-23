// Package tracing implements a common tracing initialization for GD2 and CLI
package tracing

import (
	errs "errors"
	"sync"

	"github.com/gluster/glusterd2/glusterd2/gdctx"

	log "github.com/sirupsen/logrus"
	flag "github.com/spf13/pflag"
	config "github.com/spf13/viper"
	"go.opencensus.io/exporter/jaeger"
	"go.opencensus.io/trace"
)

// Commandline options for exporter endpoints
const (
	jaegerEndpointOpt       = "jaeger-endpoint"
	jaegerAgentEndpointOpt  = "jaeger-agent-endpoint"
	jaegerSamplerOpt        = "jaeger-sampler"
	jaegerSampleFractionOpt = "jaeger-sample-fraction"
)

// JaegerSamplerType to indicate different Jaeger sampler type
type JaegerSamplerType uint8

// Sampler types in Jaeger
const (
	// 'Never' - Don't sample any trace
	Never JaegerSamplerType = iota
	// 'Always' - Sample every trace
	Always
	// 'Probabilistic' - Sample based on sample fraction
	Probabilistic
)

// SamplerTypeToString returns string representation of sampler type.
func SamplerTypeToString(samplerType JaegerSamplerType) string {
	sampler := ""
	switch samplerType {
	case Never:
		sampler = "Never"
	case Always:
		sampler = "Always"
	case Probabilistic:
		sampler = "Probabilistic"
	default:
		sampler = "Unknown"
	}
	return sampler
}

// DefaultSampleFraction - Default sample fraction. By default every 1 in 10
// traces will be sampled.
var DefaultSampleFraction = 0.1

// JaegerTraceConfig holds Jaeger trace information
type JaegerTraceConfig struct {
	// Jaeger collector endpoint to which agent sends spans in batches
	JaegerEndpoint string
	// Jaeger agent endpoint to which spans are sent
	JaegerAgentEndpoint string
	// Jaeger sampler (0 - never or 1 - always or 2 - probabilistic)
	JaegerSampler JaegerSamplerType
	// Jaeger sample fraction to use in case sampler type is probabilistic
	JaegerSampleFraction float64
}

// Jaeger exporter
var jaegerExporter *jaeger.Exporter

// Global Jaeger trace config
var jaegerTraceConfig *JaegerTraceConfig
var once sync.Once

// ValidateJaegerEndpoints validates Jaeger endpoints
func ValidateJaegerEndpoints(endpoint, agentEndpoint string) error {
	// Return error if either endpoints are not specified
	if endpoint == "" || agentEndpoint == "" {
		log.WithFields(log.Fields{
			"jaegerEndpoint":      endpoint,
			"jaegerAgentEndpoint": agentEndpoint,
		}).Warning("tracing: One or more Jaeger endpoints not specified")
		return errs.New("Invalid Jaeger endpoints")
	}
	return nil
}

// ValidateJaegerSampler validates sampler type
func ValidateJaegerSampler(sampler int) error {
	jsampler := JaegerSamplerType(sampler)
	if !(jsampler == Never || jsampler == Always || jsampler == Probabilistic) {
		log.WithFields(log.Fields{
			"jaegerSamplerTypeInt": jsampler,
			"jaegerSamplerTypeStr": SamplerTypeToString(jsampler),
		}).Warning("tracing: Invalid sampler type option provided. Tracing is disabled.")
		return errs.New("Invalid sampler type")
	}
	return nil
}

// ValidateJaegerProbSampleFraction validates sample fraction in case sampler
// type is probabilistic
func ValidateJaegerProbSampleFraction(sampler int, sampleFraction float64) error {
	jsampler := JaegerSamplerType(sampler)
	// Nothing to do if sampler is not probabilistic
	if jsampler != Probabilistic {
		return nil
	}

	if sampleFraction <= 0.0 || sampleFraction >= 1.0 {
		// Invalid sample fraction provided, use default value.
		log.WithFields(log.Fields{
			"jaegerSampleFraction":  sampleFraction,
			"defaultSampleFraction": DefaultSampleFraction,
		}).Warning("tracing: Invalid sample fraction provided. Applying default value.")
		return errs.New("Invalid sample fraction")
	}
	return nil
}

// InitJaegerTraceConfig initializes and returns the global Jaeger trace config.
// Note that jaegerTraceConfig is instantiated only once.
// Subsequent calls to this function returns a pointer to
// the instantiated object.
func InitJaegerTraceConfig() *JaegerTraceConfig {
	once.Do(func() {
		jaegerTraceConfig = new(JaegerTraceConfig)
	})
	return jaegerTraceConfig
}

// SetJaegerTraceConfig sets the global Jaeger trace config values
func SetJaegerTraceConfig(t JaegerTraceConfig) {
	// Get the global Jaeger trace config
	j := InitJaegerTraceConfig()

	// Set the Jaeger trace config
	j.JaegerEndpoint = t.JaegerEndpoint
	j.JaegerAgentEndpoint = t.JaegerAgentEndpoint
	j.JaegerSampler = t.JaegerSampler
	j.JaegerSampleFraction = t.JaegerSampleFraction

	log.WithFields(log.Fields{
		"jaegerEndpoint":       j.JaegerEndpoint,
		"jaegerAgentEndpoint":  j.JaegerAgentEndpoint,
		"jaegerSamplerType":    SamplerTypeToString(j.JaegerSampler),
		"jaegerSampleFraction": j.JaegerSampleFraction,
	}).Info("tracing: Registered opencensus jaeger exporter for traces and stats")
}

// SetJaegerExporter sets the current active Jaeger exporter
func SetJaegerExporter(exporter *jaeger.Exporter) {
	jaegerExporter = exporter
}

// JaegerExporter gets the current active Jaeger exporter
func JaegerExporter() *jaeger.Exporter {
	return jaegerExporter
}

// Flush any outstanding spans to the Jaeger endpoint
func Flush() {
	if jaegerExporter != nil {
		jaegerExporter.Flush()
	}
}

// InitFlags initializes the command line options for GD2 tracing endpoints
func InitFlags() {
	flag.String(jaegerEndpointOpt, "", "Jaeger collector endpoint that accepts spans from Jaeger agent.")
	flag.String(jaegerAgentEndpointOpt, "", "Jaeger agent endpoint that the Jaeger client sends spans to.")
	flag.String(jaegerSamplerOpt, "", "Jaeger sampler to employ (0 - never or 1 - always or 2 - probabilistic).")
	flag.String(jaegerSampleFractionOpt, "", "Jaeger sample fraction to use if sampler type is set to probabilistic.")
}

// ApplySampler sets the desired sampler type
func ApplySampler(sampler int, sampleFraction float64) {
	// Convert sampler to JaegerSamplerType before applying
	jsampler := JaegerSamplerType(sampler)
	// Apply the sampler type based on config settings
	switch jsampler {
	case Never: // Disable tracing
		trace.ApplyConfig(trace.Config{DefaultSampler: trace.NeverSample()})
	case Always: // Sample ALL traces.
		trace.ApplyConfig(trace.Config{DefaultSampler: trace.AlwaysSample()})
	case Probabilistic: // Sample traces based on sample fraction.
		trace.ApplyConfig(trace.Config{DefaultSampler: trace.ProbabilitySampler(sampleFraction)})
	default: // Use the default sample fraction
		trace.ApplyConfig(trace.Config{DefaultSampler: trace.ProbabilitySampler(DefaultSampleFraction)})
	}
}

// InitJaegerExporter initializes the jaeger exporter as the tracing endpoint
// This should be called early when a process starts.
// This creates and returns an exporter if successful.
// Otherwise, a warning is logged and a 'nil' exporter is returned.
func InitJaegerExporter() *jaeger.Exporter {
	// Get the Jaeger endpoints
	endpoint := config.GetString(jaegerEndpointOpt)
	agentEndpoint := config.GetString(jaegerAgentEndpointOpt)

	// Validate endpoints and return nil exporter if invalid
	if err := ValidateJaegerEndpoints(endpoint, agentEndpoint); err != nil {
		return nil
	}

	// Create the Opencensus Jaeger exporter
	exporter, err := jaeger.NewExporter(jaeger.Options{
		Endpoint:      endpoint,
		AgentEndpoint: agentEndpoint,
		ServiceName:   gdctx.HostName,
	})

	if err != nil {
		log.WithError(err).WithFields(log.Fields{
			"jaegerEndpoint":      endpoint,
			"jaegerAgentEndpoint": agentEndpoint,
		}).Warning("tracing: Unable to create opencensus jaeger exporter")
		return nil
	}

	// Register the Jaeger exporter
	trace.RegisterExporter(exporter)

	// Get the sampler type & sample fraction if required
	sampler := config.GetInt(jaegerSamplerOpt)

	// Validate sampler. Disable tracing if invalid.
	if err := ValidateJaegerSampler(sampler); err != nil {
		sampler = int(Never)
	}

	// Validate the sample fraction. Implicitly set the sample fraction if the
	// sampler is either "Never" or "Always". Client is not expected to set
	// sample fraction for such samplers.
	sampleFraction := config.GetFloat64(jaegerSampleFractionOpt)
	switch JaegerSamplerType(sampler) {
	case Never:
		sampleFraction = 0.0
	case Always:
		sampleFraction = 1.0
	case Probabilistic:
		if err := ValidateJaegerProbSampleFraction(sampler, sampleFraction); err != nil {
			// Set default sample fraction
			sampleFraction = DefaultSampleFraction
		}
	}

	// Apply the sample type based on config settings
	ApplySampler(sampler, sampleFraction)

	// Initialize and set the global jaegerTraceConfig
	traceConfig := JaegerTraceConfig{
		JaegerEndpoint:       endpoint,
		JaegerAgentEndpoint:  agentEndpoint,
		JaegerSampler:        JaegerSamplerType(sampler),
		JaegerSampleFraction: sampleFraction,
	}
	SetJaegerTraceConfig(traceConfig)

	// Set the new Jaeger exporter
	jaegerExporter = exporter

	return jaegerExporter
}
