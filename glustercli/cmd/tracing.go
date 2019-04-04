package cmd

import (
	"errors"
	"fmt"
	"os"

	"github.com/gluster/glusterd2/pkg/tracing"
	tracemgmtapi "github.com/gluster/glusterd2/plugins/tracemgmt/api"

	"github.com/olekukonko/tablewriter"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

const (
	helpTraceCmd            = "Gluster Trace Management"
	helpTraceEnableCmd      = "Enable Tracing"
	helpTraceStatusCmd      = "Show Tracing Status"
	helpTraceUpdateCmd      = "Update Tracing Options"
	helpTraceDisableCmd     = "Disable Tracing"
	errTraceEnableReqFailed = "Failed to enable tracing"
	errTraceUpdateReqFailed = "Failed to update trace configuration"
)

var (
	flagTraceJaegerEndpoint       string
	flagTraceJaegerAgentEndpoint  string
	flagTraceJaegerSampler        int
	flagTraceJaegerSampleFraction float64
)

func init() {
	// Trace Enable
	traceEnableCmd.Flags().IntVar(&flagTraceJaegerSampler, "jaeger-sampler", 2, "jaeger sampler (1: Always sample, 2: Probabilistic)")
	traceEnableCmd.Flags().Float64Var(&flagTraceJaegerSampleFraction, "jaeger-sample-fraction", 0.1, "jaeger sample fraction (min: >0.0, max: <1.0)")
	// Trace Update
	traceUpdateCmd.Flags().StringVar(&flagTraceJaegerEndpoint, "jaeger-endpoint", "", "jaeger endpoint")
	traceUpdateCmd.Flags().StringVar(&flagTraceJaegerAgentEndpoint, "jaeger-agent-endpoint", "", "jaeger agent endpoint")
	traceUpdateCmd.Flags().IntVar(&flagTraceJaegerSampler, "jaeger-sampler", 2, "jaeger sampler (1: Always sample, 2: Probabilistic)")
	traceUpdateCmd.Flags().Float64Var(&flagTraceJaegerSampleFraction, "jaeger-sample-fraction", 0.1, "jaeger sample fraction (min: >0.0, max: <1.0)")
	traceCmd.AddCommand(traceEnableCmd)
	traceCmd.AddCommand(traceStatusCmd)
	traceCmd.AddCommand(traceUpdateCmd)
	traceCmd.AddCommand(traceDisableCmd)
}

func validateJaegerSampler() {
	if flagTraceJaegerSampler <= 0 || flagTraceJaegerSampler > 2 {
		err := errors.New("invalid Jaeger sampler specified")
		failure(errTraceEnableReqFailed, err, 1)
	}

	if flagTraceJaegerSampler == 2 {
		if flagTraceJaegerSampleFraction <= 0.0 || flagTraceJaegerSampleFraction >= 1.0 {
			err := errors.New("invalid Jaeger sample fraction specified")
			failure(errTraceEnableReqFailed, err, 1)
		}
	}
}

var traceCmd = &cobra.Command{
	Use:   "trace",
	Short: helpTraceCmd,
}

var traceEnableCmd = &cobra.Command{
	Use:   "enable <jaeger-endpoint> <jaeger-agent-endpoint>",
	Short: helpTraceEnableCmd,
	Args:  cobra.ExactArgs(2),
	Run: func(cmd *cobra.Command, args []string) {
		jaegerEndpoint := args[0]
		jaegerAgentEndpoint := args[1]

		if jaegerEndpoint == "" || jaegerAgentEndpoint == "" {
			err := errors.New("one or more Jaeger endpoints not specified")
			failure(errTraceEnableReqFailed, err, 1)
		}

		validateJaegerSampler()

		_, err := client.TraceEnable(tracemgmtapi.SetupTracingReq{
			JaegerEndpoint:       jaegerEndpoint,
			JaegerAgentEndpoint:  jaegerAgentEndpoint,
			JaegerSampler:        flagTraceJaegerSampler,
			JaegerSampleFraction: flagTraceJaegerSampleFraction,
		})

		if err != nil {
			if GlobalFlag.Verbose {
				log.WithError(err).WithFields(log.Fields{
					"JaegerEndpoint":       jaegerEndpoint,
					"JaegerAgentEndpoint":  jaegerAgentEndpoint,
					"JaegerSampler":        flagTraceJaegerSampler,
					"JaegerSampleFraction": flagTraceJaegerSampleFraction,
				}).Error("trace Enable failed")
			}
			failure(errTraceEnableReqFailed, err, 1)
		}
		fmt.Println("Trace enable successful")
	},
}

var traceStatusCmd = &cobra.Command{
	Use:   "status",
	Short: helpTraceStatusCmd,
	Run: func(cmd *cobra.Command, args []string) {
		jaegerConfigInfo, err := client.TraceStatus()
		if err != nil {
			failure("Error getting trace status", err, 1)
		}

		table := tablewriter.NewWriter(os.Stdout)
		table.SetAlignment(tablewriter.ALIGN_LEFT)
		table.SetHeader([]string{"Trace Option", "Value"})

		table.Append([]string{"Status", jaegerConfigInfo.Status})
		table.Append([]string{"Jaeger Endpoint", jaegerConfigInfo.JaegerEndpoint})
		table.Append([]string{"Jaeger Agent Endpoint", jaegerConfigInfo.JaegerAgentEndpoint})
		jaegerSampler := tracing.JaegerSamplerType(jaegerConfigInfo.JaegerSampler)
		table.Append([]string{"Jaeger Sampler", fmt.Sprintf("%d (%s)", jaegerConfigInfo.JaegerSampler, tracing.SamplerTypeToString(jaegerSampler))})
		table.Append([]string{"Jaeger Sample Fraction", fmt.Sprintf("%0.2f", jaegerConfigInfo.JaegerSampleFraction)})
		table.Render()
		fmt.Println()
	},
}

var traceUpdateCmd = &cobra.Command{
	Use:   "update",
	Short: helpTraceUpdateCmd,
	Run: func(cmd *cobra.Command, args []string) {

		validateJaegerSampler()

		_, err := client.TraceUpdate(tracemgmtapi.SetupTracingReq{
			JaegerEndpoint:       flagTraceJaegerEndpoint,
			JaegerAgentEndpoint:  flagTraceJaegerAgentEndpoint,
			JaegerSampler:        flagTraceJaegerSampler,
			JaegerSampleFraction: flagTraceJaegerSampleFraction,
		})

		if err != nil {
			if GlobalFlag.Verbose {
				log.WithError(err).WithFields(log.Fields{
					"JaegerEndpoint":       flagTraceJaegerEndpoint,
					"JaegerAgentEndpoint":  flagTraceJaegerAgentEndpoint,
					"JaegerSampler":        flagTraceJaegerSampler,
					"JaegerSampleFraction": flagTraceJaegerSampleFraction,
				}).Error("Trace Update failed")
			}
			failure(errTraceUpdateReqFailed, err, 1)
		}
		fmt.Println("Trace update successful")
	},
}

var traceDisableCmd = &cobra.Command{
	Use:   "disable",
	Short: helpTraceDisableCmd,
	Run: func(cmd *cobra.Command, args []string) {
		err := client.TraceDisable()
		if err != nil {
			failure("Trace disable failed", err, 1)
		}
		fmt.Println("Trace disable successful")
	},
}
