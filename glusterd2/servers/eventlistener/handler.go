package eventlistener

import (
	"net"
	"strconv"
	"strings"

	gd2events "github.com/gluster/glusterd2/glusterd2/events"
	"github.com/gluster/glusterd2/plugins/events"
	eventsapi "github.com/gluster/glusterd2/plugins/events/api"

	log "github.com/sirupsen/logrus"
)

func getWebhooks() []*eventsapi.Webhook {
	webhooks, err := events.GetWebhookList()
	if err != nil {
		log.WithError(err).Error("Error retriving the wehook list from etcd")
	}

	return webhooks
}

func handleMessage(inMessage string, addr *net.UDPAddr) {
	data := strings.SplitN(inMessage, " ", 3)
	if len(data) != 3 {
		log.WithFields(log.Fields{
			"data": inMessage,
		}).Error("UDP Message received is in incorrect format")
		return
	}

	var msgDict = make(map[string]string)
	msgParts := strings.Split(data[2], ";")
	for _, msg := range msgParts {
		keyValue := strings.Split(msg, "=")
		key := strings.Trim(keyValue[0], " ")
		msgDict[key] = strings.Trim(keyValue[1], " ")
	}
	code, err := strconv.Atoi(data[1])
	if err != nil {
		log.WithError(err).Error("Error getting event code")
		return
	}
	if code > len(eventtypes)-1 {
		log.WithError(err).Error("Error in fetching event type")
		return
	}
	e := gd2events.New(eventtypes[code], msgDict, true)
	gd2events.Broadcast(e)
}
