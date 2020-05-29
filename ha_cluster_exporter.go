package main

import (
	"fmt"
	"net/http"
	"os"
	"runtime"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	log "github.com/sirupsen/logrus"
	flag "github.com/spf13/pflag"
	config "github.com/spf13/viper"

	"github.com/ClusterLabs/ha_cluster_exporter/collector/corosync"
	"github.com/ClusterLabs/ha_cluster_exporter/collector/drbd"
	"github.com/ClusterLabs/ha_cluster_exporter/collector/pacemaker"
	"github.com/ClusterLabs/ha_cluster_exporter/collector/sbd"
	"github.com/ClusterLabs/ha_cluster_exporter/internal"
)

var (
	// the released version
	version string
	// the time the binary was built
	buildDate string
	// global --help flag
	helpFlag *bool
	// global --version flag
	versionFlag *bool
)

func init() {
	config.SetConfigName("ha_cluster_exporter")
	config.AddConfigPath("./")
	config.AddConfigPath("$HOME/.config/")
	config.AddConfigPath("/etc/")
	config.AddConfigPath("/usr/etc/")

	flag.String("port", "9664", "The port number to listen on for HTTP requests")
	flag.String("address", "0.0.0.0", "The address to listen on for HTTP requests")
	flag.String("log-level", "info", "The minimum logging level; levels are, in ascending order: debug, info, warn, error")
	flag.String("crm-mon-path", "/usr/sbin/crm_mon", "path to crm_mon executable")
	flag.String("cibadmin-path", "/usr/sbin/cibadmin", "path to cibadmin executable")
	flag.String("corosync-cfgtoolpath-path", "/usr/sbin/corosync-cfgtool", "path to corosync-cfgtool executable")
	flag.String("corosync-quorumtool-path", "/usr/sbin/corosync-quorumtool", "path to corosync-quorumtool executable")
	flag.String("sbd-path", "/usr/sbin/sbd", "path to sbd executable")
	flag.String("sbd-config-path", "/etc/sysconfig/sbd", "path to sbd configuration")
	flag.String("drbdsetup-path", "/sbin/drbdsetup", "path to drbdsetup executable")
	flag.String("drbdsplitbrain-path", "/var/run/drbd/splitbrain", "path to drbd splitbrain hooks temporary files")
	flag.Bool("enable-timestamps", false, "Add the timestamp to every metric line")
	flag.CommandLine.SortFlags = false

	err := config.BindPFlags(flag.CommandLine)
	if err != nil {
		log.Fatalf("Could not bind config to CLI flags: %v", err)
	}

	helpFlag = flag.BoolP("help", "h", false, "show this help message")
	versionFlag = flag.Bool("version", false, "show version and build information")
}

func main() {
	flag.Parse()

	switch {
	case *helpFlag:
		showHelp()
	case *versionFlag:
		showVersion()
	default:
		run()
	}
}

func run() {
	var err error

	err = config.ReadInConfig()
	if err != nil {
		log.Warn(err)
		log.Info("Default config values will be used")
	} else {
		log.Info("Using config file: ", config.ConfigFileUsed())
	}

	internal.SetLogLevel(config.GetString("log-level"))

	pacemakerCollector, err := pacemaker.NewCollector(
		config.GetString("crm-mon-path"),
		config.GetString("cibadmin-path"),
	)
	if err != nil {
		log.Warn(err)
	} else {
		prometheus.MustRegister(pacemakerCollector)
		log.Info("Pacemaker collector registered")
	}

	corosyncCollector, err := corosync.NewCollector(
		config.GetString("corosync-cfgtoolpath-path"),
		config.GetString("corosync-quorumtool-path"),
	)
	if err != nil {
		log.Warn(err)
	} else {
		prometheus.MustRegister(corosyncCollector)
		log.Info("Corosync collector registered")
	}

	sbdCollector, err := sbd.NewCollector(
		config.GetString("sbd-path"),
		config.GetString("sbd-config-path"),
	)
	if err != nil {
		log.Warn(err)
	} else {
		prometheus.MustRegister(sbdCollector)
		log.Info("SBD collector registered")
	}

	drbdCollector, err := drbd.NewCollector(config.GetString("drbdsetup-path"), config.GetString("drbdsplitbrain-path"))
	if err != nil {
		log.Warn(err)
	} else {
		prometheus.MustRegister(drbdCollector)
		log.Info("DRBD collector registered")
	}

	// if we're not in debug log level, we unregister the Go runtime metrics collector that gets registered by default
	if !log.IsLevelEnabled(log.DebugLevel) {
		prometheus.Unregister(prometheus.NewGoCollector())
	}

	fullListenAddress := fmt.Sprintf("%s:%s", config.Get("address"), config.Get("port"))

	http.HandleFunc("/", internal.Landing)
	http.Handle("/metrics", promhttp.Handler())

	log.Infof("Serving metrics on %s", fullListenAddress)
	log.Fatal(http.ListenAndServe(fullListenAddress, nil))
}

func showHelp() {
	flag.Usage()
	os.Exit(0)
}

func showVersion() {
	if buildDate == "" {
		buildDate = "at unknown time"
	}
	fmt.Printf("version %s\nbuilt with %s %s/%s %s\n", version, runtime.Version(), runtime.GOOS, runtime.GOARCH, buildDate)
	os.Exit(0)
}
