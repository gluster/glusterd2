package volumedelete

import (
	"fmt"
	"net/http"

	"github.com/kshlm/glusterd2/context"
	"github.com/kshlm/glusterd2/rest"

	log "github.com/Sirupsen/logrus"
	"github.com/gorilla/mux"
)

type VolumeDeleteCommand struct {
}

func (c *VolumeDeleteCommand) VolumeDelete(w http.ResponseWriter, r *http.Request) {
	p := mux.Vars(r)
	volname := p["volname"]

	log.Info("In Volume info API")

	e := context.Store.DeleteVolume(volname)
	if e != nil {
		http.Error(w, e.Error(), http.StatusInternalServerError)
		return
	}

	// Write nsg
	w.Header().Set("Content-Type", "text/plain; charset=UTF-8")
	w.WriteHeader(http.StatusOK)
	fmt.Fprint(w, "Volume deleted successfully")
}

func (c *VolumeDeleteCommand) Routes() rest.Routes {
	return rest.Routes{
		// VolumeDelete
		rest.Route{
			Name:        "VolumeDelete",
			Method:      "DELETE",
			Pattern:     "/volumes/{volname}",
			HandlerFunc: c.VolumeDelete},
	}
}
