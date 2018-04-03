package peer

import (
	"fmt"
	"net"

	"github.com/gluster/glusterd2/glusterd2/gdctx"

	config "github.com/spf13/viper"
)

func normalizeAddrs() ([]string, error) {

	shost, sport, err := net.SplitHostPort(config.GetString("clientaddress"))
	if err != nil {
		return nil, err
	}

	if shost != "" {
		return []string{config.GetString("clientaddress")}, nil
	}

	addrs, err := net.InterfaceAddrs()
	if err != nil {
		return nil, err
	}

	var clientAddrs []string
	for _, addr := range addrs {
		if ipnet, ok := addr.(*net.IPNet); ok {
			if ipnet.IP.To4() != nil {
				clientAddrs = append(clientAddrs, fmt.Sprintf("%s:%s", ipnet.IP.String(), sport))
			}
		}
	}

	return clientAddrs, nil
}

// AddSelfDetails results in the peer adding its own details into etcd
func AddSelfDetails() error {

	var err error
	p := &Peer{
		ID:            gdctx.MyUUID,
		Name:          gdctx.HostName,
		PeerAddresses: []string{config.GetString("peeraddress")},
	}
	p.ClientAddresses, err = normalizeAddrs()
	if err != nil {
		return err
	}
	peerInfo, err := GetPeer(string(gdctx.MyUUID))
	if err != nil {

		if p.MetaData == nil {
			p.MetaData = make(map[string]string)
		}
		if p.MetaData["_zone"] == "" {
			p.MetaData["_zone"] = p.ID.String()
		}

		return AddOrUpdatePeer(p)
	}
	peerInfo.MetaData = p.MetaData
	return nil
}
