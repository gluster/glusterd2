package e2e

import (
	"testing"

	tracemgmtapi "github.com/gluster/glusterd2/plugins/tracemgmt/api"
	"github.com/stretchr/testify/require"
)

const (
	jaegerEndpoint      = "http://localhost:14268"
	jaegerAgentEndpoint = "http://localhost:6831"
)

func TestTraceEnableAlwaysSample(t *testing.T) {
	r := require.New(t)

	tc, err := setupCluster(t, "./config/1.toml", "./config/2.toml")
	r.Nil(err)
	defer teardownCluster(tc)

	client, err := initRestclient(tc.gds[0])
	r.Nil(err)
	r.NotNil(client)

	jaegerSampler := 1 // Always sample
	reqTraceEnable := tracemgmtapi.SetupTracingReq{
		JaegerEndpoint:      jaegerEndpoint,
		JaegerAgentEndpoint: jaegerAgentEndpoint,
		JaegerSampler:       jaegerSampler,
	}
	jaegerCfgInfo, err := client.TraceEnable(reqTraceEnable)
	r.Nil(err)

	r.Equal(tracemgmtapi.TracingEnabled, jaegerCfgInfo.Status)
	r.Equal(jaegerEndpoint, jaegerCfgInfo.JaegerEndpoint)
	r.Equal(jaegerAgentEndpoint, jaegerCfgInfo.JaegerAgentEndpoint)
	r.Equal(jaegerSampler, jaegerCfgInfo.JaegerSampler)
	r.Equal(1.0, jaegerCfgInfo.JaegerSampleFraction)
}

func TestTraceEnableProbSampler(t *testing.T) {
	r := require.New(t)

	tc, err := setupCluster(t, "./config/1.toml", "./config/2.toml")
	r.Nil(err)
	defer teardownCluster(tc)

	client, err := initRestclient(tc.gds[0])
	r.Nil(err)
	r.NotNil(client)

	// Provide valid endpoint
	jaegerSampler := 2 // Probabilistic
	reqTraceEnable := tracemgmtapi.SetupTracingReq{
		JaegerEndpoint:      jaegerEndpoint,
		JaegerAgentEndpoint: jaegerAgentEndpoint,
		JaegerSampler:       jaegerSampler,
	}
	jaegerCfgInfo, err := client.TraceEnable(reqTraceEnable)
	r.Nil(err)

	// Get the trace config
	jaegerCfgInfo, err = client.TraceStatus()
	r.Nil(err)

	r.Equal(tracemgmtapi.TracingEnabled, jaegerCfgInfo.Status)
	r.Equal(jaegerEndpoint, jaegerCfgInfo.JaegerEndpoint)
	r.Equal(jaegerAgentEndpoint, jaegerCfgInfo.JaegerAgentEndpoint)
	r.Equal(jaegerSampler, jaegerCfgInfo.JaegerSampler)
	// The default sample fraction should be applied
	r.Equal(0.1, jaegerCfgInfo.JaegerSampleFraction)
}

func TestTraceEnableInvalidEndpoint(t *testing.T) {
	r := require.New(t)

	tc, err := setupCluster(t, "./config/1.toml", "./config/2.toml")
	r.Nil(err)
	defer teardownCluster(tc)

	client, err := initRestclient(tc.gds[0])
	r.Nil(err)
	r.NotNil(client)

	// Provide an invalid endpoint
	reqTraceEnable := tracemgmtapi.SetupTracingReq{
		JaegerEndpoint:      "",
		JaegerAgentEndpoint: jaegerAgentEndpoint,
	}
	jaegerCfgInfo, err := client.TraceEnable(reqTraceEnable)
	r.NotNil(err)

	// Get the trace config
	jaegerCfgInfo, err = client.TraceStatus()
	r.Nil(err)

	r.Equal(tracemgmtapi.TracingDisabled, jaegerCfgInfo.Status)
	r.Equal("", jaegerCfgInfo.JaegerEndpoint)
	r.Equal("", jaegerCfgInfo.JaegerAgentEndpoint)
	r.Equal(0, jaegerCfgInfo.JaegerSampler)
	r.Equal(0.0, jaegerCfgInfo.JaegerSampleFraction)
}

