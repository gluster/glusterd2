package api

// LabelCreateReq represents a lebel Create Request
type LabelCreateReq struct {
	Name             string `json:"labelname"`
	SnapMaxHardLimit uint64 `json:"snap-max-hard-limit"`
	SnapMaxSoftLimit uint64 `json:"snap-max-soft-limit"`
	ActivateOnCreate bool   `json:"activate-on-create,omitempty"`
	AutoDelete       bool   `json:"auto-delete,omitempty"`
	Description      string `json:"description,omitempty"`
}

// LabelSetReq represents a lebel Create Request
type LabelSetReq struct {
	Configurations map[string]string `json:"configurations"`
}

// LabelResetReq represents a lebel Create Request
type LabelResetReq struct {
	Configurations []string `json:"configurations"`
}
