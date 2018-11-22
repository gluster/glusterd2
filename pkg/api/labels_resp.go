package api

// LabelInfo contains static information of a label
type LabelInfo struct {
	Name             string   `json:"labelname"`
	SnapMaxHardLimit uint64   `json:"snap-max-hard-limit"`
	SnapMaxSoftLimit uint64   `json:"snap-max-soft-limit"`
	ActivateOnCreate bool     `json:"activate-on-create"`
	AutoDelete       bool     `json:"auto-delete"`
	Description      string   `json:"description"`
	SnapList         []string `json:"snap-list"`
}

//LabelCreateResp is the response sent for a label get request.
type LabelCreateResp LabelInfo

//LabelGetResp is the response sent for a label get request.
type LabelGetResp LabelInfo

//LabelListResp is the response sent for a label list request.
type LabelListResp []LabelGetResp

//LabelConfigResp is the response sent for a label config request.
type LabelConfigResp LabelInfo