func TestTraceEnableInvalidSampler(t *testing.T) {
	r := require.New(t)

	tc, err := setupCluster(t, "./config/1.toml", "./config/2.toml")
	r.Nil(err)
	defer teardownCluster(tc)

	client, err := initRestclient(tc.gds[0])
	r.Nil(err)
	r.NotNil(client)

	// Provide invalid sampler
	jaegerSampler := 3
	reqTraceEnable := tracemgmtapi.SetupTracingReq{
		JaegerEndpoint:      jaegerEndpoint,
		JaegerAgentEndpoint: jaegerAgentEndpoint,
		JaegerSampler:       jaegerSampler,
	}
	jaegerCfgInfo, err := client.TraceEnable(reqTraceEnable)
	r.Nil(err)

	r.Equal(tracemgmtapi.TracingDisabled, jaegerCfgInfo.Status)
	r.Equal(jaegerEndpoint, jaegerCfgInfo.JaegerEndpoint)
	r.Equal(jaegerAgentEndpoint, jaegerCfgInfo.JaegerAgentEndpoint)
	r.Equal(0, jaegerCfgInfo.JaegerSampler)
	r.Equal(0.0, jaegerCfgInfo.JaegerSampleFraction)
}

func TestTraceEnableInvalidSampleFraction(t *testing.T) {
	r := require.New(t)

	tc, err := setupCluster(t, "./config/1.toml", "./config/2.toml")
	r.Nil(err)
	defer teardownCluster(tc)

	client, err := initRestclient(tc.gds[0])
	r.Nil(err)
	r.NotNil(client)

	// Provide an invalid sample fraction
	jaegerSampler := 2          // probabilistic
	jaegerSampleFraction := 1.1 // Invalid value

	reqTraceEnable := tracemgmtapi.SetupTracingReq{
		JaegerEndpoint:       jaegerEndpoint,
		JaegerAgentEndpoint:  jaegerAgentEndpoint,
		JaegerSampler:        jaegerSampler,
		JaegerSampleFraction: jaegerSampleFraction,
	}
	jaegerCfgInfo, err := client.TraceEnable(reqTraceEnable)
	r.Nil(err)

	// Get the trace config
	jaegerCfgInfo, err = client.TraceStatus()
	r.Nil(err)

	r.Equal(tracemgmtapi.TracingEnabled, jaegerCfgInfo.Status)
	r.Equal(jaegerEndpoint, jaegerCfgInfo.JaegerEndpoint)
	r.Equal(jaegerAgentEndpoint, jaegerCfgInfo.JaegerAgentEndpoint)
	r.Equal(jaegerSampler, jaegerCfgInfo.JaegerSampler)
	r.Equal(0.1, jaegerCfgInfo.JaegerSampleFraction)
}

func TestTraceUpdate(t *testing.T) {
	r := require.New(t)

	tc, err := setupCluster(t, "./config/1.toml", "./config/2.toml")
	r.Nil(err)
	defer teardownCluster(tc)

	client, err := initRestclient(tc.gds[0])
	r.Nil(err)
	r.NotNil(client)

	jaegerSampler := 1
	reqTraceEnable := tracemgmtapi.SetupTracingReq{
		JaegerEndpoint:      jaegerEndpoint,
		JaegerAgentEndpoint: jaegerAgentEndpoint,
		JaegerSampler:       jaegerSampler,
	}
	jaegerCfgInfo, err := client.TraceEnable(reqTraceEnable)
	r.Nil(err)

	r.Equal(tracemgmtapi.TracingEnabled, jaegerCfgInfo.Status)
	r.Equal(jaegerEndpoint, jaegerCfgInfo.JaegerEndpoint)
	r.Equal(jaegerAgentEndpoint, jaegerCfgInfo.JaegerAgentEndpoint)
	r.Equal(jaegerSampler, jaegerCfgInfo.JaegerSampler)
	r.Equal(1.0, jaegerCfgInfo.JaegerSampleFraction)

	// Send another request to update the trace options
	newJaegerEndpoint := "http://newhost:14268"
	newJaegerAgentEndpoint := "http://newhost:6831"
	newJaegerSampler := 2
	newJaegerSampleFraction := 0.7

	reqTraceUpdate := tracemgmtapi.SetupTracingReq{
		JaegerEndpoint:       newJaegerEndpoint,
		JaegerAgentEndpoint:  newJaegerAgentEndpoint,
		JaegerSampler:        newJaegerSampler,
		JaegerSampleFraction: newJaegerSampleFraction,
	}
	_, err = client.TraceUpdate(reqTraceUpdate)
	r.Nil(err)

	// Get the trace config
	jaegerCfgInfo, err = client.TraceStatus()
	r.Nil(err)
	r.Equal(tracemgmtapi.TracingEnabled, jaegerCfgInfo.Status)
	r.Equal(newJaegerEndpoint, jaegerCfgInfo.JaegerEndpoint)
	r.Equal(newJaegerAgentEndpoint, jaegerCfgInfo.JaegerAgentEndpoint)
	r.Equal(newJaegerSampler, jaegerCfgInfo.JaegerSampler)
	r.Equal(newJaegerSampleFraction, jaegerCfgInfo.JaegerSampleFraction)
}

