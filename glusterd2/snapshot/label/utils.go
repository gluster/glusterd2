package label

import (
	"github.com/gluster/glusterd2/pkg/api"
)

//CreateLabelInfoResp parses volume information for response
func CreateLabelInfoResp(info *Info) *api.LabelInfo {

	resp := &api.LabelInfo{
		Name:             info.Name,
		SnapMaxHardLimit: info.SnapMaxHardLimit,
		SnapMaxSoftLimit: info.SnapMaxSoftLimit,
		ActivateOnCreate: info.ActivateOnCreate,
		AutoDelete:       info.AutoDelete,
		Description:      info.Description,
		SnapList:         info.SnapList,
	}
	return resp
}
