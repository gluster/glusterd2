// Gd3
// Copyright (C) 2017-2018+ James Shubin and the project contributors
// Written by James Shubin <james@shubin.ca> and the project contributors
//
// This program is free software: you can redistribute it and/or modify
// it under the terms of the GNU Affero General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// This program is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
// GNU Affero General Public License for more details.
//
// You should have received a copy of the GNU Affero General Public License
// along with this program.  If not, see <http://www.gnu.org/licenses/>.

package gapi

// ./gd3 run --hostname h1 --tmp-prefix
// ./gd3 run --hostname h2 --tmp-prefix --seeds http://127.0.0.1:2379 --client-urls http://127.0.0.1:2381 --server-urls http://127.0.0.1:2382
// ./gd3 run --hostname h3 --tmp-prefix --seeds http://127.0.0.1:2379 --client-urls http://127.0.0.1:2383 --server-urls http://127.0.0.1:2384
// ./gd3 run --hostname h4 --tmp-prefix --seeds http://127.0.0.1:2379 --client-urls http://127.0.0.1:2385 --server-urls http://127.0.0.1:2386

import (
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"path"
	"strconv"
	"strings"
	"sync"

	//"github.com/purpleidea/gd3/resources/brick"

	"github.com/purpleidea/mgmt/gapi"
	"github.com/purpleidea/mgmt/pgraph"
	"github.com/purpleidea/mgmt/resources"

	systemdUnit "github.com/coreos/go-systemd/unit"
	errwrap "github.com/pkg/errors"
)

const (
	pkgName0           = "centos-release-gluster39"
	pkgName            = "glusterfs-server"   // package name
	pathVarLibGlusterd = "/var/lib/glusterd/" // requires a trailing slash!
	// TODO: we should implement per module namespacing at some point...
	worldNamespace = "gd3::state" // key name of shared string data
	brickBasePort  = 49152
)

// Gd3GAPI implements the main GAPI interface.
type Gd3GAPI struct {
	Program string // program name
	Version string // program name

	data        gapi.Data
	initialized bool
	closeChan   chan struct{}
	wg          sync.WaitGroup // sync group for tunnel go routines
}

// Init initializes the Gd3GAPI struct.
func (obj *Gd3GAPI) Init(data gapi.Data) error {
	if obj.initialized {
		return fmt.Errorf("already initialized")
	}
	if obj.Program == "" {
		return fmt.Errorf("the program name must be specified")
	}

	obj.data = data // store for later
	obj.closeChan = make(chan struct{})
	obj.initialized = true
	return nil
}

// stageCount takes a map of hostname->stage (as strings) and returns the list
// of hostnames at each stage as a map of stage->[]hostname.
func stageCount(stages map[string]string) map[int][]string {
	result := make(map[int][]string)
	var s int
	for key, val := range stages {
		ival, err := strconv.Atoi(val)
		if err != nil {
			s = 0
		}
		s = ival // we found this stage

		if _, ok := result[s]; !ok {
			result[s] = make([]string, 0) // initialize list
		}
		result[s] = append(result[s], key) // add hostname
	}
	return result
}

// stageMin takes a map of hostname->stage (as strings) and a minimum stage
// count and returns the list of hosts that meet this minimum value.
func stageMin(stages map[string]string, min int) []string {
	stageMap := stageCount(stages)
	result := []string{}
	for key, val := range stageMap {
		if key >= min {
			result = append(result, val...) // add those hosts
		}
	}
	return result
}

