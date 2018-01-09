package volgen

type staticGraph struct {
	name    string
	content string
}

var defaultGraphs = []staticGraph{
	{
		name: "brick.graph",
		content: `protocol/server
performance/decompounder, {{ brick.path }}
debug/io-stats
features/quota
features/index
features/barrier
features/marker
performance/io-threads
features/upcall
features/leases
features/read-only
features/worm
features/locks
features/access-control
features/bitrot-stub
features/changelog
features/changetimerecorder
features/trash
storage/posix`,
	},
	{
		name: "distreplicate.graph",
		content: `cluster/distribute
cluster/replicate
protocol/client`,
	},
	{
		name: "fuse.graph",
		content: `debug/io-stats
performance/io-threads
performance/md-cache
performance/open-behind
performance/quick-read
performance/io-cache
performance/readdir-ahead
performance/read-ahead
performance/write-behind
cluster.graph`,
	},
	{
		name: "distribute.graph",
		content: `cluster/dht
protocol/client`,
	},
	{
		name: "replicate.graph",
		content: `cluster/distribute
cluster/replicate
protocol/client`,
	},
	{
		name: "disperse.graph",
		content: `cluster/distribute
cluster/disperse
protocol/client`,
	},
}
