package volgen

import (
	"flag"
	"fmt"
	"os"
)

var (
	File_name, Daemon                     string
	Volname, Vtype, Gtype                 string
	arg_len, Bcount, Dcount, ReplicaCount int
)

func Init() {
	flag.StringVar(&File_name, "vpath", "", "volfile path")
	flag.StringVar(&Volname, "volname", "", "volume name")
	flag.StringVar(&Daemon, "daemon", "", "daemon for which volfile generated")
	flag.IntVar(&ReplicaCount, "replica", 0, "Replica count for replicate volume")
	flag.StringVar(&Gtype, "gtype", "", "Graph type for eg: fuse, nfs, bitd, scrubd, shd etc..")

	flag.Parse()

	fmt.Printf("file name is %v, volume name is %s, daemon is %v\n",
		File_name, Volname, Daemon)

	if len(File_name) != 0 {
	} else {
		fmt.Println("Exiting! Please give volfile path")
		os.Exit(2) /*Exiting with error status 2*/
	}

	if len(Volname) != 0 {
	} else {
		fmt.Println("Exiting! Please give volume name")
		os.Exit(2) /*Exiting with error status 2*/
	}

	if len(Daemon) != 0 {
	} else {
		fmt.Println("Exiting! Please give daemon name")
		os.Exit(2)
	}

	if len(Gtype) != 0 {
	} else {
		fmt.Println("Exiting! Please give Graph name for which you want to generate graph")
		os.Exit(2)
	}

	fmt.Println("How many brick")
	fmt.Scanf("%d", &Bcount)

	if Bcount == 0 {
		fmt.Println("Brick count must be greater then 0")
		os.Exit(2)
	}

	if ReplicaCount != 0 {
		if ReplicaCount == 1 {
			fmt.Println("Error! Replica count must be grater then 1")
			os.Exit(2)
		} else if Bcount%ReplicaCount != 0 {
			fmt.Println("Exiting! Replica count should be multiple of brick count")
			os.Exit(2)
		} else {
			Dcount = Bcount / ReplicaCount
			Vtype = "REPLICATE"
		}
	}
}
