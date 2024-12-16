package main

import (
	"context"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/url"
	"os"
	"os/exec"
	"reflect"
	"runtime"
	"strings"
	"syscall"
	"testing"
	"time"

	"github.com/go-kit/log"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/stretchr/testify/assert"
)

func TestRegisterCollectors(t *testing.T) {
	*haClusterCrmMonPath = "test/fake_crm_mon.sh"
	*haClusterCibadminPath = "test/fake_cibadmin.sh"
	*haClusterCorosyncCfgtoolpathPath = "test/fake_corosync-cfgtool.sh"
	*haClusterCorosyncQuorumtoolPath = "test/fake_corosync-quorumtool.sh"
	*haClusterSbdPath = "test/fake_sbd.sh"
	*haClusterSbdConfigPath = "test/fake_sbdconfig"
	*haClusterDrbdsetupPath = "test/fake_drbdsetup.sh"
	*haClusterDrbdsplitbrainPath = "test/fake_drbdsplitbrain"

	t.Run("success", func(t *testing.T) {
		wantCollectors := 4
		wantErrors := 0
		prometheus.DefaultRegisterer = prometheus.NewRegistry()
		prometheus.DefaultGatherer = prometheus.NewRegistry()
		collectors, errors := registerCollectors(log.NewNopLogger())
		assert.Len(t, collectors, wantCollectors)
		assert.Len(t, errors, wantErrors)
	})

	*haClusterCrmMonPath = "does_not_exist"
	t.Run("1 failure", func(t *testing.T) {
		wantCollectors := 3
		wantErrors := 1
		prometheus.DefaultRegisterer = prometheus.NewRegistry()
		prometheus.DefaultGatherer = prometheus.NewRegistry()
		collectors, errors := registerCollectors(log.NewNopLogger())
		assert.Len(t, collectors, wantCollectors)
		assert.Len(t, errors, wantErrors)
	})

	*haClusterCorosyncCfgtoolpathPath = "does_not_exist"
	t.Run("2 failures", func(t *testing.T) {
		wantCollectors := 2
		wantErrors := 2
		prometheus.DefaultRegisterer = prometheus.NewRegistry()
		prometheus.DefaultGatherer = prometheus.NewRegistry()
		collectors, errors := registerCollectors(log.NewNopLogger())
		assert.Len(t, collectors, wantCollectors)
		assert.Len(t, errors, wantErrors)
	})

	*haClusterSbdPath = "does_not_exist"
	t.Run("3 failures", func(t *testing.T) {
		wantCollectors := 1
		wantErrors := 3
		prometheus.DefaultRegisterer = prometheus.NewRegistry()
		prometheus.DefaultGatherer = prometheus.NewRegistry()
		collectors, errors := registerCollectors(log.NewNopLogger())
		assert.Len(t, collectors, wantCollectors)
		assert.Len(t, errors, wantErrors)
	})

	*haClusterDrbdsetupPath = "does_not_exist"
	t.Run("4 failures", func(t *testing.T) {
		wantCollectors := 0
		wantErrors := 4
		prometheus.DefaultRegisterer = prometheus.NewRegistry()
		prometheus.DefaultGatherer = prometheus.NewRegistry()
		collectors, errors := registerCollectors(log.NewNopLogger())
		assert.Len(t, collectors, wantCollectors)
		assert.Len(t, errors, wantErrors)
	})
}

//// Kudos for the build/run tests to https://github.com/prometheus/mysqld_exporter
// TestBin builds, runs and tests binary.

// bin stores information about path of executable and attached port
type bin struct {
	path string
	port int
}

