package eventlistener

import (
	"encoding/json"
	"net"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/gluster/glusterd2/glusterd2/gdctx"
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

func getJWTToken(url string, secret string) string {
	//TODO generate the gwt token from the sceret
	return ""
}

func webhookPublish(webhook *eventsapi.Webhook, message string) {
	body := strings.NewReader(message)

	req, err := http.NewRequest("POST", webhook.URL, body)
	if err != nil {
		log.WithError(err).Error("Error forming the request object")
		return
	}
	req.Header.Set("Content-Type", "application/json")

	if webhook.Token != "" {
		req.Header.Set("Authorization", "bearer "+webhook.Token)
	}

	if webhook.Secret != "" {
		token := getJWTToken(webhook.URL, webhook.Secret)
		req.Header.Set("Authorization", "bearer "+token)
	}

	tr := &http.Transport{
		DisableCompression:    true,
		DisableKeepAlives:     true,
		ResponseHeaderTimeout: 3 * time.Second,
	}

	client := &http.Client{Transport: tr}

	resp, err := client.Do(req)
	if err != nil {
		log.WithError(err).Error("Error while publishing event to webhook")
		return
	}

	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		log.Error("Webhook responded with status: ", string(resp.StatusCode))
		return
	}
	return
}

func handleMessage(inMessage string, addr *net.UDPAddr) {
	data := strings.SplitN(inMessage, " ", 3)
	if len(data) != 3 {
		log.WithFields(log.Fields{
			"data": inMessage,
		}).Error("UDP Message received is in incorrect format")
		return
	}

	var msgDict = make(map[string]interface{})
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
	if eventtypes[code] == "VOLUME_SET" {
		optsdata := ""
		optsdataRaw, ok := msgDict["options"]
		if ok {
			optsdata = optsdataRaw.(string)
		}

		optsdata = strings.Trim(optsdata, ",")

		var opts [][]string

		optpair := []string{}
		for i, opt := range strings.Split(optsdata, ",") {
			if i%2 == 0 {
				optpair = []string{opt}
			} else {
				optpair = append(optpair, opt)
				opts = append(opts, optpair)
			}
		}
		msgDict["options"] = opts
	}

	message := make(map[string]interface{})
	message["ts"] = time.Now().Unix()
	message["nodeid"] = gdctx.MyUUID.String()
	message["message"] = msgDict

	marshalledMsg, err := json.Marshal(message)
	if err != nil {
		log.WithError(err).Error("Error while marshalling the message")
		return
	}
	log.Info("Posting the event: ", string(marshalledMsg))

	// Broadcast internally
	// TODO

	// Get the list of registered Webhooks and then Push
	for _, w := range getWebhooks() {
		// Below func is called as async, failures are handled by
		// goroutine itself
		go webhookPublish(w, string(marshalledMsg))
	}
}
