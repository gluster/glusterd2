package api

// VersionResp is the response for request sent to /version endpoint.
type VersionResp struct {
	GlusterdVersion string `json:"glusterd-version"`
	APIVersion      int    `json:"api-version"`
}
