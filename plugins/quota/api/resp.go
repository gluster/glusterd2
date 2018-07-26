package api

type crawlInfo struct {
	CrawlPid int `json:"crawl-pid"`
	MountPid int `json:"crawl-mount-pid"`
}

type list struct {
	Path      string `json:"path"`
	HardLimit int64  `json:"hard-limit"`
	SoftLimit int64  `json:"soft-limit"`
	Used      int64  `json:"used"`
	Available int64  `json:"available"`
	/*
	 * TODO / Review: "Merge soft limit exceeded and hard limit
	 * exceeded" flags into a single field called status.
	 */
	SoftLimitExceeded bool `json:"soft-limit-exceeded"`
	HardLimitExceeded bool `json:"hard-limit-exceeded"`
	// LimitType indicates directory quota/inode-quota/...
	LimitType int32 `json:"limit-type"`
}

//ListResp is an array of structs representing individual limits.
type ListResp []list

//DisableResp gives the information of disable crawler on success
type DisableResp crawlInfo

//EnableResp gives the information of enable crawler on success
type EnableResp crawlInfo
