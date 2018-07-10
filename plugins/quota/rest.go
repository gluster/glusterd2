package quota

import (
	"bytes"
	"context"
	"encoding/binary"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"os/exec"
	"path"
	"strconv"
	"strings"
	"syscall"
	"unsafe"

	"github.com/gluster/glusterd2/glusterd2/gdctx"
	restutils "github.com/gluster/glusterd2/glusterd2/servers/rest/utils"
	"github.com/gluster/glusterd2/glusterd2/store"
	"github.com/gluster/glusterd2/glusterd2/transaction"
	"github.com/gluster/glusterd2/glusterd2/volume"
	"github.com/gluster/glusterd2/pkg/errors"
	quotaapi "github.com/gluster/glusterd2/plugins/quota/api"

	"github.com/gorilla/mux"
	"github.com/pborman/uuid"
	log "github.com/sirupsen/logrus"
	config "github.com/spf13/viper"
	"golang.org/x/sys/unix"
)

const (
	quotaConfPrefix string = "quota/"
)

func quotaListHandler(w http.ResponseWriter, r *http.Request) {
	// implement the help logic and send response back as below
	ctx := r.Context()
	restutils.SendHTTPResponse(ctx, w, http.StatusOK, "todo: quota list")
}

func auxMount(volname string, mountdir string) error {

	pidfile := fmt.Sprintf("%s/gluster/%s_quota_limit.pid", config.GetString("rundir"), volname)

	logfiledir := path.Join(
		config.GetString("logdir"), "glusterfs",
	)
	logfile := fmt.Sprintf("%s/quota-mount-%s.log", logfiledir, volname)

	var buffer bytes.Buffer
	buffer.WriteString(fmt.Sprintf(" --volfile-server localhost"))
	buffer.WriteString(fmt.Sprintf(" --volfile-id %s", volname))
	buffer.WriteString(fmt.Sprintf(" --log-file %s", logfile))
	buffer.WriteString(fmt.Sprintf(" -p %s", pidfile))
	buffer.WriteString(" --client-pid -5 ")
	buffer.WriteString(mountdir)

	args := strings.Fields(buffer.String())
	cmd := exec.Command("glusterfs", args...)
	if err := cmd.Start(); err != nil {
		log.WithError(err).WithFields(log.Fields{
			"logfile":  logfile,
			"pidfile":  pidfile,
			"mountdir": mountdir,
			"volname":  volname}).Error("failed to start aux mount")
		return err
	}

	return cmd.Wait() // glusterfs daemonizes itself
}

//quotaAuxMount is the function to create an auxillary mount
func quotaAuxMount(c transaction.TxnCtx) error {

	var volname string
	if err := c.Get("volname", &volname); err != nil {
		c.Logger().WithError(err).WithField(
			"key", "volname").Error("failed to get volname from context")
		return err
	}

	mountpath := fmt.Sprintf("%s/gluster/%s_quota_limit/", config.GetString("rundir"), volname)

	// Create logFiledir dir
	if err := os.MkdirAll(path.Dir(mountpath),
		os.ModeDir|os.ModePerm); err != nil {
		return err
	}

	// Create auxillary mount through which limit/crawl is excuted
	if err := auxMount(volname, mountpath); err != nil {
		c.Logger().WithError(err).Error("aux mount failed")
		return err
	}
	return nil
}

//limitSet is the function to set/change quota limit
func limitSet(c transaction.TxnCtx) error {

	var parentStat unix.Stat_t
	var volname string
	var path string
	var hardLimit string
	var softLimit string
	var limit quotaapi.Limit
	var newlimit quotaapi.Limit
	var structSize = unsafe.Sizeof(quotaapi.Limit{})

	if err := c.Get("volname", &volname); err != nil {
		c.Logger().WithError(err).WithField(
			"key", "volname").Error("failed to get volname from context")
		return err
	}

	if err := c.Get("path", &path); err != nil {
		c.Logger().WithError(err).WithField(
			"key", "volname").Error("failed to get paath from context")
		return err
	}

	if err := c.Get("hard-limit", &hardLimit); err != nil {
		c.Logger().WithError(err).WithField(
			"key", "volname").Error("failed to get hard-limit from context")
		return err
	}

	if err := c.Get("soft-limit", &softLimit); err != nil {
		c.Logger().WithError(err).WithField(
			"key", "volname").Error("failed to get value for key from context")
	}

	mountdir := fmt.Sprintf("%s/gluster/%s_quota_limit/", config.GetString("rundir"), volname)
	mountpath := fmt.Sprintf("%s/%s", mountdir, path)

	getbytes := make([]byte, structSize)
	if strings.Compare(softLimit, "") != 0 {
		newsoftlimit, err := strconv.ParseInt(softLimit, 10, 64)
		c.Logger().WithError(err).Error("failed to convert softlimit")
		newlimit.Slpercent = newsoftlimit
	} else {
		if err := unix.Lstat(mountpath, &parentStat); err != nil {
			c.Logger().WithError(err).Error("lstat failed")
			return err
		}
		_, err := unix.Getxattr(mountpath, "trusted.glusterfs.quota.limit-set", getbytes)
		if err != nil {
			c.Logger().WithError(err).Error("getxattr failed")
			if err != syscall.ENODATA {
				return err
			}
		}

		getbuf := bytes.NewBuffer(getbytes)
		if err := binary.Read(getbuf, binary.BigEndian, &limit); err != nil {
			c.Logger().WithError(err).Error("conversion of limit from bytes failed")
			return err
		}
		newlimit.Slpercent = limit.Slpercent
	}

	newhardlimit, err := strconv.ParseInt(hardLimit, 10, 64)
	newlimit.Hlbytes = newhardlimit

	setbuf := bytes.NewBuffer(make([]byte, 0, structSize))
	if err := binary.Write(setbuf, binary.BigEndian, newlimit); err != nil {
		c.Logger().WithError(err).Error("conversion to bytes for limit failed")
		return err
	}
	err = unix.Setxattr(mountpath, "trusted.glusterfs.quota.limit-set", setbuf.Bytes(), 0)
	if err != nil {
		c.Logger().WithError(err).WithFields(log.Fields{
			"path": path,
			"key":  "trusted.glusterfs.quota.limit-set"}).Error("Setxattr failed")
		return err
	}
	resp, err := store.Store.Get(context.TODO(), quotaConfPrefix+volname)
	if err != nil {
		c.Logger().WithError(err).WithFields(log.Fields{
			"path":      path,
			"hardlimit": hardLimit}).Error("failed to get quota conf from store")
		return err
	}

	var gfids []string
	if resp.Count != 1 {
		c.Logger().WithFields(log.Fields{
			"path":      path,
			"hardlimit": hardLimit}).Error("store returned empty")
		goto store
	}

	if err := json.Unmarshal(resp.Kvs[0].Value, &gfids); err != nil {
		c.Logger().WithError(err).WithFields(log.Fields{
			"path":      path,
			"hardlimit": hardLimit}).Error("failed to unmarshal")
		return err
	}

store:
	gfids = append(gfids, path)
	gfidsJSON, err := json.Marshal(gfids)

	_, err = store.Store.Put(context.TODO(), quotaConfPrefix+volname, string(gfidsJSON))
	if err != nil {
		c.Logger().WithError(err).WithFields(log.Fields{
			"path":      path,
			"hardlimit": hardLimit}).Error("failed to store in quota conf")
		return err
	}

	defer syscall.Unmount(mountpath, syscall.MNT_FORCE)
	// does the gf_umount_lazy after the execution of the command
	return nil
}

func quotaLimitHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	logger := gdctx.GetReqLogger(ctx)

	// Collect inputs from URL
	p := mux.Vars(r)
	volName := p["volname"]

	txn, err := transaction.NewTxnWithLocks(ctx, volName)
	if err != nil {
		status, err := restutils.ErrToStatusCode(err)
		restutils.SendHTTPError(ctx, w, status, err)
		return
	}
	defer txn.Done()

	// Parse the JSON body to get additional details of request
	var req quotaapi.SetLimitReq
	if err := restutils.UnmarshalRequest(r, &req); err != nil {
		restutils.SendHTTPError(ctx, w, http.StatusUnprocessableEntity, errors.ErrJSONParsingFailed.Error())
		return
	}

	if req.Path == "" {
		restutils.SendHTTPError(ctx, w, http.StatusBadRequest, errors.ErrInvalidPath.Error())
		logger.Error("no path found")
		return
	}
	if req.SizeUsageLimit == "" {
		restutils.SendHTTPError(ctx, w, http.StatusBadRequest, errors.ErrInvalidSize.Error())
		logger.Error("no size found")
		return
	}
	vol, err := volume.GetVolume(volName)
	if err != nil {
		logger.Error("volume not found")
		restutils.SendHTTPError(ctx, w, http.StatusNotFound, errors.ErrVolNotFound.Error())
		return
	}
	// Check if volume is started
	if vol.State != volume.VolStarted {
		restutils.SendHTTPError(ctx, w, http.StatusBadRequest, errors.ErrVolNotStarted.Error())
		return
	}

	// Check if quota is enabled
	if !isQuotaEnabled(vol) {
		restutils.SendHTTPError(ctx, w, http.StatusBadRequest, errors.ErrQuotadNotEnabled.Error())
		return
	}

	err = txn.Ctx.Set("volname", volName)
	if err != nil {
		logger.WithError(err).Error("failed to set volinfo in transaction context")
		restutils.SendHTTPError(ctx, w, http.StatusInternalServerError, err.Error())
		return
	}

	err = txn.Ctx.Set("path", req.Path)
	if err != nil {
		logger.WithError(err).Error("failed to set path in transaction context")
		restutils.SendHTTPError(ctx, w, http.StatusInternalServerError, err.Error())
		return
	}

	err = txn.Ctx.Set("hard-limit", req.SizeUsageLimit)
	if err != nil {
		logger.WithError(err).Error("failed to set hard-limit in transaction context")
		restutils.SendHTTPError(ctx, w, http.StatusInternalServerError, err.Error())
		return
	}

	err = txn.Ctx.Set("soft-limit", req.SoftUsagePercent)

	txn.Steps = []*transaction.Step{
		{
			DoFunc: "quota-mount.aux",
			Nodes:  []uuid.UUID{gdctx.MyUUID},
		},
		{
			DoFunc: "quota-limit.set",
			Nodes:  vol.Nodes(),
		},
	}

	err = txn.Do()
	if err != nil {
		logger.WithError(err).Error("quota limit transaction failed")
		if err == transaction.ErrLockTimeout {
			restutils.SendHTTPError(ctx, w, http.StatusConflict, err.Error())
		} else {
			restutils.SendHTTPError(ctx, w, http.StatusInternalServerError, err.Error())
		}
		return
	}

	txn.Ctx.Logger().WithField("volname", vol.Name).Info("quota limit set")

	restutils.SendHTTPResponse(ctx, w, http.StatusOK, "limitusage successful")
}

func quotaRemoveHandler(w http.ResponseWriter, r *http.Request) {
	// implement the help logic and send response back as below
	ctx := r.Context()
	restutils.SendHTTPResponse(ctx, w, http.StatusOK, "Todo: quota Remove")
}
