package main

import (
	"fmt"
	"os"
	"os/exec"
	"regexp"
	"syscall"

	"github.com/gluster/glusterd2/glustercli/cmd"
	"github.com/gluster/glusterd2/pkg/logging"
)

var unknownCommandExp = regexp.MustCompile(`unknown command "([^"]+)" `)

func getUnknownCommandName(err error) string {
	m := unknownCommandExp.FindStringSubmatch(err.Error())
	if len(m) == 2 {
		return m[1]
	}
	return ""
}

func main() {
	// Migrate old format Args into new Format. Modifies os.Args[]
	argsMigrate()

	cmd.RootCmd.SilenceErrors = true

	if err := cmd.RootCmd.Execute(); err != nil {
		unknownCmd := getUnknownCommandName(err)
		if unknownCmd != "" {
			// May be external sub Command
			cmdPath, err := exec.LookPath(cmd.RootCmd.Use + "-" + unknownCmd)
			if err == nil {
				cmdArgs := []string{}
				err = cmd.RootCmd.ParseFlags(os.Args[1:])
				if err == nil {
					if cmd.FlagXMLOutput {
						cmdArgs = append(cmdArgs, "--xml")
					}

					if cmd.FlagJSONOutput {
						cmdArgs = append(cmdArgs, "--json")
					}
					cmdArgs = append(cmdArgs, "--glusterd-host", cmd.FlagHostname)
					if cmd.FlagHTTPS {
						cmdArgs = append(cmdArgs, "--glusterd-https")
					}
					cmdArgs = append(cmdArgs, "--glusterd-port", fmt.Sprintf("%d", cmd.FlagPort))
					cmdArgs = append(cmdArgs, "--"+logging.DirFlag, cmd.FlagLogDir)
					cmdArgs = append(cmdArgs, "--"+logging.FileFlag, cmd.FlagLogFile)
					cmdArgs = append(cmdArgs, "--"+logging.LevelFlag, cmd.FlagLogLevel)
					cmdArgs = append(cmdArgs, "--cacert", cmd.FlagCacert)
					if cmd.FlagInsecure {
						cmdArgs = append(cmdArgs, "--insecure")
					}
				}

				command := exec.Command(cmdPath, cmdArgs...)
				command.Stdout = os.Stdout
				command.Stderr = os.Stderr

				// If Return code is accessible
				if err = command.Run(); err != nil {
					if exiterr, ok := err.(*exec.ExitError); ok {
						if status, ok := exiterr.Sys().(syscall.WaitStatus); ok {
							os.Exit(status.ExitStatus())
						}
					}
					os.Exit(1)
				}
				return
			}
		}

		// If Known command error or no external command available
		fmt.Println("Error:", err)
		os.Exit(1)
	}
}
