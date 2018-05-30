package glustershd

import (
	"bytes"
	"encoding/xml"
	"fmt"
	"net/http"
	"os/exec"
	"path"
	"strings"

	"github.com/gluster/glusterd2/glusterd2/gdctx"
	restutils "github.com/gluster/glusterd2/glusterd2/servers/rest/utils"
	"github.com/gluster/glusterd2/glusterd2/transaction"
	"github.com/gluster/glusterd2/glusterd2/volume"
	gderrors "github.com/gluster/glusterd2/pkg/errors"
	glustershdapi "github.com/gluster/glusterd2/plugins/glustershd/api"

	"github.com/gorilla/mux"
	config "github.com/spf13/viper"
)

func runGlfshealBin(volname string, args []string) (string, error) {
	var out bytes.Buffer
	var buffer bytes.Buffer
	var healInfoOutput string

	buffer.WriteString(fmt.Sprintf("%s", volname))
	for _, arg := range args {
		buffer.WriteString(fmt.Sprintf(" %s", arg))
	}

	args = strings.Fields(buffer.String())
	path, err := exec.LookPath("glfsheal")
	if err != nil {
		return healInfoOutput, err
	}

	cmd := exec.Command(path, args...)
	cmd.Stdout = &out
	err = cmd.Run()
	if err != nil {
		return healInfoOutput, err
	}

	healInfoOutput = out.String()

	return healInfoOutput, nil
}

func getHealInfo(volname string, option string) (string, error) {
	var options []string
	glusterdSockpath := path.Join(config.GetString("rundir"), "glusterd2.socket")
	options = append(options, option, "xml", "glusterd-sock", glusterdSockpath)

	return runGlfshealBin(volname, options)
}

func selfhealInfoHandler(w http.ResponseWriter, r *http.Request) {
	var option string
	p := mux.Vars(r)
	volname := p["name"]
	if val, ok := p["opts"]; ok {
		option = val
	}

	ctx := r.Context()
	logger := gdctx.GetReqLogger(ctx)

	lock, unlock := transaction.CreateLockFuncs(volname)
	if err := lock(ctx); err != nil {
		if err == transaction.ErrLockTimeout {
			restutils.SendHTTPError(ctx, w, http.StatusConflict, err)
		} else {
			restutils.SendHTTPError(ctx, w, http.StatusInternalServerError, err)
		}
		return
	}
	defer unlock(ctx)

	// Validate volume existence
	volinfo, err := volume.GetVolume(volname)
	if err != nil {
		if err == gderrors.ErrVolNotFound {
			logger.WithError(err).WithField(
				"volname", volname).Debug("volume not found")
			restutils.SendHTTPError(ctx, w, http.StatusNotFound, err)
		} else {
			logger.WithError(err).WithField(
				"volname", volname).Debug("error occurred while looking for volume")
			restutils.SendHTTPError(ctx, w, http.StatusInternalServerError, err)
		}
		return
	}

	// Validate volume type
	if !isVolReplicate(volinfo.Type) {
		restutils.SendHTTPError(ctx, w, http.StatusBadRequest, "invalid operation for this volume type")
		return
	}

	// Validate volume state
	if volinfo.State != volume.VolStarted {
		restutils.SendHTTPError(ctx, w, http.StatusBadRequest, gderrors.ErrVolNotStarted)
		return
	}
	healInfoOutput, err := getHealInfo(volname, option)
	if err != nil {
		logger.WithField("volname", volname).Debug("heal info operation failed")
		restutils.SendHTTPError(ctx, w, http.StatusInternalServerError, "heal info operation failed")
		return
	}

	output := []byte(healInfoOutput)

	var info glustershdapi.HealInfo
	err = xml.Unmarshal(output, &info)
	if err != nil {
		logger.WithError(err).Error("Error unmarshalling XML output from heal info command")
		restutils.SendHTTPError(ctx, w, http.StatusInternalServerError, err)
		return
	}

	restutils.SendHTTPResponse(ctx, w, http.StatusOK, &info.Bricks)

}
