package events

import (
	gd2events "github.com/gluster/glusterd2/glusterd2/events"
	"github.com/gluster/glusterd2/glusterd2/gdctx"
	"github.com/gluster/glusterd2/pkg/api"
	eventsapi "github.com/gluster/glusterd2/plugins/events/api"
	"github.com/pborman/uuid"

	log "github.com/sirupsen/logrus"
)

type webhooksNotifier struct{}

func (w *webhooksNotifier) Handle(e *api.Event) {
	//send events only from originator node
	if !uuid.Equal(e.Origin, gdctx.MyUUID) {
		return
	}
	// Get the list of registered Webhooks
	webhooks, err := GetWebhookList()
	if err != nil {
		log.WithError(err).Error("error retriving webhook list from etcd")
		return
	}

	for _, w := range webhooks {
		go func(e *api.Event, w *eventsapi.Webhook) {
			err = gd2events.WebhookPublish(w, e)
			if err != nil {
				log.WithError(err).Error("error in pushing data to webhook")
			}
		}(e, w)
	}

}

func (w *webhooksNotifier) Events() []string {
	return []string{}
}

func init() {
	w := new(webhooksNotifier)
	gd2events.Register(w)
}
