package api

// BrickProfileInfo holds profile info of each brick
type BrickProfileInfo struct {
	BrickName       string   `json:"brick-name"`
	CumulativeStats StatType `json:"cumulative-stats"`
	IntervalStats   StatType `json:"interval-stats"`
}

// StatType contains profile info of cumulative/interval stats of a brick
type StatType struct {
	Duration             string                       `json:"duration"`
	DataRead             string                       `json:"data-read"`
	DataWrite            string                       `json:"data-write"`
	Interval             string                       `json:"interval"`
	PercentageAvgLatency float64                      `json:"percentage-avg-latency"`
	StatsInfo            map[string]map[string]string `json:"stat-info,omitempty"`
}
