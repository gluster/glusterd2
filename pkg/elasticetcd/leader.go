package elasticetcd

import (
	"context"
	"strconv"
	"time"

	"github.com/coreos/etcd/clientv3"
	"github.com/coreos/etcd/clientv3/concurrency"
	"github.com/coreos/etcd/etcdserver/etcdserverpb"
	"github.com/coreos/etcd/pkg/types"
)

const (
	eePrefix        = "elastic"
	electionKey     = eePrefix + "/election"
	volunteerPrefix = eePrefix + "/volunteers/"
	nomineePrefix   = eePrefix + "/nominees/"
	idealSizeKey    = eePrefix + "/idealSize"
)

func (ee *ElasticEtcd) startCampaign() {
	go func() {
		election := concurrency.NewElection(ee.session, electionKey)
		for {
			// Campaign for the becoming the leader
			// Stop campaigning if the context is canceled
			// else ignore errors and retry campaign
			ee.log.Debug("campaigning to become leader")
			err := election.Campaign(ee.cli.Ctx(), ee.conf.Name)
			if err != nil {
				switch {
				case err == context.Canceled:
					return
				default:
					continue
				}
			}
			ee.log.Debug("won election to become leader")
			// Resign as leader if you cannot start the leader funtions
			if err := ee.startLeader(); err != nil {
				ee.log.Debug("failed to start leader functions, resigning")
				election.Resign(ee.cli.Ctx())
				continue
			}
			// You are the leader till. So stop campaigning
			// XXX: Do we need to check for leadership periodically, or are we the leader till we die?
			// XXX: What happens during cluster splits? Etcd should handle that, but need to verify.
			return
		}
	}()
}

func (ee *ElasticEtcd) startLeader() error {
	ee.watchVolunteers()
	ee.watchIdealSize()

	return nil
}

func (ee *ElasticEtcd) watchVolunteers() {
	ee.log.Debug("watching for changes to volunteers list")

	f := func(_ clientv3.WatchResponse) {
		ee.log.Debug("volunteer list had a change, doing nominations again")
		ee.doNominations()
	}

	ee.watch(volunteerPrefix, f, clientv3.WithPrefix())
}

func (ee *ElasticEtcd) watchIdealSize() {
	ee.log.Debug("watching for changes to ideal cluster size")

	f := func(resp clientv3.WatchResponse) {
		for _, ev := range resp.Events {
			i, err := strconv.Atoi(string(ev.Kv.Value))
			if err != nil {
				ee.log.WithError(err).Error("could not parse idealsize value, ignoring update")
				continue
			}
			if i != ee.conf.IdealSize {
				ee.log.WithField("idealsize", i).Debug("idealsize changed, doing nominations again")
				ee.conf.IdealSize = i
				ee.doNominations()
			}
		}
	}
	ee.watch(idealSizeKey, f)
}

