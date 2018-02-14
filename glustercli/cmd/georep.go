package cmd

import (
	"errors"
	"fmt"
	"os"
	"strings"

	"github.com/gluster/glusterd2/pkg/restclient"
	georepapi "github.com/gluster/glusterd2/plugins/georeplication/api"

	"github.com/olekukonko/tablewriter"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

const (
	helpGeorepCmd                  = "Gluster Geo-replication"
	helpGeorepCreateCmd            = "Create a Geo-replication Session"
	helpGeorepStartCmd             = "Start a Geo-replication Session"
	helpGeorepStopCmd              = "Stop a Geo-replication Session"
	helpGeorepDeleteCmd            = "Delete a Geo-replication Session"
	helpGeorepPauseCmd             = "Pause a Geo-replication Session"
	helpGeorepResumeCmd            = "Resume a Geo-replication Session"
	helpGeorepStatusCmd            = "Status of Geo-replication Sessions"
	helpGeorepConfigGetCmd         = "Geo-replication Session Configurations"
	helpGeorepConfigSetCmd         = "Geo-replication Session Config management"
	helpGeorepConfigResetCmd       = "Reset Geo-replication Session Configurations"
	errGeorepSessionCreationFailed = "Georep session creation failed.\n"
	errGeorepSSHKeysGenerate       = `Failed to create SSH Keys in one or more Master Volume nodes.
Please check the log file for more details`
	errGeorepMasterInfoNotAvailable = `Failed to get Master Volume details, Please check
Glusterd is running and Master Volume name is Valid`
	errGeorepRemoteInfoNotAvailable = `Failed to get Remote Volume details, Please check Glusterd is running
in the remote Cluster and reachable from this node`
	errGeorepSessionAlreadyExists = `Geo-replication session already exists, Use --force to
update the existing session and redistribute the SSH Keys to Remote Cluster`
	errGeorepSSHKeysPush = `Geo-replication session created successfully. But failed to push
SSH Keys to Remote Cluster. Please check the following
- Glusterd is running in the remote node and reachable from this node.
- If any nodes belonging to the remote Volume are down

Rerun the Session Create command with --force once the issues related to Remote Cluster is resolved`
	errGeorepStatusCommandFailed = "Geo-replication Status command failed.\n"
)

var (
	flagGeorepCmdForce            bool
	flagGeorepShowAllConfig       bool
	flagGeorepRemoteGlusterdHTTPS bool
	flagGeorepRemoteGlusterdHost  string
	flagGeorepRemoteGlusterdPort  int
)

func init() {
	// Geo-rep Create
	georepCreateCmd.Flags().BoolVarP(&flagGeorepRemoteGlusterdHTTPS, "remote-glusterd-https", "", false, "Remote Glusterd HTTPS")
	georepCreateCmd.Flags().StringVarP(&flagGeorepRemoteGlusterdHost, "remote-glusterd-host", "", "", "Remote Glusterd Host")
	georepCreateCmd.Flags().IntVarP(&flagGeorepRemoteGlusterdPort, "remote-glusterd-port", "", 24007, "Remote Glusterd Port")
	georepCreateCmd.Flags().BoolVarP(&flagGeorepCmdForce, "force", "f", false, "Force")
	georepCmd.AddCommand(georepCreateCmd)

	// Geo-rep Start
	georepStartCmd.Flags().BoolVarP(&flagGeorepCmdForce, "force", "f", false, "Force")
	georepCmd.AddCommand(georepStartCmd)

	// Geo-rep Stop
	georepStopCmd.Flags().BoolVarP(&flagGeorepCmdForce, "force", "f", false, "Force")
	georepCmd.AddCommand(georepStopCmd)

	// Geo-rep Delete
	georepDeleteCmd.Flags().BoolVarP(&flagGeorepCmdForce, "force", "f", false, "Force")
	georepCmd.AddCommand(georepDeleteCmd)

	// Geo-rep Pause
	georepPauseCmd.Flags().BoolVarP(&flagGeorepCmdForce, "force", "f", false, "Force")
	georepCmd.AddCommand(georepPauseCmd)

	// Geo-rep Resume
	georepResumeCmd.Flags().BoolVarP(&flagGeorepCmdForce, "force", "f", false, "Force")
	georepCmd.AddCommand(georepResumeCmd)

	// Geo-rep Status
	georepCmd.AddCommand(georepStatusCmd)

	// Geo-rep Config
	georepGetCmd.Flags().BoolVarP(&flagGeorepShowAllConfig, "show-all", "a", false, "Show all Configurations")
	georepCmd.AddCommand(georepGetCmd)
	georepCmd.AddCommand(georepSetCmd)
	georepCmd.AddCommand(georepResetCmd)

	RootCmd.AddCommand(georepCmd)
}

var georepCmd = &cobra.Command{
	Use:   "geo-replication",
	Short: helpGeorepCmd,
}

type volumeDetails struct {
	volname string
	id      string
	nodes   []georepapi.GeorepRemoteHostReq
}

func getVolumeDetails(volname string, rclient *restclient.Client) (*volumeDetails, error) {
	var c *restclient.Client
	var master bool

	if rclient == nil {
		// Local Cluster, Use Default Client
		c = client
		master = true
	} else {
		// Use the Client specified(Client to connect to remote Cluster)
		c = rclient
		master = false
	}

	vols, err := c.Volumes(volname)
	if err != nil || len(vols) == 0 {
		emsg := errGeorepMasterInfoNotAvailable
		if !master {
			emsg = errGeorepRemoteInfoNotAvailable
		}

		log.WithFields(log.Fields{
			"volume": volname,
			"error":  err.Error(),
		}).Error("failed to get Volume details")
		return nil, errors.New(emsg)
	}

	var nodesdata []georepapi.GeorepRemoteHostReq

	if rclient != nil {
		// Node details are not required for Master Volume
		nodes := make(map[string]bool)
		for _, subvol := range vols[0].Subvols {
			for _, brick := range subvol.Bricks {
				if _, ok := nodes[brick.NodeID.String()]; !ok {
					nodes[brick.NodeID.String()] = true
					nodesdata = append(nodesdata, georepapi.GeorepRemoteHostReq{NodeID: brick.NodeID.String(), Hostname: brick.Hostname})
				}
			}
		}
	}

	return &volumeDetails{
		volname: volname,
		id:      vols[0].ID.String(),
		nodes:   nodesdata,
	}, nil
}

var georepCreateCmd = &cobra.Command{
	Use:   "create <master-volume> [<remote-user>@]<remote-host>::<remote-volume>",
	Short: helpGeorepCreateCmd,
	Args:  cobra.ExactArgs(2),
	Run: func(cmd *cobra.Command, args []string) {
		volname := args[0]
		remotehostvol := strings.Split(args[1], "::")
		remoteuserhost := strings.Split(remotehostvol[0], "@")
		remoteuser := "root"
		remotehost := remoteuserhost[0]
		remotevol := remotehostvol[1]
		if len(remoteuserhost) > 1 {
			remotehost = remoteuserhost[1]
			remoteuser = remoteuserhost[0]
		}

		masterdata, err := getVolumeDetails(volname, nil)
		if err != nil {
			failure(errGeorepSessionCreationFailed, err, 1)
		}

		rclient := getRemoteClient(remotehost)

		remotevoldata, err := getVolumeDetails(remotevol, rclient)
		if err != nil {
			handleGlusterdConnectFailure(errGeorepSessionCreationFailed, err, flagGeorepRemoteGlusterdHTTPS, flagGeorepRemoteGlusterdHost, flagGeorepRemoteGlusterdPort, 1)

			// If not Glusterd connect Failure
			failure(errGeorepSessionCreationFailed, err, 1)
		}

		// Generate SSH Keys from all nodes of Master Volume
		sshkeys, err := client.GeorepSSHKeysGenerate(volname)
		if err != nil {
			log.WithFields(log.Fields{
				"volume": volname,
				"error":  err.Error(),
			}).Error("failed to generate SSH Keys")

			failure(errGeorepSessionCreationFailed+errGeorepSSHKeysGenerate, err, 1)
		}

		_, err = client.GeorepCreate(masterdata.id, remotevoldata.id, georepapi.GeorepCreateReq{
			MasterVol:   volname,
			RemoteUser:  remoteuser,
			RemoteHosts: remotevoldata.nodes,
			RemoteVol:   remotevol,
			Force:       flagGeorepCmdForce,
		})

		if err != nil {
			log.WithField("volume", volname).Println("georep session creation failed")
			failure(errGeorepSessionCreationFailed, err, 1)
		}

		err = rclient.GeorepSSHKeysPush(remotevol, sshkeys)
		if err != nil {
			log.WithFields(log.Fields{
				"volume": remotevol,
				"error":  err.Error(),
			}).Error("failed to push SSH Keys to Remote Cluster")

			handleGlusterdConnectFailure(errGeorepSessionCreationFailed, err, flagGeorepRemoteGlusterdHTTPS, flagGeorepRemoteGlusterdHost, flagGeorepRemoteGlusterdPort, 1)

			// If not Glusterd connect issue
			failure(errGeorepSSHKeysPush, err, 1)
		}

		fmt.Println("Geo-replication session created successfully")
	},
}

type georepAction int8

const (
	georepStart georepAction = iota
	georepStop
	georepPause
	georepResume
	georepDelete
)

func (a *georepAction) String() string {
	switch *a {
	case georepStart:
		return "Start"
	case georepStop:
		return "Stop"
	case georepPause:
		return "Pause"
	case georepResume:
		return "Resume"
	case georepDelete:
		return "Delete"
	}
	return ""
}

func handleGeorepAction(args []string, action georepAction) {
	masterdata, remotedata, err := getVolIDs(args)
	if err != nil {
		failure(fmt.Sprintf("Geo-replication %s failed.\n", action.String()), err, 1)
	}
	switch action {
	case georepStart:
		_, err = client.GeorepStart(masterdata.id, remotedata.id, flagGeorepCmdForce)
	case georepStop:
		_, err = client.GeorepStop(masterdata.id, remotedata.id, flagGeorepCmdForce)
	case georepPause:
		_, err = client.GeorepPause(masterdata.id, remotedata.id, flagGeorepCmdForce)
	case georepResume:
		_, err = client.GeorepResume(masterdata.id, remotedata.id, flagGeorepCmdForce)
	case georepDelete:
		err = client.GeorepDelete(masterdata.id, remotedata.id, flagGeorepCmdForce)
	}

	if err != nil {
		log.WithFields(log.Fields{
			"volume": masterdata.volname,
			"error":  err.Error(),
		}).Error("geo-replication", action.String(), "failed")
		failure(fmt.Sprintf("Geo-replication %s failed", action.String()), err, 1)
	}
	fmt.Println("Geo-replication session", action.String(), "successful")
}

var georepStartCmd = &cobra.Command{
	Use:   "start <master-volume> [<remote-user>@]<remote-host>::<remote-volume>",
	Short: helpGeorepStartCmd,
	Args:  cobra.ExactArgs(2),
	Run: func(cmd *cobra.Command, args []string) {
		handleGeorepAction(args, georepStart)
	},
}

var georepStopCmd = &cobra.Command{
	Use:   "stop <master-volume> [<remote-user>@]<remote-host>::<remote-volume>",
	Short: helpGeorepStopCmd,
	Args:  cobra.ExactArgs(2),
	Run: func(cmd *cobra.Command, args []string) {
		handleGeorepAction(args, georepStop)
	},
}

var georepPauseCmd = &cobra.Command{
	Use:   "pause <master-volume> [<remote-user>@]<remote-host>::<remote-volume>",
	Short: helpGeorepPauseCmd,
	Args:  cobra.ExactArgs(2),
	Run: func(cmd *cobra.Command, args []string) {
		handleGeorepAction(args, georepPause)
	},
}

var georepResumeCmd = &cobra.Command{
	Use:   "resume <master-volume> [<remote-user>@]<remote-host>::<remote-volume>",
	Short: helpGeorepResumeCmd,
	Args:  cobra.ExactArgs(2),
	Run: func(cmd *cobra.Command, args []string) {
		handleGeorepAction(args, georepResume)
	},
}

var georepDeleteCmd = &cobra.Command{
	Use:   "delete <master-volume> [<remote-user>@]<remote-host>::<remote-volume>",
	Short: helpGeorepDeleteCmd,
	Args:  cobra.ExactArgs(2),
	Run: func(cmd *cobra.Command, args []string) {
		handleGeorepAction(args, georepDelete)
	},
}

func getRemoteClient(host string) *restclient.Client {
	// TODO: Handle Remote Cluster Authentication and certificates and URL scheme
	scheme := "http"
	if flagGeorepRemoteGlusterdHTTPS {
		scheme = "https"
	}
	if flagGeorepRemoteGlusterdHost != "" {
		host = flagGeorepRemoteGlusterdHost
	}

	// Set Global based on final decision
	flagGeorepRemoteGlusterdHost = host

	return restclient.New(fmt.Sprintf("%s://%s:%d", scheme, host, flagGeorepRemoteGlusterdPort),
		"", "", "", true)
}

func getVolIDs(pargs []string) (*volumeDetails, *volumeDetails, error) {
	var masterdata *volumeDetails
	var remotedata *volumeDetails
	var err error

	if len(pargs) >= 1 {
		masterdata, err = getVolumeDetails(pargs[0], nil)
		if err != nil {
			return nil, nil, err
		}
	}

	if len(pargs) >= 2 {
		remotehostvol := strings.Split(pargs[1], "::")
		rclient := getRemoteClient(remotehostvol[0])
		remotedata, err = getVolumeDetails(remotehostvol[1], rclient)
		if err != nil {
			return nil, nil, err
		}
	}
	return masterdata, remotedata, nil
}

var georepStatusCmd = &cobra.Command{
	Use:   "status [<master-volume> [[<remote-user>@]<remote-host>::<remote-volume>]]",
	Short: helpGeorepStatusCmd,
	Args:  cobra.RangeArgs(0, 2),
	Run: func(cmd *cobra.Command, args []string) {
		var err error
		masterdata, remotedata, err := getVolIDs(args)
		if err != nil {
			failure(errGeorepStatusCommandFailed, err, 1)
		}

		var sessions []georepapi.GeorepSession
		// If mastervolid or remotevolid is empty then get status of all and then filter
		if masterdata == nil || remotedata == nil {
			allSessions, err := client.GeorepStatus("", "")
			if err != nil {
				failure(errGeorepStatusCommandFailed, err, 1)
			}
			for _, s := range allSessions {
				if masterdata != nil && s.MasterID.String() != masterdata.id {
					continue
				}
				if remotedata != nil && s.RemoteID.String() != remotedata.id {
					continue
				}
				sessionDetail, err := client.GeorepStatus(s.MasterID.String(), s.RemoteID.String())
				if err != nil {
					failure(errGeorepStatusCommandFailed, err, 1)
				}
				sessions = append(sessions, sessionDetail[0])
			}
		} else {
			sessions, err = client.GeorepStatus(masterdata.id, remotedata.id)
			if err != nil {
				failure(errGeorepStatusCommandFailed, err, 1)
			}
		}

		for _, session := range sessions {
			fmt.Println()
			fmt.Printf("SESSION: %s ==> %s@%s::%s  STATUS: %s\n",
				session.MasterVol,
				session.RemoteUser,
				session.RemoteHosts[0].Hostname,
				session.RemoteVol,
				session.Status,
			)

			// Status Detail
			if len(session.Workers) > 0 {
				table := tablewriter.NewWriter(os.Stdout)
				table.SetHeader([]string{"Master Brick", "Status", "Crawl Status", "Remote Node", "Last Synced", "Checkpoint Time", "Checkpoint Completion Time"})
				for _, worker := range session.Workers {
					table.Append([]string{
						worker.MasterNode + ":" + worker.MasterBrickPath,
						worker.Status,
						worker.CrawlStatus,
						worker.RemoteNode,
						worker.LastSyncedTime,
						worker.CheckpointTime,
						worker.CheckpointCompletedTime,
					})
				}
				table.Render()
				fmt.Println()
			}
		}
	},
}

var georepGetCmd = &cobra.Command{
	Use:   "get <master-volume> [<remote-user>@]<remote-host>::<remote-volume>",
	Short: helpGeorepConfigGetCmd,
	Args:  cobra.ExactArgs(2),
	Run: func(cmd *cobra.Command, args []string) {
		var err error

		masterdata, remotedata, err := getVolIDs(args)
		if err != nil {
			failure("Error getting Volume IDs", err, 1)
		}

		opts, err := client.GeorepGet(masterdata.id, remotedata.id)
		if err != nil {
			failure("Error getting Options", err, 1)
		}

		table := tablewriter.NewWriter(os.Stdout)
		table.SetAlignment(tablewriter.ALIGN_LEFT)
		if len(opts) > 0 {
			table.SetHeader([]string{"Name", "Value", "Default Value"})
		}
		// User Configured Values
		numOpts := 0
		for _, opt := range opts {
			if !opt.Modified {
				continue
			}

			configurableMsg := ""
			if !opt.Configurable {
				configurableMsg = "(Not configurable)"
			}
			numOpts++
			table.Append([]string{opt.Name + configurableMsg, opt.Value, opt.DefaultValue})
		}

		if numOpts > 0 {
			fmt.Println()
			fmt.Println("Session Configurations:")
			table.Render()
		}

		// Other Configs
		if flagGeorepShowAllConfig {
			table = tablewriter.NewWriter(os.Stdout)
			table.SetAlignment(tablewriter.ALIGN_LEFT)
			table.SetHeader([]string{"Name", "Value"})
			numOpts = 0
			for _, opt := range opts {
				if opt.Modified {
					continue
				}

				configurableMsg := ""
				if !opt.Configurable {
					configurableMsg = "(Not configurable)"
				}
				numOpts++
				// Show value as empty since not modified
				table.Append([]string{opt.Name + configurableMsg, opt.Value})
			}
			if numOpts > 0 {
				fmt.Println()
				fmt.Println("Default Configurations:")
				table.Render()
			}
		}

	},
}

var georepSetCmd = &cobra.Command{
	Use:   "set <master-volume> [<remote-user>@]<remote-host>::<remote-volume> <name> <value>",
	Short: helpGeorepConfigSetCmd,
	Args:  cobra.ExactArgs(4),
	Run: func(cmd *cobra.Command, args []string) {
		var err error
		masterdata, remotedata, err := getVolIDs(args)
		if err != nil {
			failure("Error getting Volume IDs", err, 1)
		}

		opts := make(map[string]string)
		opts[args[2]] = args[3]

		err = client.GeorepSet(masterdata.id, remotedata.id, opts)
		if err != nil {
			failure("Geo-replication session config set failed", err, 1)
		}
		fmt.Println("Geo-replication session config set successfully")
	},
}

var georepResetCmd = &cobra.Command{
	Use:   "reset <master-volume> [<remote-user>@]<remote-host>::<remote-volume> <name>",
	Short: helpGeorepConfigResetCmd,
	Args:  cobra.MinimumNArgs(3),
	Run: func(cmd *cobra.Command, args []string) {
		var err error
		masterdata, remotedata, err := getVolIDs(args)
		if err != nil {
			failure(err.Error(), err, 1)
		}

		err = client.GeorepReset(masterdata.id, remotedata.id, args[2:])
		if err != nil {
			failure("Geo-replication session config reset failed", err, 1)
		}
		fmt.Println("Geo-replication session config reset successfully")
	},
}
