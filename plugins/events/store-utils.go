package events

import (
	"context"
	"encoding/json"
	"strings"

	"github.com/coreos/etcd/clientv3"
	"github.com/gluster/glusterd2/glusterd2/store"
	eventsapi "github.com/gluster/glusterd2/plugins/events/api"

	log "github.com/sirupsen/logrus"
)

const (
	webhookPrefix string = "config/events/webhooks/"
)

func webhookExists(webhook eventsapi.Webhook) (bool, error) {
	resp, e := store.Store.Get(context.TODO(), webhookPrefix+strings.Replace(webhook.URL, "/", "|", -1))
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
	resp, e := store.Store.Get(context.TODO(), webhookPrefix, clientv3.WithPrefix())
	if e != nil {
		return nil, e
	}

	webhooks := make([]*eventsapi.Webhook, len(resp.Kvs))

	for i, kv := range resp.Kvs {
		var wh eventsapi.Webhook

		if err := json.Unmarshal(kv.Value, &wh); err != nil {
			log.WithFields(log.Fields{
				"webhook": string(kv.Key),
				"error":   err,
			}).Error("Failed to unmarshal webhook")
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

	_, err := store.Store.Put(context.TODO(), webhookPrefix+strings.Replace(webhook.URL, "/", "|", -1), string(wh))
	if err != nil {
		log.WithError(err).Error("Couldn't add webhook to store")
		return err
	}
	return nil
}

func deleteWebhook(webhook eventsapi.Webhook) error {
	_, e := store.Store.Delete(context.TODO(), webhookPrefix+strings.Replace(webhook.URL, "/", "|", -1))
	return e
}
