package api

import (
	"github.com/gluster/glusterd2/pkg/api"
)

// WebhookList holds list of webhooks containing just its URL
type WebhookList []string

// EventList holds list of events happened in last 10 mins(configurable)
type EventList []api.Event
