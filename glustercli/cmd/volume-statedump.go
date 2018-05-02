package cmd

import (
	"errors"
	"fmt"
	"strconv"
	"strings"

	"github.com/gluster/glusterd2/pkg/api"

	"github.com/spf13/cobra"
)

var (
	flagStatedumpClient string
	flagStatedumpQuotad bool

	volumeStatedumpCmd = &cobra.Command{
		Use:   "statedump <volname> [--quota] [--client=<host>:<pid>]",
		Short: "Generate statedump of of a volume",
		Long:  "Generate statedump of various processes (bricks, client, quota) of a volume. Takes statedump of all bricks by default.",
		Args:  volumeStatedumpCmdArgs,
		Run:   volumeStatedumpCmdRun,
	}
)

func init() {
	volumeStatedumpCmd.Flags().StringVar(&flagStatedumpClient, "client", "", "client process in the format <ip>:<pid>")
	volumeStatedumpCmd.Flags().BoolVar(&flagStatedumpQuotad, "quota", false, "generate statedump of quotad process")
	volumeCmd.AddCommand(volumeStatedumpCmd)
}

func volumeStatedumpCmdArgs(cmd *cobra.Command, args []string) error {

	if len(args) != 1 {
		return errors.New("need exactly one argument i.e name of the volume")
	}

	if _, err := cmd.Flags().GetBool("quota"); err != nil {
		return err
	}

	if cmd.Flags().Changed("client") {
		client, err := cmd.Flags().GetString("client")
		if err != nil {
			return err
		}

		s := strings.Split(client, ":")
		if len(s) != 2 {
			return errors.New("client must be specified in the format <ip>:<pid>")
		}

		pid, err := strconv.Atoi(s[1])
		if err != nil || pid < 0 {
			return fmt.Errorf("invalid pid specified: %s", s[1])
		}
	}

	return nil
}

func volumeStatedumpCmdRun(cmd *cobra.Command, args []string) {

	var req api.VolStatedumpReq

	if cmd.Flags().Changed("quota") {
		req.Quota, _ = cmd.Flags().GetBool("quota")
	} else if cmd.Flags().Changed("client") {
		// validation is already done in volumeStatedumpCmdArgs()
		s := strings.Split(cmd.Flag("client").Value.String(), ":")
		pid, _ := strconv.Atoi(s[1])
		req.Client.Host = s[0]
		req.Client.Pid = pid
	} else {
		// TODO: The REST API doesn't support taking statedump of a single specified brick.
		req.Bricks = true
	}

	volname := args[0]
	if err := client.VolumeStatedump(volname, req); err != nil {
		fmt.Println(err)
	}
}
