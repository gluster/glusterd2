package mgmt

import (
	"fmt"
	"sync"

	"github.com/purpleidea/mgmt/gapi"
	"github.com/purpleidea/mgmt/pgraph"
)

// GlusterGAPI implements the main GAPI interface for Gluster.
type GlusterGAPI struct {
	Name string // graph name

	data        gapi.Data
	initialized bool
	closeChan   chan struct{}
	wg          sync.WaitGroup // sync group for tunnel go routines
}

// Init initializes the GlusterGAPI struct.
func (obj *GlusterGAPI) Init(data gapi.Data) error {
	if obj.initialized {
		return fmt.Errorf("already initialized")
	}
	if obj.Name == "" {
		return fmt.Errorf("the graph name must be specified")
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

// Next returns nil errors every time there could be a new graph.
func (obj *GlusterGAPI) Next() chan error {
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
