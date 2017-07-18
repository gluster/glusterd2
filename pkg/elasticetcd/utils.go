package elasticetcd

import (
	"strings"

	"github.com/coreos/etcd/clientv3"
	"github.com/coreos/etcd/pkg/types"
)

func urlsMapFromGetResp(resp *clientv3.GetResponse, prefix string) (types.URLsMap, error) {
	urlsmap := make(types.URLsMap)
	for _, kv := range resp.Kvs {
		k := string(kv.Key)
		v := string(kv.Value)

		if prefix != "" {
			k = strings.TrimSpace(strings.TrimPrefix(k, prefix))
		}

		if k == "" {
			continue
		}

		urls, err := types.NewURLs(strings.Split(v, ","))
		if err != nil {
			return nil, err
		}
		urlsmap[k] = urls
	}
	return urlsmap, nil
}

func keysFromGetResp(resp *clientv3.GetResponse, prefix string) []string {
	var keys []string

	for _, kv := range resp.Kvs {
		k := string(kv.Key)

		if prefix != "" {
			k = strings.TrimSpace(strings.TrimPrefix(k, prefix))
		}

		if k == "" {
			continue
		}

		keys = append(keys, k)
	}

	return keys
}

// compareStringSlices compares two sorted slices
func compareStringSlices(a, b []string) bool {
	if len(a) != len(b) {
		return false
	}

	for i, v := range a {
		if v != b[i] {
			return false
		}
	}
	return true
}

// diffStringSlices returns a slice with items which are unique to the first
// Only works when the given slices are both sorted
func diffStringSlices(a, b []string) []string {
	var diff []string

	j := 0
	for _, v := range a {
		if j < len(b) && v == b[j] {
			j++
			continue
		}
		diff = append(diff, v)
	}

	return diff
}
