// Package elasticetcd implements an autoscaling etcd cluster.
// This package was inspired by, and implements the elastic scaling algorithm as described in, the github.com/purpleidea/mgmt/etcd package.
//
// The elastic scaling algorithm is as follows,
// 	- The elasticetcd instance is started, and can be passed a list of existing etcd endpoints
//  	- If no endpoints are given, assume you are the first instance up and start your embedded etcd server
//	- Connect to the given endpoints (or your own etcd endpoints if you were the first one up)
//	- Volunteer to be a server, and wait to be nominated as a server
//		- If you are nominated, start your embedded etcd server and join the existing cluster
//		- If your nomination is removed, stop your embedded etcd server
//	- Begin a campaign to become the leader of the elastic cluster
// 		- When elected as the leader, make nominations from the volunteer list, to keep the right number of servers.
// 		- Watch for changes to the volunteer list, online servers and the ideal size, and make/remove nominations as required.
//
// Right now the server nominations are selected in a round-robin fashion, using the list of volunteers sorted by name.
//
// TODO: Figure out and implement recovery steps, for recovering from a complete cluster shutdown
//
// TODO: Add more and better logging throughout the package
//
// TODO: Allow the ability to select alternative selection algorithms
//
// TODO: Add functional tests
//
// TODO: Add rate limiting for nominations, trying to do many nominations at once will lead to a bad cluster
//
// TODO: Protect access to elastic namespace in etcd
package elasticetcd
