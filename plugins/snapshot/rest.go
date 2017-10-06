package snapshot

import (
	"net/http"

	restutils "github.com/gluster/glusterd2/servers/rest/utils"
)

func snapshotCreateHandler(w http.ResponseWriter, r *http.Request) {
	restutils.SendHTTPResponse(w, http.StatusOK, "Snapshot Create")
}

func snapshotActivateHandler(w http.ResponseWriter, r *http.Request) {
	restutils.SendHTTPResponse(w, http.StatusOK, "Snapshot Activate")
}

func snapshotDeactivateHandler(w http.ResponseWriter, r *http.Request) {
	restutils.SendHTTPResponse(w, http.StatusOK, "Snapshot Deactivate")
}

func snapshotCloneHandler(w http.ResponseWriter, r *http.Request) {
	restutils.SendHTTPResponse(w, http.StatusOK, "Snapshot Clone")
}

func snapshotRestoreHandler(w http.ResponseWriter, r *http.Request) {
	restutils.SendHTTPResponse(w, http.StatusOK, "Snapshot Restore")
}

func snapshotStatusHandler(w http.ResponseWriter, r *http.Request) {
	restutils.SendHTTPResponse(w, http.StatusOK, "Snapshot Status/Info")
}

func snapshotDeleteHandler(w http.ResponseWriter, r *http.Request) {
	restutils.SendHTTPResponse(w, http.StatusOK, "Snapshot Delete")
}

func snapshotConfigGetHandler(w http.ResponseWriter, r *http.Request) {
	restutils.SendHTTPResponse(w, http.StatusOK, "Snapshot Config Get")
}

func snapshotConfigSetHandler(w http.ResponseWriter, r *http.Request) {
	restutils.SendHTTPResponse(w, http.StatusOK, "Snapshot Config Set")
}

func snapshotConfigResetHandler(w http.ResponseWriter, r *http.Request) {
	restutils.SendHTTPResponse(w, http.StatusOK, "Snapshot Config Reset")
}
