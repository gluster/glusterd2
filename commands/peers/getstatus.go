package peercommands

import (
	"io/ioutil"
	"net/http"
	"time"

	"github.com/gluster/glusterd2/etcdmgmt"
	"github.com/gluster/glusterd2/peer"
	"github.com/gluster/glusterd2/rest"

	"github.com/gorilla/mux"
)

func peerEtcdStatusHandler(w http.ResponseWriter, r *http.Request) {
	p := mux.Vars(r)

	id := p["peerid"]
	if id == "" {
		rest.SendHTTPError(w, http.StatusBadRequest, "Peer ID absent in request.")
		return
	}

	// Check that the peer is present in the store.
	if peerInfo, err := peer.GetPeerF(id); err != nil {
		rest.SendHTTPError(w, http.StatusNotFound, err.Error())
	} else {
		// Check the status of etcd instance running on that peer.
		resp, err := etcdmgmt.EtcdMemberStatus(peerInfo.MemberID)
		if err != nil {
			rest.SendHTTPError(w, http.StatusInternalServerError, "Could not fetch member status.")
			return
		}
		rest.SendHTTPResponse(w, http.StatusOK, resp)
	}
}

func peerEtcdHealthHandler(w http.ResponseWriter, r *http.Request) {
	p := mux.Vars(r)

	id := p["peerid"]
	if id == "" {
		rest.SendHTTPError(w, http.StatusBadRequest, "Peer ID absent in request.")
		return
	}

	// Check that the peer is present in the store.
	if peerInfo, err := peer.GetPeerF(id); err != nil {
		rest.SendHTTPError(w, http.StatusNotFound, err.Error())
	} else {
		// Check the health of etcd instance running on that peer.
		endpoint := "http://" + peerInfo.Addresses[0] + ":2379" + "/health"
		timeout := time.Duration(5 * time.Second)
		client := http.Client{
			Timeout: timeout,
		}
		resp, err := client.Get(endpoint)
		if err != nil {
			rest.SendHTTPError(w, http.StatusInternalServerError, "Health check failed.")
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