// Graph returns a current Graph.
func (obj *Gd3GAPI) Graph() (*pgraph.Graph, error) {
	if !obj.initialized {
		return nil, fmt.Errorf("%s: Gd3GAPI is not initialized", obj.Program)
	}

	g := pgraph.NewGraph(obj.Program)
	defaultMetaParams := resources.DefaultMetaParams

	// XXX: hard coded for now, easy to store in etcd and populate from initial peer
	brickCount := 2
	distribute, replicate := 2, 2
	totalHosts := distribute * replicate // 4

	// put user services into $XDG_RUNTIME_DIR/systemd/user/
	// eg: /run/user/1000/systemd/user/
	dir := os.Getenv("XDG_RUNTIME_DIR")
	if dir == "" {
		return nil, fmt.Errorf("the XDG_RUNTIME_DIR env variable is missing or empty")
	}
	// TODO: on Fedora 24 this is what is used. On Fedora 25 systemd-run
	// uses the .../systemd/transient/ directory, but .../systemd/user/ is
	// still seen and it works! Which is the correct dir long term?
	userSvcDir := path.Join(dir, "/systemd/user/")
	log.Printf("%s: User svc dir is: %s", obj.Program, userSvcDir)

	log.Printf("%s: Want %d hosts (%d bricks each) for a dist x repl cluster of %d x %d", obj.Program, totalHosts, brickCount, distribute, replicate)

	keyMap, err := obj.data.World.StrGet(worldNamespace) // map: hostname->state
	if err != nil {
		return nil, errwrap.Wrapf(err, "the Graph can't StrGet the namespace of %s", worldNamespace)
	}

	var stageOne, stageTwo, stageThree bool
	//sc := stageCount(keyMap) // which hosts are at which stages
	sm1 := stageMin(keyMap, 1) // which hosts are at this stage or greater
	if len(sm1) >= totalHosts {
		stageOne = true
	}
	sm2 := stageMin(keyMap, 2)
	if len(sm2) >= totalHosts {
		stageTwo = true
	}
	sm3 := stageMin(keyMap, 3)
	if len(sm3) >= totalHosts {
		stageThree = true
	}
	log.Printf("")
	log.Printf("%s: Found %d hosts at stage one or greater: %+v", obj.Program, len(sm1), sm1)
	log.Printf("%s: Found %d hosts at stage two or greater: %+v", obj.Program, len(sm2), sm2)
	log.Printf("%s: Found %d hosts at stage three or greater: %+v", obj.Program, len(sm3), sm3)
	log.Printf("")

	// key value state resource
	one := "1" // "hello, i'm available for this cluster"
	kv1 := pgraph.NewVertex(&resources.KVRes{
		BaseRes: resources.BaseRes{
			Name:       "kv1",
			MetaParams: defaultMetaParams,
		},
		Key:          worldNamespace,
		Value:        &one,
		SkipLessThan: true, // allow upgrades to two
	})
	g.AddVertex(kv1)
	if !stageOne { // wait for everyone to be at stage one for no particular reason
		return g, nil
	}

	//pkg0 := pgraph.NewVertex(&resources.PkgRes{
	//	BaseRes: resources.BaseRes{
	//		Name:       pkgName0,
	//		MetaParams: defaultMetaParams,
	//	},
	//	State: "installed",
	//})
	//g.AddVertex(pkg0)
	// NOTE: we should do this in parallel, but let's explicitly hello first
	// for the purposes of this demo.
	//g.AddEdge(kv1, pkg0, pgraph.NewEdge("kv1->pkg0"))

	// glusterfs package which includes the glusterfsd binary
	pkg := pgraph.NewVertex(&resources.PkgRes{
		BaseRes: resources.BaseRes{
			Name:       pkgName,
			MetaParams: defaultMetaParams,
		},
		State: "installed",
	})
	g.AddVertex(pkg)
	//g.AddEdge(pkg0, pkg, pgraph.NewEdge("pkg0->pkg"))
	g.AddEdge(kv1, pkg, pgraph.NewEdge("kv1->pkg"))

	// var directory
	d0 := pgraph.NewVertex(&resources.FileRes{
		BaseRes: resources.BaseRes{
			Name:       pathVarLibGlusterd, // directory
			MetaParams: defaultMetaParams,
		},
		//Path: pathVarLibGlusterd,
		State: "present",
	})
	g.AddVertex(d0)
	g.AddEdge(pkg, d0, pgraph.NewEdge("pkg->d0"))

	// key value state resource
	two := "2"
	kv2 := pgraph.NewVertex(&resources.KVRes{
		BaseRes: resources.BaseRes{
			Name:       "kv2",
			MetaParams: defaultMetaParams,
		},
		Key:          worldNamespace,
		Value:        &two,
		SkipLessThan: true, // allow upgrades to three or higher
	})
	g.AddVertex(kv2)
	g.AddEdge(kv1, kv2, pgraph.NewEdge("kv1->kv2")) // for safety

	three := "3"
	kv3 := pgraph.NewVertex(&resources.KVRes{
		BaseRes: resources.BaseRes{
			Name:       "kv3",
			MetaParams: defaultMetaParams,
		},
		Key:          worldNamespace,
		Value:        &three,
		SkipLessThan: true, // allow upgrades to three or higher
	})
	g.AddVertex(kv3)
	g.AddEdge(kv2, kv3, pgraph.NewEdge("kv2->kv3")) // for safety

	dr := pgraph.NewVertex(&resources.ExecRes{
		BaseRes: resources.BaseRes{
			Name:       "reload systemd user units",
			MetaParams: defaultMetaParams,
		},
		Cmd:   "/usr/bin/systemctl --user daemon-reload",
	})
	g.AddVertex(dr)

	for i := 0; i < brickCount; i++ {

		unitName := fmt.Sprintf("glusterfsd%d", i) // unit name!!
		brickName := fmt.Sprintf("b%d", i)         // eg: /bricks/b0
		brickPort := brickBasePort + i             // 49152 + 0 ...
		fullPath := path.Join(userSvcDir, fmt.Sprintf("%s.service", unitName))
		volumeName := "foovol"
		someUUID := "cf7359b0-476a-4336-b7ef-26f106f2986d" // same on all bricks
		socketID := "bcca2de57a5b2e5e9e9941637d873193"     // XXX: different on each brick

		cmd := []string{"/usr/sbin/glusterfsd"}
		cmd = append(cmd, fmt.Sprintf("-s %s", obj.data.Hostname))
		cmd = append(cmd, fmt.Sprintf("--volfile-id %s.%s.bricks-%s", volumeName, obj.data.Hostname, brickName)) // XXX: bricks means /bricks/bX iirc
		cmd = append(cmd, fmt.Sprintf("-p /var/lib/glusterd/vols/%s/run/%s-bricks-%s.pid", volumeName, obj.data.Hostname, brickName))
		cmd = append(cmd, fmt.Sprintf("-S /var/run/gluster/%s.socket", socketID))
		cmd = append(cmd, fmt.Sprintf("--brick-name /bricks/%s", brickName))
		cmd = append(cmd, fmt.Sprintf("-l /var/log/glusterfs/bricks/bricks-%s.log", brickName)) // XXX: mkdir
		cmd = append(cmd, fmt.Sprintf("--xlator-option *-posix.glusterd-uuid=%s", someUUID))    // XXX some uuid same on all bricks
		cmd = append(cmd, fmt.Sprintf("--brick-port %d", brickPort))
		cmd = append(cmd, fmt.Sprintf("--xlator-option %s-server.listen-port=%d", volumeName, brickPort))

		options := []*systemdUnit.UnitOption{
			&systemdUnit.UnitOption{"Unit", "Description", fmt.Sprintf("Gd3 brick %d", i)},
			&systemdUnit.UnitOption{"Service", "ExecStart", strings.Join(cmd, " ")},
			//&systemdUnit.UnitOption{"Unit", "BindsTo", "bar.service"},
			//&systemdUnit.UnitOption{"X-Foo", "Bar", "baz"},
			//&systemdUnit.UnitOption{"Service", "ExecStop", "/usr/bin/sleep 1"},
			&systemdUnit.UnitOption{"Unit", "Documentation", "https://github.com/purpleidea/gd3"},
		}

		outReader := systemdUnit.Serialize(options)
		outBytes, err := ioutil.ReadAll(outReader)
		if err != nil {
			return nil, errwrap.Wrapf(err, "can't generate brick %d unit file", i)
		}
		content := string(outBytes)

		// brick unit file
		bf := pgraph.NewVertex(&resources.FileRes{
			BaseRes: resources.BaseRes{
				Name:       fullPath,
				MetaParams: defaultMetaParams,
			},
			Path:    fullPath,
			State:   "present",
			Content: &content,
		})
		g.AddVertex(bf)
		g.AddEdge(kv2, bf, pgraph.NewEdge(fmt.Sprintf("kv2->bf%d", i)))
		edge := pgraph.NewEdge(fmt.Sprintf("bf%d->dr", i))
		edge.Notify = true // send a notification from brick file to exec reloader
		g.AddEdge(bf, dr, edge)

		if stageTwo {
			// glusterfsd service
			// bonus: now we can use cgroups to limit all these!!!
			svc := pgraph.NewVertex(&resources.SvcRes{
				BaseRes: resources.BaseRes{
					Name:       unitName, // no .service postfix
					MetaParams: defaultMetaParams,
				},
				State:   "running", // TODO: upstream should use a constant
				Startup: "enabled",
				Session: true, // user services that don't run as root!
			})
			g.AddVertex(svc)
			g.AddEdge(bf, svc, pgraph.NewEdge(fmt.Sprintf("bf%d->svc%d", i, i)))
			g.AddEdge(dr, svc, pgraph.NewEdge(fmt.Sprintf("dr->svc%d", i))) // reload before svc
			g.AddEdge(svc, kv3, pgraph.NewEdge(fmt.Sprintf("svc%d->kv3", i)))
		}

		// TODO: build a proper brick resource instead
		//brickRes := pgraph.NewVertex(&brick.BrickRes{
		//})
		//g.AddVertex(brickRes)
	}

	if stageTwo {
		log.Printf("")
		log.Printf("%s: The bricks are now started...", obj.Program)
		log.Printf("")
	}

	if stageThree {
		log.Printf("")
		log.Printf("%s: The volume is now started...", obj.Program)
		log.Printf("")
	}

	//g, err := config.NewGraphFromConfig(obj.data.Hostname, obj.data.World, obj.data.Noop)
	return g, nil
}

