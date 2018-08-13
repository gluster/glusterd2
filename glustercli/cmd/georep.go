package cmd

import (
	"errors"
	"fmt"
	"net/url"
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

const (
	geoRepHTTPScheme   = "http"
	geoRepGlusterdPort = 24007
)

var (
	flagGeorepCmdForce        bool
	flagGeorepShowAllConfig   bool
	flagGeorepRemoteEndpoints string
)

func init() {
	// Geo-rep Create
	georepCreateCmd.Flags().StringVar(&flagGeorepRemoteEndpoints, "remote-endpoints", "", "remote glusterd2 endpoints")
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
		if GlobalFlag.Verbose {
			log.WithError(err).WithField("volume", volname).Error("failed to get Volume details")
		}
		return nil, errors.New(emsg)
	}

	var nodesdata []georepapi.GeorepRemoteHostReq

	if rclient != nil {
		// Node details are not required for Master Volume
		nodes := make(map[string]bool)
		for _, subvol := range vols[0].Subvols {
			for _, brick := range subvol.Bricks {
				if _, ok := nodes[brick.PeerID.String()]; !ok {
					nodes[brick.PeerID.String()] = true
					nodesdata = append(nodesdata, georepapi.GeorepRemoteHostReq{PeerID: brick.PeerID.String(), Hostname: brick.Hostname})
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

func parseRemoteData(data string) (string, string, string, error) {
	remotehostvol := strings.Split(data, "::")

	if len(remotehostvol) != 2 {
		return "", "", "", errors.New("invalid Remote Volume details, use <remoteuser>@<remotehost>::<remotevol> format")
	}

	remoteuserhost := strings.Split(remotehostvol[0], "@")
	remoteuser := "root"
	remotehost := remoteuserhost[0]
	remotevol := remotehostvol[1]
	if len(remoteuserhost) > 1 {
		remotehost = remoteuserhost[1]
		remoteuser = remoteuserhost[0]
	}
	return remoteuser, remotehost, remotevol, nil
}

var georepCreateCmd = &cobra.Command{
	Use:   "create <master-volume> [<remote-user>@]<remote-host>::<remote-volume>",
	Short: helpGeorepCreateCmd,
	Args:  cobra.ExactArgs(2),
	Run: func(cmd *cobra.Command, args []string) {
		volname := args[0]
		remoteuser, remotehost, remotevol, err := parseRemoteData(args[1])
		if err != nil {
			failure(errGeorepSessionCreationFailed, err, 1)
		}

		masterdata, err := getVolumeDetails(volname, nil)
		if err != nil {
			failure(errGeorepSessionCreationFailed, err, 1)
		}

		remoteEndpoint, rclient, err := getRemoteClient(remotehost)
		if err != nil {
			failure(errGeorepSessionCreationFailed, err, 1)
		}
		remotevoldata, err := getVolumeDetails(remotevol, rclient)
		if err != nil {
			handleGlusterdConnectFailure(errGeorepSessionCreationFailed, remoteEndpoint, err, 1)

			// If not Glusterd connect Failure
			failure(errGeorepSessionCreationFailed, err, 1)
		}

		// Generate SSH Keys from all nodes of Master Volume
		sshkeys, err := client.GeorepSSHKeysGenerate(volname)
		if err != nil {
			if GlobalFlag.Verbose {
				log.WithError(err).WithField("volume", volname).Error("failed to generate SSH Keys")
			}
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
			if GlobalFlag.Verbose {
				log.WithError(err).WithField("volume", volname).Error("georep session creation failed")
			}
			failure(errGeorepSessionCreationFailed, err, 1)
		}

		err = rclient.GeorepSSHKeysPush(remotevol, sshkeys)
		if err != nil {
			if GlobalFlag.Verbose {
				log.WithError(err).WithField("volume", volname).Error("failed to push SSH Keys to Remote Cluster")
			}
			handleGlusterdConnectFailure(errGeorepSessionCreationFailed, remoteEndpoint, err, 1)

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
	masterVolID, remoteVolID, err := getVolIDs(args)
	if err != nil {
		failure(fmt.Sprintf("Geo-replication %s failed.\n", action.String()), err, 1)
	}
	switch action {
	case georepStart:
		_, err = client.GeorepStart(masterVolID, remoteVolID, flagGeorepCmdForce)
	case georepStop:
		_, err = client.GeorepStop(masterVolID, remoteVolID, flagGeorepCmdForce)
	case georepPause:
		_, err = client.GeorepPause(masterVolID, remoteVolID, flagGeorepCmdForce)
	case georepResume:
		_, err = client.GeorepResume(masterVolID, remoteVolID, flagGeorepCmdForce)
	case georepDelete:
		err = client.GeorepDelete(masterVolID, remoteVolID, flagGeorepCmdForce)
	}

	if err != nil {
		if GlobalFlag.Verbose {
			log.WithError(err).WithField("volume", args[0]).Error("geo-replication", action.String(), "failed")
		}
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

func getRemoteClient(host string) (string, *restclient.Client, error) {
	// TODO: Handle Remote Cluster Authentication and certificates and URL scheme
	clienturl := flagGeorepRemoteEndpoints

	if flagGeorepRemoteEndpoints != "" {
		_, err := url.Parse(flagGeorepRemoteEndpoints)
		if err != nil {
			return "", nil, errors.New("failed to parse geo-replication remote endpoints")
		}
	} else {
		clienturl = fmt.Sprintf("%s://%s:%d", geoRepHTTPScheme, host, geoRepGlusterdPort)
	}
	client, err := restclient.New(clienturl, "", "", "", true)
	return clienturl, client, err
}

func getVolIDs(pargs []string) (string, string, error) {
	var (
		masterVolID string
		remoteVolID string
	)

	allSessions, err := client.GeorepStatus("", "")
	if err != nil {
		failure(errGeorepStatusCommandFailed, err, 1)
	}

	if len(pargs) >= 1 {
		for _, s := range allSessions {
			if s.MasterVol == pargs[0] {
				masterVolID = s.MasterID.String()
			}
		}
		if masterVolID == "" {
			return "", "", errors.New("failed to get master volume info")
		}
	}

	if len(pargs) >= 2 {
		_, remotehost, remotevol, err := parseRemoteData(pargs[1])
		if err != nil {
			return "", "", err
		}

		for _, s := range allSessions {
			if s.RemoteVol == remotevol {

				for _, host := range s.RemoteHosts {
					if host.Hostname == remotehost {
						remoteVolID = s.RemoteID.String()
					}
				}
			}
		}
		if remoteVolID == "" {
			return "", "", errors.New("failed to get remote volume info")
		}
	}
	return masterVolID, remoteVolID, nil
}

var georepStatusCmd = &cobra.Command{
	Use:   "status [<master-volume> [[<remote-user>@]<remote-host>::<remote-volume>]]",
	Short: helpGeorepStatusCmd,
	Args:  cobra.RangeArgs(0, 2),
	Run: func(cmd *cobra.Command, args []string) {
		var err error
		masterVolID, remoteVolID, err := getVolIDs(args)
		if err != nil {
			failure(errGeorepStatusCommandFailed, err, 1)
		}

		var sessions []georepapi.GeorepSession
		// If masterVolID or remoteVolID is empty then get status of all and then filter
		if masterVolID == "" || remoteVolID == "" {
			allSessions, err := client.GeorepStatus("", "")
			if err != nil {
				failure(errGeorepStatusCommandFailed, err, 1)
			}
			for _, s := range allSessions {
				if masterVolID != "" && s.MasterID.String() != masterVolID {
					continue
				}
				if remoteVolID != "" && s.RemoteID.String() != remoteVolID {
					continue
				}
				sessionDetail, err := client.GeorepStatus(s.MasterID.String(), s.RemoteID.String())
				if err != nil {
					failure(errGeorepStatusCommandFailed, err, 1)
				}
				sessions = append(sessions, sessionDetail[0])
			}
		} else {
			sessions, err = client.GeorepStatus(masterVolID, remoteVolID)
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
						worker.MasterPeerHostname + ":" + worker.MasterBrickPath,
						worker.Status,
						worker.CrawlStatus,
						worker.RemotePeerHostname,
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

		masterVolID, remoteVolID, err := getVolIDs(args)
		if err != nil {
			failure("Error getting Volume IDs", err, 1)
		}

		opts, err := client.GeorepGet(masterVolID, remoteVolID)
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
		masterVolID, remoteVolID, err := getVolIDs(args)
		if err != nil {
			failure("Error getting Volume IDs", err, 1)
		}

		opts := make(map[string]string)
		opts[args[2]] = args[3]

		err = client.GeorepSet(masterVolID, remoteVolID, opts)
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
		masterVolID, remoteVolID, err := getVolIDs(args)
		if err != nil {
			failure(err.Error(), err, 1)
		}

		err = client.GeorepReset(masterVolID, remoteVolID, args[2:])
		if err != nil {
			failure("Geo-replication session config reset failed", err, 1)
		}
		fmt.Println("Geo-replication session config reset successfully")
	},
}
