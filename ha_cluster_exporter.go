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
	"github.com/spf13/viper"
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

func CheckExecutables(paths ...string) error {
	// check that all executables we rely on exist and are executables
	for _, path := range paths {
		fileInfo, err := os.Stat(path)
		if err != nil || os.IsNotExist(err) {
			return err
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

func loglevel(opt string) {
	switch opt {
	case "error":
		log.SetLevel(log.ErrorLevel)
	case "warn":
		log.SetLevel(log.WarnLevel)
	case "info":
		log.SetLevel(log.InfoLevel)
	case "debug":
		log.SetLevel(log.DebugLevel)
	default:
		log.Warnln("Unrecognized log level, default to `info` level")
	}
}

var (
	clock      Clock = &SystemClock{}
	portNumber       = flag.String("port", "9002", "The port number to listen on for HTTP requests.")
	address          = flag.String("address", "0.0.0.0", "The address to listen on for HTTP requests.")
	logLevel         = flag.String("level", "info", "The level of logs to log")
)

func init() {
	viper.SetConfigName("ha_cluster_exporter")
	viper.AddConfigPath("/etc/")
	viper.AddConfigPath("$HOME/.config")
	viper.AddConfigPath(".")

	viper.SetDefault("crm_mon_path", "/usr/sbin/crm_mon")
	viper.SetDefault("cibadmin_path", "/usr/sbin/cibadmin")
	viper.SetDefault("corosync_cfgtoolpath_path", "/usr/sbin/corosync-cfgtool")
	viper.SetDefault("corosync_quorumtool_path", "/usr/sbin/corosync-quorumtool")
	viper.SetDefault("sbd_path", "/usr/sbin/sbd")
	viper.SetDefault("sbd_config_path", "/etc/sysconfig/sbd")
	viper.SetDefault("drbdsetup_path", "/usr/sbin/drbdsetup")
}

func main() {
	err := viper.ReadInConfig()
	if err != nil {
		log.Warn(err)
		log.Info("Default config values will be used")
	}

	// parse CLI flags and bind them to the viper config container
	flag.Parse()
	err = viper.BindPFlags(flag.CommandLine)
	if err != nil {
		log.Errorf("Could not bind config to CLI flags: %v", err)
	}

	loglevel(*logLevel)

	pacemakerCollector, err := NewPacemakerCollector(
		viper.GetString("crm_mon_path"),
		viper.GetString("cibadmin_path"),
	)
	if err != nil {
		log.Warnf("Could not initialize Pacemaker collector: %v\n", err)
	} else {
		prometheus.MustRegister(pacemakerCollector)
	}

	corosyncCollector, err := NewCorosyncCollector(
		viper.GetString("corosync_cfgtoolpath_path"),
		viper.GetString("corosync_quorumtool_path"),
	)
	if err != nil {
		log.Warnf("Could not initialize Corosync collector: %v\n", err)
	} else {
		prometheus.MustRegister(corosyncCollector)
	}

	sbdCollector, err := NewSbdCollector(
		viper.GetString("sbd_path"),
		viper.GetString("sbd_config_path"),
	)
	if err != nil {
		log.Warnf("Could not initialize SBD collector: %v\n", err)
	} else {
		prometheus.MustRegister(sbdCollector)
	}

	drbdCollector, err := NewDrbdCollector(viper.GetString("drbdsetup_path"),)
	if err != nil {
		log.Warnf("Could not initialize DRBD collector: %v\n", err)
	} else {
		prometheus.MustRegister(drbdCollector)
	}

	http.HandleFunc("/", landingpage)
	http.Handle("/metrics", promhttp.Handler())
	log.Infof("Serving metrics on %s:%s", *address, *portNumber)
	log.Fatal(http.ListenAndServe(fmt.Sprintf("%s:%s", *address, *portNumber), nil))
}
