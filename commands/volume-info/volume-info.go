package volumeinfo

import (
	"encoding/json"
	"net/http"

	"github.com/kshlm/glusterd2/context"
	"github.com/kshlm/glusterd2/rest"

	log "github.com/Sirupsen/logrus"
	"github.com/gorilla/mux"
)

type VolumeInfoCommand struct {
}

func (c *VolumeInfoCommand) VolumeInfo(w http.ResponseWriter, r *http.Request) {
	p := mux.Vars(r)
	volname := p["volname"]

	log.Info("In Volume info API")

	vol, e := context.Store.GetVolume(volname)
	if e != nil {
		http.Error(w, e.Error(), http.StatusInternalServerError)
		return
	}

	// Write nsg
	w.Header().Set("Content-Type", "text/plain; charset=UTF-8")
	w.WriteHeader(http.StatusOK)
	if e = json.NewEncoder(w).Encode(vol); e != nil {
		panic(e)
	}
}

func (c *VolumeInfoCommand) SetRoutes(router *mux.Router) error {
	routes := rest.Routes{
		// VolumeInfo
		rest.Route{
			Name:        "VolumeInfo",
			Method:      "GET",
			Pattern:     "/volumes/{volname}",
			HandlerFunc: c.VolumeInfo},
	}
	// Register all routes
	for _, route := range routes {
		// Add routes from the table
		router.
			Methods(route.Method).
			Path(route.Pattern).
			Name(route.Name).
			Handler(route.HandlerFunc)
	}

	return nil

}
