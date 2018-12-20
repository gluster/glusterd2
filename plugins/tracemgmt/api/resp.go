package api

const (
	// TracingEnabled represents enabled state
	TracingEnabled = "enabled"

	// TracingDisabled represents disabled state
	TracingDisabled = "disabled"
)

// JaegerConfigInfo represents tracing configration
type JaegerConfigInfo struct {
	Status               string  `json:"tracing-status"`
	JaegerEndpoint       string  `json:"jaeger-endpoint"`
	JaegerAgentEndpoint  string  `json:"jaeger-agent-endpoint"`
	JaegerSampler        int     `json:"jaeger-sampler"`
	JaegerSampleFraction float64 `json:"jaeger-sample-fraction"`
}
