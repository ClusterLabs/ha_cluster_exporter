package sbd

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"log/slog"
	"os"
	"time"

	assertcustom "github.com/ClusterLabs/ha_cluster_exporter/internal/assert"
)

func TestReadSbdConfFileError(t *testing.T) {
	sbdConfFile, err := readSdbFile("../../../test/nonexistent")

	assert.Nil(t, sbdConfFile)
	assert.Error(t, err)
}

func TestGetSbdDevicesWithoutDoubleQuotes(t *testing.T) {
	// this is a full config file more or less , in other tests it is cutted
	sbdConfig := `
	 # SBD_DEVICE specifies the devices to use for exchanging sbd messages
	 # and to monitor. If specifying more than one path, use ";" as
	 # separator.
	 #
	 #SBD_DEVICE=""
	 
	 ## Type: yesno
	 ## Default: yes
	 #
	 # Whether to enable the pacemaker integration.
	 #
	 SBD_PACEMAKER=yes
	 
	 ## Type: list(always,clean)
	 ## Default: always
	 #
	 # Specify the start mode for sbd. Setting this to "clean" will only
	 # allow sbd to start if it was not previously fenced. See the -S option
	 # in the man page.
	 #
	 SBD_STARTMODE=always
	 
	 ## Type: yesno / integer
	 ## Default: no
	 #
	 # Whether to delay after starting sbd on boot for "msgwait" seconds.
	 # This may be necessary if your cluster nodes reboot so fast that the
	 # other nodes are still waiting in the fence acknowledgement phase.
	 # This is an occasional issue with virtual machines.
	 #
	 # This can also be enabled by being set to a specific delay value, in
	 # seconds. Sometimes a longer delay than the default, "msgwait", is
	 # needed, for example in the cases where it's considered to be safer to
	 # wait longer than:
	 # corosync token timeout + consensus timeout + pcmk_delay_max + msgwait
	 #
	 # Be aware that the special value "1" means "yes" rather than "1s".
	 #
	 # Consider that you might have to adapt the startup-timeout accordingly
	 # if the default isn't sufficient. (TimeoutStartSec for systemd)
	 #
	 # This option may be ignored at a later point, once pacemaker handles
	 # this case better.
	 #
	 SBD_DELAY_START=no
	 
	 ## Type: string
	 ## Default: /dev/watchdog
	 #
	 # Watchdog device to use. If set to /dev/null, no watchdog device will
	 # be used.
	 #
	 SBD_WATCHDOG_DEV=/dev/watchdog
	 
	 ## Type: integer
	 ## Default: 5
	 #
	 # How long, in seconds, the watchdog will wait before panicking the
	 # node if no-one tickles it.
	 #
	 # This depends mostly on your storage latency; the majority of devices
	 # must be successfully read within this time, or else the node will
	 # self-fence.
	 #
	 # If your sbd device(s) reside on a multipath setup or iSCSI, this
	 # should be the time required to detect a path failure.
	 #
	 # Be aware that watchdog timeout set in the on-disk metadata takes
	 # precedence.
	 #
	 SBD_WATCHDOG_TIMEOUT=5
	 
	 ## Type: string
	 ## Default: "flush,reboot"
	 #
	 # Actions to be executed when the watchers don't timely report to the sbd
	 # master process or one of the watchers detects that the master process
	 # has died.
	 #
	 # Set timeout-action to comma-separated combination of
	 # noflush|flush plus reboot|crashdump|off.
	 # If just one of both is given the other stays at the default.
	 #
	 # This doesn't affect actions like off, crashdump, reboot explicitly
	 # triggered via message slots.
	 # And it does as well not configure the action a watchdog would
	 # trigger should it run off (there is no generic interface).
	 #
	 SBD_TIMEOUT_ACTION=flush,reboot
	 
	 ## Type: string
	 ## Default: ""
	 #
	 # Additional options for starting sbd
	 #
	 SBD_OPTS=
	 SBD_DEVICE=/dev/vda;/dev/vdb;/dev/vdc
`

	sbdDevices := getSbdDevices([]byte(sbdConfig))

	assert.Len(t, sbdDevices, 3)
	assert.Equal(t, "/dev/vda", sbdDevices[0])
	assert.Equal(t, "/dev/vdb", sbdDevices[1])
	assert.Equal(t, "/dev/vdc", sbdDevices[2])
}

