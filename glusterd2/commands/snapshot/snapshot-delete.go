package snapshotcommands

import (
	"net/http"
	"os"
	"strings"

	"github.com/gluster/glusterd2/glusterd2/gdctx"
	restutils "github.com/gluster/glusterd2/glusterd2/servers/rest/utils"
	"github.com/gluster/glusterd2/glusterd2/snapshot"
	"github.com/gluster/glusterd2/glusterd2/snapshot/lvm"
	"github.com/gluster/glusterd2/glusterd2/transaction"
	"github.com/gluster/glusterd2/glusterd2/volume"

	"github.com/gorilla/mux"
	"github.com/pborman/uuid"
)

func snapshotDelete(c transaction.TxnCtx) error {

	var snapinfo snapshot.Snapinfo
	if err := c.Get("snapinfo", &snapinfo); err != nil {
		return err
	}

	snapVol := snapinfo.SnapVolinfo
	for _, b := range snapVol.GetLocalBricks() {
		if snapVol.State == volume.VolStarted {
			if err := snapshot.StopBrick(b, c.Logger()); err != nil {
				c.Logger().WithError(err).WithField(
					"brick", b.Path).Warning("Failed to cleanup the brick.Earlier it might have stopped abnormaly")

			}
		}
		if err := lvm.RemoveBrickSnapshot(b.MountInfo.DevicePath); err != nil {
			c.Logger().WithError(err).WithField(
				"brick", b.Path).Debug("Failed to remove snapshotted LVM")
			return err
		}
		mountRoot := strings.TrimSuffix(b.Path, b.MountInfo.Mountdir)
		os.RemoveAll(mountRoot)
	}

	//TODO Delete the volfiles in etcd.
	return nil
}

func snapshotDeleteStore(c transaction.TxnCtx) error {

	var snapinfo snapshot.Snapinfo
	if err := c.Get("snapinfo", &snapinfo); err != nil {
		return err
	}

	return snapshot.DeleteSnapshot(&snapinfo)
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
		sf   transaction.StepFunc
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
		transaction.RegisterStepFunc(sf.sf, sf.name)
	}
}

func snapshotDeleteHandler(w http.ResponseWriter, r *http.Request) {

	ctx := r.Context()
	logger := gdctx.GetReqLogger(ctx)
	snapname := mux.Vars(r)["snapname"]
	//Fetching snapinfo to get the parent volume name. Parent volume has to be locked
	snapinfo, err := snapshot.GetSnapshot(snapname)
	if err != nil {
		status, err := restutils.ErrToStatusCode(err)
		restutils.SendHTTPError(ctx, w, status, err)
		return
	}

	txn, err := transaction.NewTxnWithLocks(ctx, snapname, snapinfo.ParentVolume)
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
	txn.Steps = []*transaction.Step{
		{
			DoFunc: "snap-delete.Commit",
			Nodes:  volinfo.Nodes(),
		},
		{
			DoFunc: "snap-delete.Store",
			Nodes:  []uuid.UUID{gdctx.MyUUID},
		},
	}

	if err := txn.Ctx.Set("snapinfo", snapinfo); err != nil {
		restutils.SendHTTPError(ctx, w, http.StatusInternalServerError, err)
		return
	}

	if err := txn.Do(); err != nil {
		logger.WithError(err).WithField(
			"snapname", snapname).Error("transaction to delete snapshot failed")
		restutils.SendHTTPError(ctx, w, http.StatusInternalServerError, err)
		return
	}

	logger.WithField("Snapshot-name", snapname).Info("snapshot deleted")

	restutils.SendHTTPResponse(ctx, w, http.StatusNoContent, "Snapshot Deleted")
}
