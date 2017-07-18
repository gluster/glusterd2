package elasticetcd

import (
	"time"

	"github.com/coreos/etcd/clientv3"
)

// volunteerSelf adds the self to the volunteer list and starts watching for the nomination
func (ee *ElasticEtcd) volunteerSelf() error {
	key := volunteerPrefix + ee.conf.Name
	var val string
	if isDefaultCURL(ee.conf.CURLs) {
		val = defaultACURLs.String()
	} else {
		val = ee.conf.CURLs.String()
	}

	_, err := ee.cli.Put(ee.cli.Ctx(), key, val, clientv3.WithLease(ee.session.Lease()))
	if err != nil {
		ee.log.WithError(err).Error("failed to add self to volunteer list")
		return err
	}
	ee.watchNomination()

	return nil
}

func (ee *ElasticEtcd) watchNomination() {
	ee.log.Debug("watching for self nomination")

	key := nomineePrefix + ee.conf.Name

	f := func(_ clientv3.WatchResponse) {
		ee.handleNomination()
	}
	ee.watch(key, f)
}

func (ee *ElasticEtcd) handleNomination() {
	ee.lock.Lock()
	defer ee.lock.Unlock()

	ee.log.Debug("handling nomination")

	nomineesResp, err := ee.cli.Get(ee.cli.Ctx(), nomineePrefix, clientv3.WithPrefix(), clientv3.WithSort(clientv3.SortByKey, clientv3.SortAscend))
	if err != nil {
		ee.log.WithError(err).Error("could not get nominees")
		return
	}

	nominees, err := urlsMapFromGetResp(nomineesResp, nomineePrefix)
	if err != nil {
		ee.log.WithError(err).Error("could not prepare nominees map")
	}

	if _, ok := nominees[ee.conf.Name]; ok {
		ee.log.Debug("nominated, starting server")
		// Sleeping to allow leader to add me as a etcd cluster member
		time.Sleep(2 * time.Second)
		ee.startServer(nominees.String())
	} else {
		ee.log.Debug("not nominated or nomination removed, stopping server")
		ee.stopServer()
	}
}
