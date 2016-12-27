package peercommands

import (
	"io/ioutil"
	"net/http"
	"time"

	"github.com/gluster/glusterd2/etcdmgmt"
	"github.com/gluster/glusterd2/peer"
	restutils "github.com/gluster/glusterd2/servers/rest/utils"

	log "github.com/Sirupsen/logrus"
	"github.com/gorilla/mux"
)

func peerEtcdStatusHandler(w http.ResponseWriter, r *http.Request) {
	p := mux.Vars(r)

	id := p["peerid"]
	if id == "" {
		restutils.SendHTTPError(w, http.StatusBadRequest, "Peer ID absent in request.")
		return
	}

	// Check that the peer is present in the store.
	if peerInfo, err := peer.GetPeerF(id); err != nil {
		restutils.SendHTTPError(w, http.StatusNotFound, err.Error())
	} else {
		// Check the status of etcd instance running on that peer.
		resp, err := etcdmgmt.EtcdMemberStatus(peerInfo.MemberID)
		if err != nil {
			restutils.SendHTTPError(w, http.StatusInternalServerError, "Could not fetch member status.")
			return
		}
		restutils.SendHTTPResponse(w, http.StatusOK, resp)
	}
}

func peerEtcdHealthHandler(w http.ResponseWriter, r *http.Request) {
	p := mux.Vars(r)

	id := p["peerid"]
	if id == "" {
		restutils.SendHTTPError(w, http.StatusBadRequest, "Peer ID absent in request.")
		return
	}

	// Check that the peer is present in the store.
	if peerInfo, err := peer.GetPeerF(id); err != nil {
		restutils.SendHTTPError(w, http.StatusNotFound, err.Error())
	} else {

		var endpoint string
		mlist, err := etcdmgmt.EtcdMemberList()
		if err != nil {
			msg := "Failed to list members in etcd cluster"
			log.WithField("error", err).Debug(msg)
			restutils.SendHTTPError(w, http.StatusInternalServerError, msg)
			return
		}

		for _, m := range mlist {
			if m.ID == peerInfo.MemberID {
				endpoint = m.ClientURLs[0]
			}
		}

		// Check the health of etcd instance running on that peer.
		healthEndpoint := endpoint + "/health"
		client := http.Client{
			Timeout: time.Duration(5 * time.Second),
		}
		resp, err := client.Get(healthEndpoint)
		if err != nil {
			restutils.SendHTTPError(w, http.StatusInternalServerError, "Health check failed.")
			return
		}
		w.Header().Set("Content-Type", "application/json; charset=UTF-8")
		w.WriteHeader(resp.StatusCode)
		payload, _ := ioutil.ReadAll(resp.Body)
		resp.Body.Close()
		w.Write(payload)
		return
	}
}
