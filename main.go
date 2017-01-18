package main

import (
	"os"
	"os/signal"
	"path"

	"github.com/gluster/glusterd2/commands"
	"github.com/gluster/glusterd2/gdctx"
	"github.com/gluster/glusterd2/peer"
	"github.com/gluster/glusterd2/rpc/server"
	"github.com/gluster/glusterd2/utils"

	mgmt "github.com/purpleidea/mgmt/lib"
	"github.com/purpleidea/mgmt/pgraph"
	log "github.com/Sirupsen/logrus"
	flag "github.com/spf13/pflag"
	config "github.com/spf13/viper"
)

// GlusterGAPI implements the main GAPI interface for Gluster.
type GlusterGAPI struct {
	Name     string // graph name

	data        gapi.Data
	initialized bool
	closeChan   chan struct{}
	wg          sync.WaitGroup // sync group for tunnel go routines
}

// Init initializes the GlusterGAPI struct.
func (obj *GlusterGAPI) Init(data gapi.Data) error {
	if obj.initialized {
		return fmt.Errorf("Already initialized!")
	}
	if obj.Name == "" {
		return fmt.Errorf("The graph name must be specified!")
	}
	obj.data = data // store for later
	obj.closeChan = make(chan struct{})
	obj.initialized = true
	return nil
}

// Graph returns a current Graph.
func (obj *GlusterGAPI) Graph() (*pgraph.Graph, error) {
	if !obj.initialized {
		return nil, fmt.Errorf("libmgmt: GlusterGAPI is not initialized")
	}

	g := pgraph.NewGraph(obj.Name)

	// XXX: nothing happens here yet, TODO!

	//g, err := config.NewGraphFromConfig(obj.data.Hostname, obj.data.World, obj.data.Noop)
	return g, nil
}

// SwitchStream returns nil errors every time there could be a new graph.
func (obj *GlusterGAPI) SwitchStream() chan error {
	if obj.data.NoWatch {
		return nil
	}
	ch := make(chan error)
	obj.wg.Add(1)
	go func() {
		defer obj.wg.Done()
		defer close(ch) // this will run before the obj.wg.Done()
		if !obj.initialized {
			ch <- fmt.Errorf("libmgmt: GlusterGAPI is not initialized")
			return
		}

		// XXX: do something!
		// arbitrarily change graph every interval seconds
		//ticker := time.NewTicker(time.Duration(obj.Interval) * time.Second)
		//defer ticker.Stop()
		for {
			select {
			//case <-ticker.C:
			//	log.Printf("libmgmt: Generating new graph...")
			//	ch <- nil // trigger a run
			case <-obj.closeChan:
				return
			}
		}
	}()
	return ch
}

// Close shuts down the GlusterGAPI.
func (obj *GlusterGAPI) Close() error {
	if !obj.initialized {
		return fmt.Errorf("libmgmt: GlusterGAPI is not initialized")
	}
	close(obj.closeChan)
	obj.wg.Wait()
	obj.initialized = false // closed = true
	return nil
}

func main() {

	// Set IP and hostname once.
	gdctx.SetHostnameAndIP()

	// Parse flags and handle version and logging before continuing
	parseFlags()

	showvers, _ := flag.CommandLine.GetBool("version")
	if showvers {
		dumpVersionInfo()
		return
	}

	logLevel, _ := flag.CommandLine.GetString("loglevel")
	initLog(logLevel, os.Stderr)

	log.WithField("pid", os.Getpid()).Info("GlusterD starting")

	// Read in config
	confFile, _ := flag.CommandLine.GetString("config")
	initConfig(confFile)

	// Change to working directory before continuing
	if e := os.Chdir(config.GetString("workdir")); e != nil {
		log.WithError(e).Fatalf("failed to change working directory")
	}

	// TODO: This really should go into its own function.
	utils.InitDir(config.GetString("localstatedir"))
	utils.InitDir(config.GetString("rundir"))
	utils.InitDir(config.GetString("logdir"))
	utils.InitDir(path.Join(config.GetString("rundir"), "gluster"))
	utils.InitDir(path.Join(config.GetString("logdir"), "glusterfs/bricks"))

	gdctx.MyUUID = gdctx.InitMyUUID()

	// XXX
	//// Start embedded etcd server
	//etcdConfig, err := etcdmgmt.GetEtcdConfig(true)
	//if err != nil {
	//	log.WithField("Error", err).Fatal("Could not fetch config options for etcd.")
	//}
	//err = etcdmgmt.StartEmbeddedEtcd(etcdConfig)
	//if err != nil {
	//	log.WithField("Error", err).Fatal("Could not start embedded etcd server.")
	//}

	// set all the options we want here...
	libmgmt := &mgmt.Main{}
	libmgmt.Program = "glusterd2"
	//libmgmt.Version = "0.0.1"   // TODO: set on compilation
	libmgmt.TmpPrefix = true // prod things probably don't want this on
	//prefix := "/tmp/testprefix/"
	//libmgmt.Prefix = &p // enable for easy debugging
	libmgmt.IdealClusterSize = -1
	libmgmt.ConvergedTimeout = -1
	libmgmt.Noop = false // FIXME: careful!

	libmgmt.GAPI = &GlusterGAPI{ // graph API
		Name:     "glusterd2", // TODO: set on compilation
	}

	if err := libmgmt.Init(); err != nil {
		log.Fatal("Init failed")
	}

	gdctx.Init()

	for _, c := range commands.Commands {
		gdctx.Rest.SetRoutes(c.Routes())
		c.RegisterStepFuncs()
	}

	// Store self information in the store if GlusterD is coming up for
	// first time
	if !gdctx.Restart {
		peer.AddSelfDetails()
	}

	// Start listening for incoming RPC requests
	err = server.StartListener()
	if err != nil {
		log.Fatal("Could not register RPC listener. Aborting")
	}

	sigCh := make(chan os.Signal)
	signal.Notify(sigCh)
	go func() {
		for s := range sigCh {
			log.WithField("signal", s).Debug("Signal recieved")
			switch s {
			case os.Interrupt:
				log.WithField("signal", s).Info("Recieved SIGTERM. Stopping GlusterD.")
				gdctx.Rest.Stop()
				//etcdmgmt.DestroyEmbeddedEtcd()
				server.StopServer()
				log.Info("Termintaing GlusterD.")
				libmgmt.Exit(nil) // pass in an error if you want to exit with error
				os.Exit(0)

			default:
				continue
			}
		}
	}()

	// this blocks until it shuts down, it causes etcd to startup based on args
	if err := libmgmt.Run(); err != nil { // this error comes from mgmt internals shutting down or from libmgmt.Exit(...)
		log.Fatal("Run failed", err) // XXX: errwrap, and return
	}

//	// Start GlusterD REST server
//	err = gdctx.Rest.Listen()
//	if err != nil {
//		log.Fatal("Could not start GlusterD Rest Server. Aborting.")
//	}

}
