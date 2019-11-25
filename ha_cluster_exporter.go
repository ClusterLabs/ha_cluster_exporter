package main

import (
	"fmt"
	"net/http"
	"os"
	"sync"
	"time"

	"github.com/pkg/errors"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	log "github.com/sirupsen/logrus"
	flag "github.com/spf13/pflag"
	config "github.com/spf13/viper"
)

const NAMESPACE = "ha_cluster"

type Clock interface {
	Now() time.Time
}

type SystemClock struct{}

func (SystemClock) Now() time.Time {
	return time.Now()
}

type metricDescriptors map[string]*prometheus.Desc

type DefaultCollector struct {
	metrics metricDescriptors
	mutex   sync.RWMutex
}

func (c *DefaultCollector) Describe(ch chan<- *prometheus.Desc) {
	for _, metric := range c.metrics {
		ch <- metric
	}
}

func (c *DefaultCollector) makeGaugeMetric(metricKey string, value float64, labelValues ...string) prometheus.Metric {
	desc, ok := c.metrics[metricKey]
	if !ok {
		// we hard panic on this because it's most certainly a coding error
		panic(errors.Errorf("undeclared metric '%s'", metricKey))
	}
	return prometheus.NewMetricWithTimestamp(clock.Now(), prometheus.MustNewConstMetric(desc, prometheus.GaugeValue, value, labelValues...))
}

// Convenience wrapper around Prometheus constructors.
// Produces a metric with name `NAMESPACE_subsystem_name`.
// `NAMESPACE` is a global project constant;
// `subsystem` is an arbitrary name used to group related metrics under the same name prefix;
// `name` is the last and most relevant part of the metrics Full Qualified Name;
// `variableLabels` is a list of labels to declare. Use `nil` to declare no labels.
func NewMetricDesc(subsystem, name, help string, variableLabels []string) *prometheus.Desc {
	return prometheus.NewDesc(prometheus.BuildFQName(NAMESPACE, subsystem, name), help, variableLabels, nil)
}

// check that all the given paths exist and are executable files
func CheckExecutables(paths ...string) error {
	for _, path := range paths {
		fileInfo, err := os.Stat(path)
		if err != nil || os.IsNotExist(err) {
			return errors.Errorf("'%s' does not exist", path)
		}
		if fileInfo.IsDir() {
			return errors.Errorf("'%s' is a directory", path)
		}
		if (fileInfo.Mode() & 0111) == 0 {
			return errors.Errorf("'%s' is not executable", path)
		}
	}
	return nil
}

// Landing Page (for /)
func landingpage(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte(`<html>
		<head>
			<title>HACluster Exporter</title>
		</head>
		<body>
			<h1>HACluster Exporter</h1>
			<p><a href="metrics">Metrics</a></p>
			<br />
			<h2>More information:</h2>
			<p><a href="https://github.com/ClusterLabs/ha_cluster_exporter">github.com/ClusterLabs/ha_cluster_exporter</a></p>
		</body>
		</html>`))
}

func loglevel(level string) {
	switch level {
	case "error":
		log.SetLevel(log.ErrorLevel)
	case "warn":
		log.SetLevel(log.WarnLevel)
	case "info":
		log.SetLevel(log.InfoLevel)
	case "debug":
		log.SetLevel(log.DebugLevel)
	default:
		log.Warnln("Unrecognized minimum log level; using 'info' as default")
	}
}

var clock Clock = &SystemClock{}

func init() {
	config.SetConfigName("ha_cluster_exporter")
	config.AddConfigPath("./")
	config.AddConfigPath("$HOME/.config/")
	config.AddConfigPath("/etc/")
	config.AddConfigPath("/usr/etc/")

	flag.String("port", "9002", "The port number to listen on for HTTP requests")
	flag.String("address", "0.0.0.0", "The address to listen on for HTTP requests")
	flag.String("log-level", "info", "The minimum logging level; levels are, in ascending order: debug, info, warn, error")
	flag.String("crm-mon-path", "/usr/sbin/crm_mon", "path to crm_mon executable")
	flag.String("cibadmin-path", "/usr/sbin/cibadmin", "path to cibadmin executable")
	flag.String("corosync-cfgtoolpath-path", "/usr/sbin/corosync-cfgtool", "path to corosync-cfgtool executable")
	flag.String("corosync-quorumtool-path", "/usr/sbin/corosync-quorumtool", "path to corosync-quorumtool executable")
	flag.String("sbd-path", "/usr/sbin/sbd", "path to sbd executable")
	flag.String("sbd-config-path", "/etc/sysconfig/sbd", "path to sbd configuration")
	flag.String("drbdsetup-path", "/usr/sbin/drbdsetup", "path to drbdsetup executable")
	flag.String("drbdsplitbrain-path", "/var/run/drbd/splitbrain", "path to drbd splitbrain hooks temporary files")

	err := config.BindPFlags(flag.CommandLine)
	if err != nil {
		log.Errorf("Could not bind config to CLI flags: %v", err)
	}
}

func main() {
	var err error

	flag.Parse()

	err = config.ReadInConfig()
	if err != nil {
		log.Warn(err)
		log.Info("Default config values will be used")
	} else {
		log.Info("Using config file: ", config.ConfigFileUsed())
	}

	loglevel(config.GetString("log-level"))

	pacemakerCollector, err := NewPacemakerCollector(
		config.GetString("crm-mon-path"),
		config.GetString("cibadmin-path"),
	)
	if err != nil {
		log.Warn(err)
	} else {
		prometheus.MustRegister(pacemakerCollector)
		log.Info("Pacemaker collector registered")
	}

	corosyncCollector, err := NewCorosyncCollector(
		config.GetString("corosync-cfgtoolpath-path"),
		config.GetString("corosync-quorumtool-path"),
	)
	if err != nil {
		log.Warn(err)
	} else {
		prometheus.MustRegister(corosyncCollector)
		log.Info("Corosync collector registered")
	}

	sbdCollector, err := NewSbdCollector(
		config.GetString("sbd-path"),
		config.GetString("sbd-config-path"),
	)
	if err != nil {
		log.Warn(err)
	} else {
		prometheus.MustRegister(sbdCollector)
		log.Info("SBD collector registered")
	}

	drbdCollector, err := NewDrbdCollector(config.GetString("drbdsetup-path"), config.GetString("drbdsplitbrain-path"))
	if err != nil {
		log.Warn(err)
	} else {
		prometheus.MustRegister(drbdCollector)
		log.Info("DRBD collector registered")
	}

	fullListenAddress := fmt.Sprintf("%s:%s", config.Get("address"), config.Get("port"))

	http.HandleFunc("/", landingpage)
	http.Handle("/metrics", promhttp.Handler())

	log.Infof("Serving metrics on %s", fullListenAddress)
	log.Fatal(http.ListenAndServe(fullListenAddress, nil))
}
