package main

import (
	"testing"

	"github.com/go-kit/log"
	"github.com/prometheus/client_golang/prometheus"
	// We could also mock this but test files alrady exist in test dir
	//"github.com/spf13/afero"
	"github.com/stretchr/testify/assert"
)

func TestRegisterCollectors(t *testing.T) {
	// We could also mock this but test files alrady exist in test dir
	// filesystem as os.Stat() is run in default_collector.go
	// fs := afero.NewMemMapFs() # does not work
	//fs := afero.NewOsFs()
	//fs.MkdirAll("test/bin", 0755)
	//afero.WriteFile(fs, "test/bin/crm-mon-path", []byte(""), 0755)
	//afero.WriteFile(fs, "test/bin/cibadmin-path", []byte(""), 0755)
	//afero.WriteFile(fs, "test/bin/corosync-cfgtoolpath-path", []byte(""), 0755)
	//afero.WriteFile(fs, "test/bin/corosync-quorumtool-path", []byte(""), 0755)
	//afero.WriteFile(fs, "test/bin/sbd-path", []byte(""), 0755)
	//afero.WriteFile(fs, "test/bin/sbd-config-path", []byte(""), 0755)
	//afero.WriteFile(fs, "test/bin/drbdsetup-path", []byte(""), 0755)
	//afero.WriteFile(fs, "test/bin/drbdsplitbrain-path", []byte(""), 0755)
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
	//fs.RemoveAll("test/bin")
}