func TestTraceUpdateInvalidOption(t *testing.T) {
	r := require.New(t)

	tc, err := setupCluster(t, "./config/1.toml", "./config/2.toml")
	r.Nil(err)
	defer teardownCluster(tc)

	client, err := initRestclient(tc.gds[0])
	r.Nil(err)
	r.NotNil(client)

	// Enable tracing with all valid options
	jaegerSampler := 1
	reqTraceEnable := tracemgmtapi.SetupTracingReq{
		JaegerEndpoint:      jaegerEndpoint,
		JaegerAgentEndpoint: jaegerAgentEndpoint,
		JaegerSampler:       jaegerSampler,
	}
	jaegerCfgInfo, err := client.TraceEnable(reqTraceEnable)
	r.Nil(err)

	r.Equal(tracemgmtapi.TracingEnabled, jaegerCfgInfo.Status)
	r.Equal(jaegerEndpoint, jaegerCfgInfo.JaegerEndpoint)
	r.Equal(jaegerAgentEndpoint, jaegerCfgInfo.JaegerAgentEndpoint)
	r.Equal(jaegerSampler, jaegerCfgInfo.JaegerSampler)
	r.Equal(1.0, jaegerCfgInfo.JaegerSampleFraction)

	// Test update with invalid sample fraction
	newJaegerSampler := 2
	newJaegerSampleFraction := 1.1

	reqTraceUpdate := tracemgmtapi.SetupTracingReq{
		JaegerSampler:        newJaegerSampler,
		JaegerSampleFraction: newJaegerSampleFraction,
	}
	_, err = client.TraceUpdate(reqTraceUpdate)
	r.Nil(err)

	// Get the trace config
	jaegerCfgInfo, err = client.TraceStatus()
	r.Nil(err)
	r.Equal(tracemgmtapi.TracingEnabled, jaegerCfgInfo.Status)
	r.Equal(jaegerEndpoint, jaegerCfgInfo.JaegerEndpoint)
	r.Equal(jaegerAgentEndpoint, jaegerCfgInfo.JaegerAgentEndpoint)
	r.Equal(newJaegerSampler, jaegerCfgInfo.JaegerSampler)
	// Since sample fraction is invalid, the default value will be set
	r.Equal(0.1, jaegerCfgInfo.JaegerSampleFraction)

	// Test update with invalid sampler type option
	newJaegerSampler = 3

	reqTraceUpdate = tracemgmtapi.SetupTracingReq{
		JaegerSampler: newJaegerSampler,
	}
	_, err = client.TraceUpdate(reqTraceUpdate)
	r.Nil(err)

	// Get the trace config
	jaegerCfgInfo, err = client.TraceStatus()
	r.Nil(err)
	r.Equal(tracemgmtapi.TracingDisabled, jaegerCfgInfo.Status)
	r.Equal(jaegerEndpoint, jaegerCfgInfo.JaegerEndpoint)
	r.Equal(jaegerAgentEndpoint, jaegerCfgInfo.JaegerAgentEndpoint)
	r.Equal(0, jaegerCfgInfo.JaegerSampler)
	r.Equal(0.0, jaegerCfgInfo.JaegerSampleFraction)
}