func TestBin(t *testing.T) {
	var err error
	binName := "ha"

	binDir, err := os.MkdirTemp("/tmp", binName+"-test-bindir-")
	if err != nil {
		t.Fatal(err)
	}
	defer func() {
		err := os.RemoveAll(binDir)
		if err != nil {
			t.Fatal(err)
		}
	}()

	importpath := "github.com/prometheus/ha_cluster_exporter/vendor/github.com/prometheus/common"
	path := binDir + "/" + binName
	xVariables := map[string]string{
		importpath + "/version.Version":  "gotest-version",
		importpath + "/version.Branch":   "gotest-branch",
		importpath + "/version.Revision": "gotest-revision",
	}
	var ldflags []string
	for x, value := range xVariables {
		ldflags = append(ldflags, fmt.Sprintf("-X %s=%s", x, value))
	}
	cmd := exec.Command(
		"go",
		"build",
		"-o",
		path,
		"-ldflags",
		strings.Join(ldflags, " "),
	)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	err = cmd.Run()
	if err != nil {
		t.Fatalf("Failed to build: %s", err)
	}

	tests := []func(*testing.T, bin){
		testLandingPage,
	}

	portStart := 56000
	t.Run(binName, func(t *testing.T) {
		for _, f := range tests {
			f := f // capture range variable
			fName := runtime.FuncForPC(reflect.ValueOf(f).Pointer()).Name()
			portStart++
			data := bin{
				path: path,
				port: portStart,
			}
			t.Run(fName, func(t *testing.T) {
				t.Parallel()
				f(t, data)
			})
		}
	})
}

func testLandingPage(t *testing.T, data bin) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// Run exporter.
	servePath := "/metrics"
	cmd := exec.CommandContext(
		ctx,
		data.path,
		"--web.listen-address", fmt.Sprintf(":%d", data.port),
		"--web.telemetry-path", fmt.Sprintf("%s", servePath),
		"--crm-mon-path=test/fake_crm_mon.sh", // needed to register at least one collector
		"--cibadmin-path=test/fake_cibadmin.sh",
	)
	if err := cmd.Start(); err != nil {
		t.Fatal(err)
	}
	defer cmd.Wait()
	defer cmd.Process.Kill()

	// Get the main page.
	urlToGet := fmt.Sprintf("http://127.0.0.1:%d", data.port)
	body, err := waitForBody(urlToGet)
	if err != nil {
		t.Fatal(err)
	}
	got := string(body)
	expected := `<html>
<head>
	<title>ClusterLabs Linux HA Cluster Exporter</title>
</head>
<body>
	<h1>ClusterLabs Linux HA Cluster Exporter</h1>
	<h2>Prometheus exporter for Pacemaker based Linux HA clusters</h2>
	<ul>
		<li><a href="` + servePath + `">Metrics</a></li>
		<li><a href="https://github.com/ClusterLabs/ha_cluster_exporter" target="_blank">GitHub</a></li>
	</ul>
</body>
</html>
`

	if got != expected {
		t.Fatalf("got '%s' but expected '%s'", got, expected)
	}
}

// waitForBody is a helper function which makes http calls until http server is up
// and then returns body of the successful call.
func waitForBody(urlToGet string) (body []byte, err error) {
	tries := 60

	// Get data, but we need to wait a bit for http server.
	for i := 0; i <= tries; i++ {
		// Try to get web page.
		body, err = getBody(urlToGet)
		if err == nil {
			return body, err
		}

		// If there is a syscall.ECONNREFUSED error (web server not available) then retry.
		if urlError, ok := err.(*url.Error); ok {
			if opError, ok := urlError.Err.(*net.OpError); ok {
				if osSyscallError, ok := opError.Err.(*os.SyscallError); ok {
					if osSyscallError.Err == syscall.ECONNREFUSED {
						time.Sleep(1 * time.Second)
						continue
					}
				}
			}
		}

		// There was an error, and it wasn't syscall.ECONNREFUSED.
		return nil, err
	}

	return nil, fmt.Errorf("failed to GET %s after %d tries: %s", urlToGet, tries, err)
}

// getBody is a helper function which retrieves http body from given address.
func getBody(urlToGet string) ([]byte, error) {
	resp, err := http.Get(urlToGet)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	return body, nil
}
