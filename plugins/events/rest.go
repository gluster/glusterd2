package events

import (
	"github.com/gluster/glusterd2/pkg/errors"
	"net/http"

	restutils "github.com/gluster/glusterd2/glusterd2/servers/rest/utils"
	"github.com/gluster/glusterd2/pkg/api"
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
			ctx, w, http.StatusUnprocessableEntity,
			errors.ErrJSONParsingFailed.Error(), api.ErrCodeDefault)
		return
	}

	if req.URL == "" {
		restutils.SendHTTPError(ctx, w, http.StatusBadRequest, "Webhook Url is required field", api.ErrCodeDefault)
		return
	}

	// Check if the webhook already exists
	exists, err := webhookExists(req)
	if err != nil {
		restutils.SendHTTPError(
			ctx, w, http.StatusInternalServerError,
			"Could not check if webhook already exists", api.ErrCodeDefault)
		return
	}
	if exists {
		restutils.SendHTTPError(ctx, w, http.StatusConflict, "Webhook already exists", api.ErrCodeDefault)
		return
	}

	if err := addWebhook(req); err != nil {
		restutils.SendHTTPError(
			ctx, w, http.StatusInternalServerError,
			"Could not add webhook", api.ErrCodeDefault)
		return
	}

	restutils.SendHTTPResponse(ctx, w, http.StatusOK, "Webhook Added")
}

func webhookDeleteHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	var req eventsapi.Webhook
	if err := restutils.UnmarshalRequest(r, &req); err != nil {
		restutils.SendHTTPError(ctx, w, http.StatusUnprocessableEntity,
			errors.ErrJSONParsingFailed.Error(), api.ErrCodeDefault)
		return
	}

	if req.URL == "" {
		restutils.SendHTTPError(ctx, w, http.StatusBadRequest, "Webhook Url is required field", api.ErrCodeDefault)
		return
	}

	// Check if the webhook already exists
	exists, err := webhookExists(req)
	if err != nil {
		restutils.SendHTTPError(
			ctx, w, http.StatusInternalServerError,
			"Could not check if webhook already exists",
			api.ErrCodeDefault)
		return
	}
	if !exists {
		restutils.SendHTTPError(ctx, w, http.StatusConflict, "Webhook does not exist", api.ErrCodeDefault)
		return
	}

	if err := deleteWebhook(req); err != nil {
		restutils.SendHTTPError(
			ctx, w, http.StatusInternalServerError,
			"Could not delete webhook",
			api.ErrCodeDefault)
		return
	}
	restutils.SendHTTPResponse(ctx, w, http.StatusOK, "Webhook Deleted")
}

func webhookListHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	webhooks, err := GetWebhookList()
	if err != nil {
		restutils.SendHTTPError(
			ctx, w, http.StatusInternalServerError,
			"Could not retrive webhook list",
			api.ErrCodeDefault)
		return
	}

	var resp eventsapi.WebhookList

	for _, wh := range webhooks {
		resp = append(resp, wh.URL)
	}

	restutils.SendHTTPResponse(ctx, w, http.StatusOK, resp)
}
