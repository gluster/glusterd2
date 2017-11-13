// Package store implements the centralized store for GlusterD
//
// ETCD is used as the GlusterD store backend.
// The GlusterD store can work with an externally managed etcd cluster, or use an embedded etcd server.
// By default the embedded etcd is used.
//
// The embedded etcd server uses the
// github.com/gluster/glusterd2/pkg/elasticetcd package, which provides an
// autoscaling etcd cluster, and allows GD2 to be used without much difficulties.
// More details on how elasticetcd works can be found in its package documentation.
package store
