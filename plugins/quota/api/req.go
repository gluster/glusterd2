package api

// SetLimitReq represents REST API request to Limit Usage/objects of a directory
type SetLimitReq struct {
	Path             string `json:"path"`
	SizeUsageLimit   int    `json:"size-usage-limit,omitempty"`
	ObjectCountLimit int    `json:"object-count-limit,omitempty"`
}

// RemoveLimitReq represents REST API request to Remove Usage/objects of a directory
type RemoveLimitReq struct {
	Path string `json:"path"`
}
