package elasticetcd

import (
	"errors"
	"time"

	"github.com/coreos/etcd/clientv3"
	"github.com/coreos/etcd/clientv3/concurrency"
)

// Client returns the etcd client of ElasticEtcd
func (ee *ElasticEtcd) Client() *clientv3.Client {
	ee.lock.RLock()
	defer ee.lock.RUnlock()
	return ee.cli
}

// Session returns the etcd session used by ElasticEtcd
func (ee *ElasticEtcd) Session() *concurrency.Session {
	ee.lock.RLock()
	defer ee.lock.RUnlock()
	return ee.session
}

func (ee *ElasticEtcd) startClient() error {
	if ee.cli != nil {
		return errors.New("client already exists")
	}

	cli, err := clientv3.New(ee.newClientConfig())
	if err != nil {
		return err
	}

	ee.cli = cli
	ee.cli.Sync(ee.cli.Ctx())

	session, err := concurrency.NewSession(ee.cli)
	if err != nil {
		ee.cli.Close()
		return err
	}
	ee.session = session

	return nil
}

func (ee *ElasticEtcd) stopClient() error {
	if ee.cli == nil {
		return errors.New("no client present")
	}

	close(ee.stopwatching)
	ee.watchers.Wait()

	if err := ee.session.Close(); err != nil {
		return err
	}

	if err := ee.cli.Close(); err != nil {
		return err
	}

	ee.cli = nil
	return nil
}

func (ee *ElasticEtcd) newClientConfig() clientv3.Config {
	return clientv3.Config{
		Endpoints:        ee.conf.Endpoints.StringSlice(),
		AutoSyncInterval: 30 * time.Second,
		DialTimeout:      10 * time.Second,
	}
}

func (ee *ElasticEtcd) watch(key string, handler func(clientv3.WatchResponse), watchopts ...clientv3.OpOption) {
	ee.watchers.Add(1)
	go func() {
		defer ee.watchers.Done()

		wch := ee.cli.Watch(ee.cli.Ctx(), key, watchopts...)
		for {
			select {
			case resp := <-wch:
				if resp.Canceled {
					return
				}
				handler(resp)
			case <-ee.stopwatching:
				return
			}
		}
	}()
}
