package api

// ScrubNodeInfo contains information about the scrub status for node
// Clients should NOT use this struct directly.
type ScrubNodeInfo struct {
	Node                   string   `json:"node"`
	ScrubRunning           string   `json:"scrub-running"`
	NumScrubbedFiles       string   `json:"num-scrubbed-files"`
	NumSkippedFiles        string   `json:"num-skipped-files"`
	LastScrubCompletedTime string   `json:"last-scrub-complete-time"`
	LastScrubDuration      string   `json:"last-scrub-duration"`
	ErrorCount             string   `json:"error-count"`
	CorruptedObjects       []string `json:"corrupted-objects"`
}

// ScrubStatus contains information about the scrub status for volume.
// Clients should NOT use this struct directly.
type ScrubStatus struct {
	Volume       string          `json:"volume"`
	State        string          `json:"state"`
	Throttle     string          `json:"throttle"`
	Frequency    string          `json:"frequency"`
	BitdLogFile  string          `json:"bitd-log-file"`
	ScrubLogFile string          `json:"scrub-log-file"`
	Nodes        []ScrubNodeInfo `json:"nodes"`
}
