package tracemgmt

import (
	"net/http"

	"github.com/gluster/glusterd2/glusterd2/gdctx"
	"github.com/gluster/glusterd2/glusterd2/peer"
	restutils "github.com/gluster/glusterd2/glusterd2/servers/rest/utils"
	"github.com/gluster/glusterd2/glusterd2/transaction"
	transactionv2 "github.com/gluster/glusterd2/glusterd2/transactionv2"
	"github.com/gluster/glusterd2/pkg/errors"
	tracemgmtapi "github.com/gluster/glusterd2/plugins/tracemgmt/api"
	"github.com/gluster/glusterd2/plugins/tracemgmt/traceutils"

	"github.com/pborman/uuid"
)

func tracingEnableHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	logger := gdctx.GetReqLogger(ctx)

	var req tracemgmtapi.SetupTracingReq
	// Parse the JSON body to get details of trace request
	if err := restutils.UnmarshalRequest(r, &req); err != nil {
		logger.WithError(err).Error("Failed to unmarshal SetupTracingReq")
		restutils.SendHTTPError(ctx, w, http.StatusBadRequest, errors.ErrJSONParsingFailed)
		return
	}

	// If trace configuration already exists in store, then the config must be
	// updated instead. Send an error back.
	if _, err := traceutils.GetTraceConfig(); err == nil {
		restutils.SendHTTPError(ctx, w, http.StatusConflict, "Trace configuration already exists")
		return
	}

	txn, err := transactionv2.NewTxnWithLocks(ctx, gdctx.MyClusterID.String())
	if err != nil {
		status, err := restutils.ErrToStatusCode(err)
		restutils.SendHTTPError(ctx, w, status, err)
		return
	}
	defer txn.Done()

	// Store new config request
	if err := txn.Ctx.Set("req", &req); err != nil {
		restutils.SendHTTPError(ctx, w, http.StatusInternalServerError, err)
		return
	}

	nodes, err := peer.GetPeerIDs()
	if err != nil {
		restutils.SendHTTPError(ctx, w, http.StatusInternalServerError, err)
		return
	}

	txn.Steps = []*transaction.Step{
		{
			DoFunc: "trace-mgmt.ValidateTraceConfig",
			Nodes:  []uuid.UUID{gdctx.MyUUID},
		},
		{
			DoFunc:   "trace-mgmt.StoreTraceConfig",
			UndoFunc: "trace-mgmt.UndoStoreTraceConfig",
			Nodes:    []uuid.UUID{gdctx.MyUUID},
			Sync:     true,
		},
		{
			DoFunc: "trace-mgmt.NotifyTraceConfigChange",
			Nodes:  nodes,
			Sync:   true,
		},
	}

	if err = txn.Do(); err != nil {
		logger.WithError(err).Error("Failed to enable trace configuration")
		restutils.SendHTTPError(ctx, w, http.StatusInternalServerError, err)
		return
	}

	traceConfig, err := traceutils.GetTraceConfig()
	if err != nil {
		restutils.SendHTTPError(ctx, w, http.StatusInternalServerError, "Failed to get trace configuration from store")
		return
	}

	restutils.SendHTTPResponse(ctx, w, http.StatusCreated, traceConfig)
}

func tracingStatusHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	// Get the trace configuration from the store.
	traceConfig, err := traceutils.GetTraceConfig()
	if err != nil {
		traceConfig = &tracemgmtapi.JaegerConfigInfo{
			Status:               tracemgmtapi.TracingDisabled,
			JaegerEndpoint:       "",
			JaegerAgentEndpoint:  "",
			JaegerSampler:        0,
			JaegerSampleFraction: 0.0,
		}
	}

	restutils.SendHTTPResponse(ctx, w, http.StatusOK, traceConfig)
}

func tracingUpdateHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	logger := gdctx.GetReqLogger(ctx)

	var req tracemgmtapi.SetupTracingReq
	// Parse the JSON body to get details of trace request
	if err := restutils.UnmarshalRequest(r, &req); err != nil {
		logger.WithError(err).Error("Failed to unmarshal SetupTracingReq")
		restutils.SendHTTPError(ctx, w, http.StatusBadRequest, errors.ErrJSONParsingFailed)
		return
	}

	// Get the current trace config from the store
	traceConfig, err := traceutils.GetTraceConfig()
	if err != nil {
		restutils.SendHTTPError(ctx, w, http.StatusNotFound, "No trace configuration exists")
		return
	}

	txn, err := transactionv2.NewTxnWithLocks(ctx, gdctx.MyClusterID.String())
	if err != nil {
		status, err := restutils.ErrToStatusCode(err)
		restutils.SendHTTPError(ctx, w, status, err)
		return
	}
	defer txn.Done()

	// Update the http request with either the old trace option
	// to preserve the existing config or set the new  one passed in
	// the request. This step is neccessary as the client may
	// not request all options to be changed.
	if req.JaegerEndpoint == "" {
		req.JaegerEndpoint = traceConfig.JaegerEndpoint
	}

	if req.JaegerAgentEndpoint == "" {
		req.JaegerAgentEndpoint = traceConfig.JaegerAgentEndpoint
	}

	// Save the current config in the context
	if err := txn.Ctx.Set("oldtraceconfig", traceConfig); err != nil {
		restutils.SendHTTPError(ctx, w, http.StatusInternalServerError, err)
		return
	}

	// Store new config request
	if err := txn.Ctx.Set("req", &req); err != nil {
		restutils.SendHTTPError(ctx, w, http.StatusInternalServerError, err)
		return
	}

	nodes, err := peer.GetPeerIDs()
	if err != nil {
		restutils.SendHTTPError(ctx, w, http.StatusInternalServerError, err)
		return
	}

	txn.Steps = []*transaction.Step{
		{
			DoFunc: "trace-mgmt.ValidateTraceConfig",
			Nodes:  []uuid.UUID{gdctx.MyUUID},
		},
		{
			DoFunc:   "trace-mgmt.StoreTraceConfig",
			UndoFunc: "trace-mgmt.RestoreTraceConfig",
			Nodes:    []uuid.UUID{gdctx.MyUUID},
			Sync:     true,
		},
		{
			DoFunc: "trace-mgmt.NotifyTraceConfigChange",
			Nodes:  nodes,
			Sync:   true,
		},
	}

	if err = txn.Do(); err != nil {
		logger.WithError(err).Error("Failed to update trace configuration")
		restutils.SendHTTPError(ctx, w, http.StatusInternalServerError, err)
		return
	}

	traceConfig, err = traceutils.GetTraceConfig()
	if err != nil {
		restutils.SendHTTPError(ctx, w, http.StatusInternalServerError, "Failed to get trace configuration from store")
		return
	}

	restutils.SendHTTPResponse(ctx, w, http.StatusOK, traceConfig)
}
