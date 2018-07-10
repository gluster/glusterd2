package quota

// SetLimitReq represents REST API request to Limit Usage/objects of a directory
type SetLimitReq struct {
	Path             string `json:"path"`
	SizeUsageLimit   string `json:"size-usage-limit,omitempty"`
	SoftUsagePercent string `json:"soft-usage-percent,omitempty"`
	ObjectCountLimit string `json:"object-count-limit,omitempty"`
}

// RemoveLimitReq represents REST API request to Remove Usage/objects of a directory
type RemoveLimitReq struct {
	Path string `json:"path"`
}

// Limit is the static information about the limits
type Limit struct {
	Hlbytes   int64
	Slpercent int64
}
