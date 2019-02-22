package snapshotcommands

import (
	"net/http"
	"os"
	"strings"
	"sync"

	"github.com/gluster/glusterd2/glusterd2/brick"
	"github.com/gluster/glusterd2/glusterd2/gdctx"
	"github.com/gluster/glusterd2/glusterd2/oldtransaction"
	restutils "github.com/gluster/glusterd2/glusterd2/servers/rest/utils"
	"github.com/gluster/glusterd2/glusterd2/snapshot"
	"github.com/gluster/glusterd2/glusterd2/volgen"
	"github.com/gluster/glusterd2/glusterd2/volume"
	"github.com/gluster/glusterd2/pkg/lvmutils"
	log "github.com/sirupsen/logrus"

	"github.com/gorilla/mux"
	"github.com/pborman/uuid"
	"go.opencensus.io/trace"
)

func snapshotBrickDelete(errCh chan error, wg *sync.WaitGroup, snapVol volume.Volinfo, b brick.Brickinfo, logger log.FieldLogger) {
	defer wg.Done()

	if snapVol.State == volume.VolStarted {
		if err := volume.StopBrick(b, logger); err != nil {
			log.WithError(err).WithField(
				"brick", b.Path).Warning("Failed to cleanup the brick.Earlier it might have stopped abnormally")

		}
	}
	if err := lvmutils.RemoveLVSnapshot(b.MountInfo.DevicePath); err != nil {
		log.WithError(err).WithField(
			"brick", b.Path).Debug("Failed to remove snapshotted LVM")
		errCh <- err
		return
	}
	mountRoot := strings.TrimSuffix(b.Path, b.MountInfo.BrickDirSuffix)
	os.RemoveAll(mountRoot)

	volfileID := brick.GetVolfileID(snapVol.Name, b.Path)
	if err := volgen.DeleteFile(volfileID); err != nil {
		errCh <- err
	}

	errCh <- nil
	return
}

func snapshotDelete(c oldtransaction.TxnCtx) error {
	var snapinfo snapshot.Snapinfo
	if err := c.Get("snapinfo", &snapinfo); err != nil {
		return err
	}

	snapVol := snapinfo.SnapVolinfo
	var wg sync.WaitGroup
	numBricks := len(snapVol.GetBricks())
	errCh := make(chan error, numBricks)

	for _, b := range snapVol.GetLocalBricks() {
		wg.Add(1)
		go snapshotBrickDelete(errCh, &wg, snapVol, b, c.Logger())
	}

	err := error(nil)
	go func() {
		for i := range errCh {
			if i != nil && err == nil {
				//Return the first error from goroutines
				err = i
			}
		}
	}()
	wg.Wait()
	close(errCh)

	//TODO Delete the volfiles in etcd.
	return err
}

func snapshotDeleteStore(c oldtransaction.TxnCtx) error {

	var (
		snapinfo snapshot.Snapinfo
		err      error
	)
	if err = c.Get("snapinfo", &snapinfo); err != nil {
		return err
	}

	err = snapshot.DeleteSnapshot(&snapinfo)
	return err
}

/*
TODO
How do we do rollbacking of lvremove command?
We can create snapshot of snapshot during validation as a backup
and remove the same if everything succeeded or rollback to that
snapshot if something fails.
*/
func registerSnapDeleteStepFuncs() {
	var sfs = []struct {
		name string
		sf   oldtransaction.StepFunc
	}{
		{"snap-delete.Commit", snapshotDelete},
		{"snap-delete.Store", snapshotDeleteStore},
		/*
			TODO
			{"snap-delete.UndoStore", undoSnapshotDeleteStore},
			{"snap-delete.UndoCommit", undoSnapshotDelete},
		*/
	}
	for _, sf := range sfs {
		oldtransaction.RegisterStepFunc(sf.sf, sf.name)
	}
}

func snapshotDeleteHandler(w http.ResponseWriter, r *http.Request) {

	ctx := r.Context()
	ctx, span := trace.StartSpan(ctx, "/snapshotDeleteHandler")
	defer span.End()

	logger := gdctx.GetReqLogger(ctx)
	snapname := mux.Vars(r)["snapname"]
	//Fetching snapinfo to get the parent volume name. Parent volume has to be locked
	snapinfo, err := snapshot.GetSnapshot(snapname)
	if err != nil {
		status, err := restutils.ErrToStatusCode(err)
		restutils.SendHTTPError(ctx, w, status, err)
		return
	}

	txn, err := oldtransaction.NewTxnWithLocks(ctx, snapname, snapinfo.ParentVolume)
	if err != nil {
		status, err := restutils.ErrToStatusCode(err)
		restutils.SendHTTPError(ctx, w, status, err)
		return
	}
	defer txn.Done()

	//Fetching snapinfo again, but this time inside a lock
	snapinfo, err = snapshot.GetSnapshot(snapname)
	if err != nil {
		status, err := restutils.ErrToStatusCode(err)
		restutils.SendHTTPError(ctx, w, status, err)
		return
	}

	volinfo := &snapinfo.SnapVolinfo
	txn.Steps = []*oldtransaction.Step{
		{
			DoFunc: "snap-delete.Commit",
			Nodes:  volinfo.Nodes(),
		},
		{
			DoFunc: "snap-delete.Store",
			Nodes:  []uuid.UUID{gdctx.MyUUID},
			Sync:   true,
		},
	}

	if err := txn.Ctx.Set("snapinfo", snapinfo); err != nil {
		restutils.SendHTTPError(ctx, w, http.StatusInternalServerError, err)
		return
	}

	span.AddAttributes(
		trace.StringAttribute("reqID", txn.Ctx.GetTxnReqID()),
		trace.StringAttribute("snapName", snapname),
		trace.StringAttribute("parentVolume", snapinfo.ParentVolume),
	)

	if err := txn.Do(); err != nil {
		logger.WithError(err).WithField(
			"snapname", snapname).Error("transaction to delete snapshot failed")
		restutils.SendHTTPError(ctx, w, http.StatusInternalServerError, err)
		return
	}

	logger.WithField("Snapshot-name", snapname).Info("snapshot deleted")

	restutils.SendHTTPResponse(ctx, w, http.StatusNoContent, nil)
}