// Next returns nil errors every time there could be a new graph.
func (obj *Gd3GAPI) Next() chan error {
	// TODO: should we use obj.data.NoConfigWatch & obj.data.NoStreamWatch ?
	ch := make(chan error)
	stringChan := obj.data.World.StrWatch(worldNamespace) // watch for var changes

	obj.wg.Add(1)
	go func() {
		defer obj.wg.Done()
		defer close(ch) // this will run before the obj.wg.Done()
		if !obj.initialized {
			ch <- fmt.Errorf("%s: Gd3GAPI is not initialized", obj.Program)
			return
		}
		startChan := make(chan struct{}) // start signal
		close(startChan)                 // kick it off!

		for {
			var x error
			var ok bool
			select {
			case <-startChan: // kick the loop once at start
				startChan = nil // disable

			// will delays here block etcd? (no b/c stringChan is buffered!)
			case x, ok = <-stringChan:
				if !ok { // channel closed
					return
				}
			case <-obj.closeChan:
				return
			}
			log.Printf("%s: Generating new graph...", obj.Program)
			select {
			case ch <- x: // trigger a run
			case <-obj.closeChan:
				return
			}
		}
	}()
	return ch
}

// Close shuts down the Gd3GAPI.
func (obj *Gd3GAPI) Close() error {
	if !obj.initialized {
		return fmt.Errorf("%s: Gd3GAPI is not initialized", obj.Program)
	}
	close(obj.closeChan)
	obj.wg.Wait()
	obj.initialized = false // closed = true
	return nil
}