func (ee *ElasticEtcd) doNominations() {
	ee.lock.Lock()
	defer ee.lock.Unlock()

	// We shouldn't ever hit this, but better be safe
	if ee.stopping {
		return
	}

	ee.log.Debug("doing nominations")

	// First prepare lists of the current nominees and volunteers.

	nomineesResp, err := ee.cli.Get(ee.cli.Ctx(), nomineePrefix, clientv3.WithPrefix(), clientv3.WithSort(clientv3.SortByKey, clientv3.SortAscend))
	if err != nil {
		ee.log.WithError(err).Error("could not get nominees")
		return
	}

	volunteersResp, err := ee.cli.Get(ee.cli.Ctx(), volunteerPrefix, clientv3.WithPrefix(), clientv3.WithSort(clientv3.SortByKey, clientv3.SortAscend))
	if err != nil {
		ee.log.WithError(err).Error("could not get volunteers")
		return
	}

	nominees := keysFromGetResp(nomineesResp, nomineePrefix)
	volunteers := keysFromGetResp(volunteersResp, volunteerPrefix)
	ee.log.WithField("nominees", nominees).Debug("current nominees")
	ee.log.WithField("volunteers", volunteers).Debug("current volunteers")

	// Check if anyone unvolunteered and remove them from the volunteer list
	unvolunteered := diffStringSlices(nominees, volunteers)
	ee.log.WithField("unvolunteered", unvolunteered).Debug("unvolunteered servers")
	for _, h := range unvolunteered {
		if err := ee.removeNomination(h); err != nil {
			// Just log failure and continue
			ee.log.WithError(err).WithField("host", h).Warn("could not remove nomination")
		}
	}

	// Update the nominee list after the nominee removals
	nominees = diffStringSlices(nominees, unvolunteered)
	nomineeCount := len(nominees)
	ee.log.WithField("nominees", nominees).Debug("updated nominees list")

	// Prepare a map of volunteers names and their published CURLs
	volunteersMap, err := urlsMapFromGetResp(volunteersResp, volunteerPrefix)
	if err != nil {
		ee.log.WithError(err).Error("could not prepare volunteers map")
		return
	}

	switch {
	// If idealSize is not met, nominate more servers till the size is met
	case nomineeCount < ee.conf.IdealSize:
		// You cannot do nominations if all volunteers have been nominated
		if compareStringSlices(nominees, volunteers) {
			ee.log.Debug("all available volunteers have been nominated")
			return
		}

		// Filter out already nominated servers
		available := diffStringSlices(volunteers, nominees)

		// Keep nominating in a round-robin fashion till the required nominations are done
		for _, h := range available {
			err := ee.nominate(h, volunteersMap[h])
			if err != nil {
				ee.log.WithError(err).WithField("host", h).Error("failed to nominate host")
				continue
			}
			ee.log.WithField("host", h).Debug("nominated new host")
			nomineeCount++
			if nomineeCount == ee.conf.IdealSize {
				break
			}
		}

		// If idealSize is exceeded, remove server nominations till idealSize is reached
	case nomineeCount > ee.conf.IdealSize:
		// Remove nominations in a round-robin fashion till the required nominations are removed
		for _, h := range nominees {
			if h == ee.conf.Name {
				// skip yourself
				continue
			}
			if err := ee.removeNomination(h); err != nil {
				ee.log.WithError(err).WithField("host", h).Warn("could not remove nomination for host")
				continue
			}
			nomineeCount--
			if nomineeCount == ee.conf.IdealSize {
				break
			}
		}
	}

	ee.log.WithField("nomineecount", nomineeCount).Debug("finished doing nominations")
}

func (ee *ElasticEtcd) nominate(host string, urls types.URLs) error {
	logger := ee.log.WithField("host", host)
	logger.Debug("nominating host")

	// Add the new host to the nominees list.
	// Wait a little for the nominee to pick up the nomination.
	// Then add the new host as an etcd member.
	// If we add the new host as a member first, the embedded etcd will begin
	// trying to connect to the newly added member and which causes new requests
	// to be blocked, for example the PUT request to add the new nominee.
	// If we don't wait for a little while before adding the nominee as a member,
	// the nominee will not be able to read the nominee list it requires to start
	// its server.
	// This is only required for the first nominee added. When more than one
	// server is present the etcd cluster will serve requests when a member is
	// added.
	// TODO: Try to figure a way to avoid the sleeping

	ee.addToNominees(host, urls)

	time.Sleep(2 * time.Second)

	_, err := ee.cli.MemberAdd(ee.cli.Ctx(), urls.StringSlice())
	if err != nil {
		logger.WithError(err).Error("failed to add host as etcd cluster member")
		err := ee.removeFromNominees(host)
		return err
	}

	return nil
}

func (ee *ElasticEtcd) addToNominees(host string, urls types.URLs) error {
	key := nomineePrefix + host
	_, err := ee.cli.Put(ee.cli.Ctx(), key, urls.String())
	if err != nil {
		ee.log.WithError(err).WithField("host", host).Error("failed to add host to nominees list")
	}
	return err
}

func (ee *ElasticEtcd) removeNomination(host string) error {
	logger := ee.log.WithField("host", host)
	logger.Debug("removing nomination for host")

	// Remove host from nomination list
	if err := ee.removeFromNominees(host); err != nil {
		return err
	}

	// Remove host as an etcd member
	memlist, err := ee.cli.MemberList(ee.cli.Ctx())
	if err != nil {
		logger.WithError(err).Error("failed to get memberlist while trying to remove nomination for host")
		return err
	}
	var m *etcdserverpb.Member
	for _, m = range memlist.Members {
		if m.Name == host {
			break
		}
	}
	_, err = ee.cli.MemberRemove(ee.cli.Ctx(), m.ID)
	if err != nil {
		logger.WithError(err).Error("failed to remove host as etcd cluster member")
	}

	return err
}

func (ee *ElasticEtcd) removeFromNominees(host string) error {
	key := nomineePrefix + host
	_, err := ee.cli.Delete(ee.cli.Ctx(), key)
	if err != nil {
		ee.log.WithError(err).WithField("host", host).Error("failed to remove host from nominees list")
	}
	return err
}
