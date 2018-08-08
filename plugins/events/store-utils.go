package events

import (
	"context"
	"encoding/json"
	"sort"
	"strings"

	"github.com/coreos/etcd/clientv3"
	"github.com/gluster/glusterd2/glusterd2/store"
	"github.com/gluster/glusterd2/pkg/api"
	eventsapi "github.com/gluster/glusterd2/plugins/events/api"

	log "github.com/sirupsen/logrus"
)

const (
	webhookPrefix string = "config/events/webhooks/"
	eventsPrefix         = "events/"
)

func webhookExists(webhookURL string) (bool, error) {
	resp, e := store.Get(context.TODO(), webhookPrefix+strings.Replace(webhookURL, "/", "|", -1))
	if e != nil {
		log.WithError(e).Error("Couldn't retrive webhook from store")
		return false, e
	}
	if resp.Count != 1 {
		return false, nil
	}
	return true, nil
}

// GetWebhookList returns list of all webhooks registered to glusterd
func GetWebhookList() ([]*eventsapi.Webhook, error) {
	resp, e := store.Get(context.TODO(), webhookPrefix, clientv3.WithPrefix())
	if e != nil {
		return nil, e
	}

	webhooks := make([]*eventsapi.Webhook, len(resp.Kvs))

	for i, kv := range resp.Kvs {
		var wh eventsapi.Webhook

		if err := json.Unmarshal(kv.Value, &wh); err != nil {
			log.WithError(err).WithField("webhook", string(kv.Key)).Error("Failed to unmarshal webhook")
			continue
		}

		webhooks[i] = &wh
	}

	return webhooks, nil
}

func addWebhook(webhook eventsapi.Webhook) error {
	wh, e := json.Marshal(webhook)
	if e != nil {
		log.WithError(e).Error("Failed to marshal the webhook object")
		return e
	}

	_, err := store.Put(context.TODO(), webhookPrefix+strings.Replace(webhook.URL, "/", "|", -1), string(wh))
	if err != nil {
		log.WithError(err).Error("Couldn't add webhook to store")
		return err
	}
	return nil
}

func deleteWebhook(webhookURL string) error {
	_, e := store.Delete(context.TODO(), webhookPrefix+strings.Replace(webhookURL, "/", "|", -1))
	return e
}

// GetEventsList returns list of Events recorded in last few minutes
func GetEventsList() ([]*api.Event, error) {
	resp, e := store.Get(context.TODO(), eventsPrefix, clientv3.WithPrefix())
	if e != nil {
		return nil, e
	}

	events := make([]*api.Event, len(resp.Kvs))

	for i, kv := range resp.Kvs {
		var ev api.Event

		if err := json.Unmarshal(kv.Value, &ev); err != nil {
			log.WithError(err).WithField(
				"event", string(kv.Key)).Error("Failed to unmarshal event")
			continue
		}

		events[i] = &ev
	}
	// Sort based on Event Timestamp
	sort.Slice(events, func(i, j int) bool { return int64(events[j].Timestamp.Sub(events[i].Timestamp)) > 0 })

	return events, nil
}
