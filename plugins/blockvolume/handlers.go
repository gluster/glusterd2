package blockvolume

import (
	"github.com/gorilla/mux"
	"net/http"

	"github.com/gluster/glusterd2/glusterd2/servers/rest/utils"
	"github.com/gluster/glusterd2/plugins/blockvolume/api"
	"github.com/gluster/glusterd2/plugins/blockvolume/blockprovider"
)

// CreateVolume is a http Handler for creating a block volume
func (b *BlockVolume) CreateVolume(w http.ResponseWriter, r *http.Request) {
	var (
		req  = &api.BlockVolumeCreateRequest{}
		resp = &api.BlockVolumeCreateResp{}
		opts = []blockprovider.BlockVolOption{}
	)

	if err := utils.UnmarshalRequest(r, req); err != nil {
		utils.SendHTTPError(r.Context(), w, http.StatusBadRequest, err)
		return
	}

	opts = append(opts,
		blockprovider.WithHostVolume(req.HostingVolume),
		blockprovider.WithHaCount(req.HaCount),
	)

	if req.Auth {
		opts = append(opts, blockprovider.WithAuthEnabled)
	}

	blockVol, err := b.blockProvider.CreateBlockVolume(req.Name, req.Size, req.Clusters, opts...)
	if err != nil {
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
	var (
		pathParams = mux.Vars(r)
	)

	if err := b.blockProvider.DeleteBlockVolume(pathParams["name"]); err != nil {
		utils.SendHTTPError(r.Context(), w, http.StatusInternalServerError, err)
		return
	}

	utils.SendHTTPResponse(r.Context(), w, http.StatusNoContent, nil)
}

// ListBlockVolumes is a http handler for listing all available block volumes
func (b *BlockVolume) ListBlockVolumes(w http.ResponseWriter, r *http.Request) {
	var (
		resp = api.BlockVolumeListResp{}
	)

	blockVols := b.blockProvider.BlockVolumes()

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

	blockVol, err := b.blockProvider.GetBlockVolume(pathParams["name"])
	if err != nil {
		utils.SendHTTPError(r.Context(), w, http.StatusInternalServerError, err)
		return
	}

	{
		resp.BlockVolumeInfo = &api.BlockVolumeInfo{}
		resp.Name = blockVol.Name()
		resp.HostingVolume = blockVol.HostVolume()
		resp.Size = int64(blockVol.Size())
		resp.Hosts = blockVol.HostAddresses()
		resp.Password = blockVol.Password()
		resp.GBID = blockVol.ID()
		resp.HaCount = blockVol.HaCount()
	}

	utils.SendHTTPResponse(r.Context(), w, http.StatusOK, resp)
}