// test the other case with double quotes, and put the string in random place
func TestGetSbdDevicesWithDoubleQuotes(t *testing.T) {
	sbdConfig := `## Type: string
	 ## Default: ""
	 #
	 # SBD_DEVICE specifies the devices to use for exchanging sbd messages
	 # and to monitor. If specifying more than one path, use ";" as
	 # separator.
	 #
	 #SBD_DEVICE=""

	 SBD_WATCHDOG_TIMEOUT=5
	 
	 SBD_DEVICE="/dev/vda;/dev/vdb;/dev/vdc"

	 SBD_TIMEOUT_ACTION=flush,reboot
	 
	 ## Type: string
	 ## Default: ""
	 #
	 # Additional options for starting sbd
	 #
	 SBD_OPTS=`

	sbdDevices := getSbdDevices([]byte(sbdConfig))

	assert.Len(t, sbdDevices, 3)
	assert.Equal(t, "/dev/vda", sbdDevices[0])
	assert.Equal(t, "/dev/vdb", sbdDevices[1])
	assert.Equal(t, "/dev/vdc", sbdDevices[2])
}

// test the other case with double quotes, and put the string in random place
func TestOnlyOneDeviceSbd(t *testing.T) {
	sbdConfig := `## Type: string
	 ## Default: ""
	
	 SBD_DEVICE=/dev/vdc

	 ## Type: string
	 ## Default: "flush,reboot"
`

	sbdDevices := getSbdDevices([]byte(sbdConfig))

	assert.Len(t, sbdDevices, 1)
	assert.Equal(t, "/dev/vdc", sbdDevices[0])
}

func TestSbdDeviceParserWithFullCommentBeforeActualSetting(t *testing.T) {
	sbdConfig := `
# SBD_DEVICE=/dev/foo
SBD_DEVICE=/dev/vdc;/dev/vdd`

	sbdDevices := getSbdDevices([]byte(sbdConfig))

	assert.Len(t, sbdDevices, 2)
	assert.Equal(t, "/dev/vdc", sbdDevices[0])
	assert.Equal(t, "/dev/vdd", sbdDevices[1])
}

func TestSbdDeviceParserWithSpaceAfterSemicolon(t *testing.T) {
	sbdConfig := `SBD_DEVICE=/dev/vdc; /dev/vdd`

	sbdDevices := getSbdDevices([]byte(sbdConfig))

	assert.Len(t, sbdDevices, 2)
	assert.Equal(t, "/dev/vdc", sbdDevices[0])
	assert.Equal(t, "/dev/vdd", sbdDevices[1])
}

func TestSbdDeviceParserWithSemicolon(t *testing.T) {
	sbdConfig := `SBD_DEVICE=/dev/vdc;/dev/vdd;`

	sbdDevices := getSbdDevices([]byte(sbdConfig))

	assert.Len(t, sbdDevices, 2)
	assert.Equal(t, "/dev/vdc", sbdDevices[0])
	assert.Equal(t, "/dev/vdd", sbdDevices[1])
}

func TestNewSbdCollector(t *testing.T) {
	_, err := NewCollector("../../../test/fake_sbd.sh", "../../../test/fake_sbdconfig", 10*time.Second, slog.New(slog.NewTextHandler(os.Stdout, nil)))

	assert.Nil(t, err)
}

func TestNewSbdCollectorChecksSbdConfigExistence(t *testing.T) {
	_, err := NewCollector("../../../test/fake_sbd.sh", "../../../test/nonexistent", 10*time.Second, slog.New(slog.NewTextHandler(os.Stdout, nil)))

	assert.NoError(t, err)
}

func TestNewSbdCollectorChecksSbdExistence(t *testing.T) {
	_, err := NewCollector("../../../test/nonexistent", "../../../test/fake_sbdconfig", 10*time.Second, slog.New(slog.NewTextHandler(os.Stdout, nil)))

	assert.NoError(t, err)
}

func TestNewSbdCollectorChecksSbdExecutableBits(t *testing.T) {
	_, err := NewCollector("../../../test/dummy", "../../../test/fake_sbdconfig", 10*time.Second, slog.New(slog.NewTextHandler(os.Stdout, nil)))

	assert.NoError(t, err)
}

func TestSBDCollector(t *testing.T) {
	collector, _ := NewCollector("../../../test/fake_sbd_dump.sh", "../../../test/fake_sbdconfig", 10*time.Second, slog.New(slog.NewTextHandler(os.Stdout, nil)))
	assertcustom.Metrics(t, collector, "../../../test/sbd.metrics")
}

func TestWatchdog(t *testing.T) {
	collector, err := NewCollector("../../../test/fake_sbd_dump.sh", "../../../test/fake_sbdconfig", 10*time.Second, slog.New(slog.NewTextHandler(os.Stdout, nil)))

	assert.Nil(t, err)
	assertcustom.Metrics(t, collector, "../../../test/sbd.metrics")
}
