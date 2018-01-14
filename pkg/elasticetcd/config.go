package elasticetcd

import (
	"net"
	"path"

	"github.com/coreos/etcd/pkg/types"
)

// Default values for the various ElasticEtcd config options
const (
	DefaultCURL      = "http://0.0.0.0:2379"
	DefaultPURL      = "http://0.0.0.0:2380"
	DefaultEndpoint  = "http://localhost:2379"
	DefaultName      = "elasticetcd"
	DefaultIdealSize = 3
	DefaultDir       = "."
)

var (
	// These are the defaults as types.URLs instead of the constant strings.
	// defaultA*URLs and defaultEndpoints are intialized from the available interfaces.
	defaultCURLs, defaultACURLs, defaultPURLs, defaultAPURLs, defaultEndpoints types.URLs
)

// init prepares the defaults on package initialization
func init() {
	defaultCURLs = types.MustNewURLs([]string{DefaultCURL})
	defaultPURLs = types.MustNewURLs([]string{DefaultPURL})
	// This will allow the cluster to be formed.  But auto syncing of addresses
	// for etcd servers and clients will not work.
	defaultACURLs = types.MustNewURLs([]string{DefaultCURL})
	defaultAPURLs = types.MustNewURLs([]string{DefaultPURL})
	defaultEndpoints = types.MustNewURLs([]string{DefaultEndpoint})

	addrs, err := net.InterfaceAddrs()
	if err != nil {
		return
	}

	var acurls, apurls []string
	for _, a := range addrs {
		// Not checking for error here as the given address is returned by the
		// stdlib and is parseable.
		i, _, _ := net.ParseCIDR(a.String())
		if i.IsLoopback() {
			// Loopback addresses are not useful when broadcast
			continue
		}
		str := i.String()
		// Enclose IPv6 adresses with '[]' or the formed URLs will fail parsing
		if i.To4() == nil {
			str = "[" + str + "]"
		}
		curl := "http://" + str + ":2379"
		acurls = append(acurls, curl)
		purl := "http://" + str + ":2380"
		apurls = append(apurls, purl)
	}

	defaultACURLs = types.MustNewURLs(acurls)
	defaultAPURLs = types.MustNewURLs(apurls)
	defaultEndpoints = defaultACURLs
}

// Config is holds the configuration for an ElasticEtcd
type Config struct {
	Name, Dir, LogDir       string
	Endpoints, CURLs, PURLs types.URLs
	IdealSize               int
	DisableLogging          bool
	UseTLS                  bool
	CAFile, TrustedCAFile   string
	CertFile                string
	KeyFile                 string
	ClntCertFile            string
	ClntKeyFile             string
}

// NewConfig returns an ElasticEtcd config with defaults filled
func NewConfig() *Config {
	return &Config{
		Name:      DefaultName,
		Dir:       DefaultDir,
		LogDir:    path.Join(DefaultDir, "log"),
		Endpoints: defaultEndpoints,
		CURLs:     defaultCURLs,
		PURLs:     defaultPURLs,
		IdealSize: DefaultIdealSize,
	}
}

func isDefaultCURL(urls types.URLs) bool {
	return isDefaultURL(urls, DefaultCURL)
}

func isDefaultPURL(urls types.URLs) bool {
	return isDefaultURL(urls, DefaultPURL)
}

func isDefaultEndpoint(urls types.URLs) bool {
	return isDefaultURL(urls, DefaultEndpoint)
}

func isDefaultURL(urls types.URLs, def string) bool {
	if len(urls) > 1 {
		return false
	}
	return urls[0].String() == def
}
