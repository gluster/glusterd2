package blockvolume

import (
	"net/http"

	"github.com/gluster/glusterd2/glusterd2/servers/rest/utils"
	"github.com/gluster/glusterd2/plugins/blockvolume/api"
	"github.com/gluster/glusterd2/plugins/blockvolume/blockprovider"

	"github.com/gorilla/mux"
)

// CreateVolume is a http Handler for creating a block volume
func (b *BlockVolume) CreateVolume(w http.ResponseWriter, r *http.Request) {
	var (
		req        = &api.BlockVolumeCreateRequest{}
		resp       = &api.BlockVolumeCreateResp{}
		opts       = []blockprovider.BlockVolOption{}
		pathParams = mux.Vars(r)
	)

	if err := utils.UnmarshalRequest(r, req); err != nil {
		utils.SendHTTPError(r.Context(), w, http.StatusBadRequest, err)
		return
	}

	opts = append(opts,
		blockprovider.WithHaCount(req.HaCount),
		blockprovider.WithHosts(req.Clusters),
		blockprovider.WithBlockType(req.BlockType),
	)

	if req.Auth {
		opts = append(opts, blockprovider.WithAuthEnabled)
	}

	blockProvider, err := blockprovider.GetBlockProvider(pathParams["provider"])
	if err != nil {
		utils.SendHTTPError(r.Context(), w, http.StatusInternalServerError, err)
		return
	}

	hostVolInfo, err := b.hostVolManager.GetOrCreateHostingVolume(req.HostingVolume, req.Name, req.Size, &req.HostVolumeInfo)
	if err != nil {
		utils.SendHTTPError(r.Context(), w, http.StatusInternalServerError, err)
		return
	}

	blockVol, err := blockProvider.CreateBlockVolume(req.Name, req.Size, hostVolInfo.Name, opts...)
	if err != nil {
		_ = b.hostVolManager.DeleteBlockInfoFromBHV(hostVolInfo.Name, req.Name, req.Size)
		utils.SendHTTPError(r.Context(), w, http.StatusInternalServerError, err)
		return
	}

	{
		resp.BlockVolumeInfo = req.BlockVolumeInfo
		resp.HostingVolume = blockVol.HostVolume()
		resp.Name = blockVol.Name()
		resp.Iqn = blockVol.IQN()
		resp.Username = blockVol.Username()
		resp.Password = blockVol.Password()
		resp.Hosts = blockVol.HostAddresses()
	}

	utils.SendHTTPResponse(r.Context(), w, http.StatusCreated, resp)
}

// DeleteVolume is a http Handler for deleting a specific block-volume
func (b *BlockVolume) DeleteVolume(w http.ResponseWriter, r *http.Request) {
	pathParams := mux.Vars(r)
	blockProvider, err := blockprovider.GetBlockProvider(pathParams["provider"])
	if err != nil {
		utils.SendHTTPError(r.Context(), w, http.StatusInternalServerError, err)
		return
	}

	blkVol, err := blockProvider.GetAndDeleteBlockVolume(pathParams["name"])
	if err != nil {
		utils.SendHTTPError(r.Context(), w, http.StatusInternalServerError, err)
		return
	}

	err = b.hostVolManager.DeleteBlockInfoFromBHV(blkVol.HostVolume(), pathParams["name"], blkVol.Size())
	if err != nil {
		utils.SendHTTPError(r.Context(), w, http.StatusInternalServerError, err)
		return
	}

	utils.SendHTTPResponse(r.Context(), w, http.StatusNoContent, nil)
}

// ListBlockVolumes is a http handler for listing all available block volumes
func (b *BlockVolume) ListBlockVolumes(w http.ResponseWriter, r *http.Request) {
	var (
		resp       = api.BlockVolumeListResp{}
		pathParams = mux.Vars(r)
	)

	blockProvider, err := blockprovider.GetBlockProvider(pathParams["provider"])
	if err != nil {
		utils.SendHTTPError(r.Context(), w, http.StatusInternalServerError, err)
		return
	}

	blockVols := blockProvider.BlockVolumes()

	for _, blockVol := range blockVols {
		resp = append(resp, api.BlockVolumeInfo{Name: blockVol.Name(), HostingVolume: blockVol.HostVolume()})
	}

	utils.SendHTTPResponse(r.Context(), w, http.StatusOK, resp)
}

// GetBlockVolume is a http Handler for getting info about a block volume.
func (b *BlockVolume) GetBlockVolume(w http.ResponseWriter, r *http.Request) {
	var (
		pathParams = mux.Vars(r)
		resp       = &api.BlockVolumeGetResp{}
	)

	blockProvider, err := blockprovider.GetBlockProvider(pathParams["provider"])
	if err != nil {
		utils.SendHTTPError(r.Context(), w, http.StatusInternalServerError, err)
		return
	}

	blockVol, err := blockProvider.GetBlockVolume(pathParams["name"])
	if err != nil {
		utils.SendHTTPError(r.Context(), w, http.StatusInternalServerError, err)
		return
	}

	{
		resp.BlockVolumeInfo = &api.BlockVolumeInfo{}
		resp.Name = blockVol.Name()
		resp.HostingVolume = blockVol.HostVolume()
		resp.Size = blockVol.Size()
		resp.Hosts = blockVol.HostAddresses()
		resp.Password = blockVol.Password()
		resp.GBID = blockVol.ID()
		resp.HaCount = blockVol.HaCount()
	}

	utils.SendHTTPResponse(r.Context(), w, http.StatusOK, resp)
}
