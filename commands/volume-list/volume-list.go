package volumelist

import (
	"encoding/json"
	"net/http"

	"github.com/kshlm/glusterd2/context"
	"github.com/kshlm/glusterd2/rest"

	log "github.com/Sirupsen/logrus"
)

type VolumeListCommand struct {
}

func (c *VolumeListCommand) VolumeList(w http.ResponseWriter, r *http.Request) {

	log.Info("In Volume list API")

	volumes, e := context.Store.GetVolumes()
	if e != nil {
		http.Error(w, e.Error(), http.StatusInternalServerError)
		return
	}
	// Write nsg
	w.Header().Set("Content-Type", "text/plain; charset=UTF-8")
	w.WriteHeader(http.StatusOK)
	if e = json.NewEncoder(w).Encode(volumes); e != nil {
		panic(e)
	}
}

func (c *VolumeListCommand) Routes() rest.Routes {
	return rest.Routes{
		// VolumeList
		rest.Route{
			Name:        "VolumeList",
			Method:      "GET",
			Pattern:     "/volumes/",
			HandlerFunc: c.VolumeList},
	}
}
