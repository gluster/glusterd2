package events

import (
	"net/http"

	gd2events "github.com/gluster/glusterd2/glusterd2/events"
	"github.com/gluster/glusterd2/glusterd2/gdctx"
	"github.com/gluster/glusterd2/glusterd2/peer"
	restutils "github.com/gluster/glusterd2/glusterd2/servers/rest/utils"
	"github.com/gluster/glusterd2/glusterd2/transaction"
	"github.com/gluster/glusterd2/pkg/errors"
	eventsapi "github.com/gluster/glusterd2/plugins/events/api"
)

const (
	eventsapiPrefix string = "events/"
)

func webhookAddHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	var req eventsapi.Webhook
	if err := restutils.UnmarshalRequest(r, &req); err != nil {
		restutils.SendHTTPError(
			ctx, w, http.StatusBadRequest,
			errors.ErrJSONParsingFailed)
		return
	}

	if req.URL == "" {
		restutils.SendHTTPError(ctx, w, http.StatusBadRequest, "Webhook Url is required field")
		return
	}

	// Check if the webhook already exists
	exists, err := webhookExists(req.URL)
	if err != nil {
		restutils.SendHTTPError(
			ctx, w, http.StatusInternalServerError,
			"Could not check if webhook already exists")
		return
	}
	if exists {
		restutils.SendHTTPError(ctx, w, http.StatusConflict, "Webhook already exists")
		return
	}

	if err := addWebhook(req); err != nil {
		restutils.SendHTTPError(
			ctx, w, http.StatusInternalServerError,
			"Could not add webhook")
		return
	}

	restutils.SendHTTPResponse(ctx, w, http.StatusOK, nil)
}

func webhookDeleteHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	var req eventsapi.WebhookDel
	if err := restutils.UnmarshalRequest(r, &req); err != nil {
		restutils.SendHTTPError(ctx, w, http.StatusBadRequest,
			errors.ErrJSONParsingFailed)
		return
	}

	if req.URL == "" {
		restutils.SendHTTPError(ctx, w, http.StatusBadRequest, "Webhook Url is required field")
		return
	}

	// Check if the webhook already exists
	exists, err := webhookExists(req.URL)
	if err != nil {
		restutils.SendHTTPError(
			ctx, w, http.StatusInternalServerError,
			"Could not check if webhook already exists")
		return
	}
	if !exists {
		restutils.SendHTTPError(ctx, w, http.StatusConflict, "Webhook does not exist")
		return
	}

	if err := deleteWebhook(req.URL); err != nil {
		restutils.SendHTTPError(
			ctx, w, http.StatusInternalServerError,
			"Could not delete webhook")
		return
	}
	restutils.SendHTTPResponse(ctx, w, http.StatusNoContent, nil)
}

func webhookListHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	webhooks, err := GetWebhookList()
	if err != nil {
		restutils.SendHTTPError(
			ctx, w, http.StatusInternalServerError,
			"Could not retrive webhook list")
		return
	}

	var resp eventsapi.WebhookList

	for _, wh := range webhooks {
		resp = append(resp, wh.URL)
	}

	restutils.SendHTTPResponse(ctx, w, http.StatusOK, resp)
}

func eventsListHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	events, err := GetEventsList()
	if err != nil {
		restutils.SendHTTPError(
			ctx, w, http.StatusInternalServerError,
			"Could not retrive events list")
		return
	}

	restutils.SendHTTPResponse(ctx, w, http.StatusOK, events)
}

func checkConnection(c transaction.TxnCtx) error {
	var req eventsapi.Webhook

	if err := c.Get("req", &req); err != nil {
		return err
	}
	e := gd2events.New("TEST", map[string]string{}, false)
	return gd2events.WebhookPublish(&req, e)
}

func registerWebhookTestStepFuncs() {
	var sfs = []struct {
		name string
		sf   transaction.StepFunc
	}{
		{"webhook-test.checkConnection", checkConnection},
	}
	for _, sf := range sfs {
		transaction.RegisterStepFunc(sf.sf, sf.name)
	}
}

func webhookTestHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	logger := gdctx.GetReqLogger(ctx)
	var req eventsapi.Webhook

	if err := restutils.UnmarshalRequest(r, &req); err != nil {
		restutils.SendHTTPError(
			ctx, w, http.StatusBadRequest,
			errors.ErrJSONParsingFailed)
		return
	}

	if req.URL == "" {
		restutils.SendHTTPError(ctx, w, http.StatusBadRequest, "webhook URL is required field")
		return
	}
	allNodes, err := peer.GetPeerIDs()
	if err != nil {
		restutils.SendHTTPError(ctx, w, http.StatusInternalServerError, err)
		return
	}

	txn := transaction.NewTxn(ctx)
	defer txn.Done()

	txn.Steps = []*transaction.Step{
		{
			DoFunc: "webhook-test.checkConnection",
			Nodes:  allNodes,
		},
	}

	if err := txn.Ctx.Set("req", &req); err != nil {
		restutils.SendHTTPError(ctx, w, http.StatusInternalServerError, err)
		return
	}

	if err := txn.Do(); err != nil {
		logger.WithError(err).Error("webhook-test transaction failed")
		status, err := restutils.ErrToStatusCode(err)
		restutils.SendHTTPError(ctx, w, status, err)
		return
	}

	restutils.SendHTTPResponse(ctx, w, http.StatusOK, nil)
}
