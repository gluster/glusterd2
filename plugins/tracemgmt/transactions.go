package tracemgmt

import (
	"github.com/gluster/glusterd2/glusterd2/gdctx"
	"github.com/gluster/glusterd2/glusterd2/transaction"
	"github.com/gluster/glusterd2/pkg/tracing"

	tracemgmtapi "github.com/gluster/glusterd2/plugins/tracemgmt/api"
	"github.com/gluster/glusterd2/plugins/tracemgmt/traceutils"

	log "github.com/sirupsen/logrus"
	"go.opencensus.io/exporter/jaeger"
	"go.opencensus.io/trace"
)

// Transaction step that validates the trace info passed
func txnTracingValidateConfig(c transaction.TxnCtx) error {
	var req tracemgmtapi.SetupTracingReq
	if err := c.Get("req", &req); err != nil {
		return err
	}

	// Validate the passed endpoints and sampling info
	if err := tracing.ValidateJaegerEndpoints(req.JaegerEndpoint, req.JaegerAgentEndpoint); err != nil {
		return err
	}

	var traceConfig tracemgmtapi.JaegerConfigInfo
	traceConfig.JaegerEndpoint = req.JaegerEndpoint
	traceConfig.JaegerAgentEndpoint = req.JaegerAgentEndpoint
	traceConfig.JaegerSampler = req.JaegerSampler
	traceConfig.JaegerSampleFraction = req.JaegerSampleFraction
	traceConfig.Status = tracemgmtapi.TracingEnabled

	// Validate the sampler
	if err := tracing.ValidateJaegerSampler(req.JaegerSampler); err != nil {
		traceConfig.JaegerSampler = int(tracing.Never)
		traceConfig.Status = tracemgmtapi.TracingDisabled
	}

	// Validate the sample fraction, if required
	switch tracing.JaegerSamplerType(req.JaegerSampler) {
	case tracing.Never:
		traceConfig.JaegerSampleFraction = 0.0
	case tracing.Always:
		traceConfig.JaegerSampleFraction = 1.0
	case tracing.Probabilistic:
		if err := tracing.ValidateJaegerProbSampleFraction(req.JaegerSampler, req.JaegerSampleFraction); err != nil {
			traceConfig.JaegerSampleFraction = tracing.DefaultSampleFraction
		}
	}

	// Save the trace config in the context
	err := c.Set("traceconfig", traceConfig)

	return err
}

// storeTraceConfig uses the passed context key to get
// trace config and updates it into the store.
func storeTraceConfig(c transaction.TxnCtx, key string) error {
	var traceConfig tracemgmtapi.JaegerConfigInfo
	if err := c.Get(key, &traceConfig); err != nil {
		return err
	}

	err := traceutils.AddOrUpdateTraceConfig(&traceConfig)
	return err
}

// Transaction step that stores the trace config
func txnTracingStoreConfig(c transaction.TxnCtx) error {
	err := storeTraceConfig(c, "traceconfig")
	return err
}

// Transaction step that reverts the trace config
func txnTracingUndoStoreConfig(c transaction.TxnCtx) error {
	err := storeTraceConfig(c, "oldtraceconfig")
	return err
}

// Transaction step that deletes the trace config
func txnTracingDeleteStoreConfig(c transaction.TxnCtx) error {
	err := traceutils.DeleteTraceConfig()
	return err
}

// Tranasaction step that reads trace config info from the store and updates
// the in-memory trace config.
func txnTracingApplyNewConfig(c transaction.TxnCtx) error {
	var traceConfig tracemgmtapi.JaegerConfigInfo
	if err := c.Get("traceconfig", &traceConfig); err != nil {
		return err
	}

	// Create the new Opencensus exporter
	exporter, err := jaeger.NewExporter(jaeger.Options{
		Endpoint:      traceConfig.JaegerEndpoint,
		AgentEndpoint: traceConfig.JaegerAgentEndpoint,
		ServiceName:   gdctx.HostName,
	})

	if err != nil {
		c.Logger().WithError(err).WithFields(log.Fields{
			"jaegerEndpoint":      traceConfig.JaegerEndpoint,
			"jaegerAgentEndpoint": traceConfig.JaegerAgentEndpoint,
		}).Warning("tracingTxn: Unable to create Opencensus jaeger exporter")
		return err
	}

	// Unregister the old Jaeger exporter if it exists
	if oldExporter := tracing.JaegerExporter(); oldExporter != nil {
		trace.UnregisterExporter(oldExporter)
	}

	// Register the new exporter
	trace.RegisterExporter(exporter)

	// Apply the sample type based on config settings
	tracing.ApplySampler(traceConfig.JaegerSampler, traceConfig.JaegerSampleFraction)

	// Set the global Jaeger trace config
	traceCfg := tracing.JaegerTraceConfig{
		JaegerEndpoint:       traceConfig.JaegerEndpoint,
		JaegerAgentEndpoint:  traceConfig.JaegerAgentEndpoint,
		JaegerSampler:        tracing.JaegerSamplerType(traceConfig.JaegerSampler),
		JaegerSampleFraction: traceConfig.JaegerSampleFraction,
	}
	tracing.SetJaegerTraceConfig(traceCfg)

	// Set the new Jaeger exporter
	tracing.SetJaegerExporter(exporter)

	return nil
}
