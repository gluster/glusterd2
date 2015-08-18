package main

import (
	"fmt"
	"os"
	"os/signal"

	"github.com/kshlm/glusterd2/commands"
	"github.com/kshlm/glusterd2/context"
	"github.com/kshlm/glusterd2/volgen"

	log "github.com/Sirupsen/logrus"
)

func volgen_generate_graph() {
	var i, j int

	volgen.Init()

	if volgen.Gtype == "SERVER" {
		i = volgen.Bcount
	} else {
		i = 1
	}

	hname, _ := os.Hostname()

	for ; j < i; j++ {
		graph := volgen.Generate_graph()

		fname := fmt.Sprintf("/tmp/%s.%s.brick%d.vol", volgen.Volname, hname, j)

		if volgen.Gtype == "SERVER" {
			f, err := os.Create(fname)
			if err != nil {
				panic(err)
			}
			defer closeFile(f)
			graph.DumpGraph(f)
		} else {
			f, err := os.Create(volgen.File_name)
			if err != nil {
				panic(err)
			}
			defer closeFile(f)
			graph.DumpGraph(f)
		}
	}
}

func main() {
	log.Info("GlusterD starting")

	context.Init()

	volgen_generate_graph()

	for _, c := range commands.Commands {
		context.Rest.SetRoutes(c.Routes())
	}

	sigCh := make(chan os.Signal)
	signal.Notify(sigCh)
	go func() {
		for s := range sigCh {
			log.WithField("signal", s).Debug("Signal recieved")
			switch s {
			case os.Interrupt:
				log.WithField("signal", s).Info("Recieved SIGTERM. Stopping GlusterD.")
				context.Rest.Stop()
				log.Info("Termintaing GlusterD.")
				os.Exit(0)

			default:
				continue
			}
		}
	}()

	err := context.Rest.Listen()
	if err != nil {
		log.Fatal("Could not start GlusterD Rest Server. Aborting.")
	}
}

func closeFile(f *os.File) {
	f.Close()
}
