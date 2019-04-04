package api

// SetupTracingReq structure
type SetupTracingReq struct {
	JaegerEndpoint       string  `json:"jaeger-endpoint"`
	JaegerAgentEndpoint  string  `json:"jaeger-agent-endpoint"`
	JaegerSampler        int     `json:"jaeger-sampler"`
	JaegerSampleFraction float64 `json:"jaeger-sample-fraction"`
}
