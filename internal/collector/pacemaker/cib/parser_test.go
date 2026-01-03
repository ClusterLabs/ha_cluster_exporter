package cib

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestConstructor(t *testing.T) {
	p := NewCibAdminParser("foo", 10*time.Second)
	assert.Equal(t, "foo", p.cibAdminPath)
}

func TestParse(t *testing.T) {
	p := NewCibAdminParser("../../../../test/fake_cibadmin.sh", 10*time.Second)
	data, err := p.Parse()
	assert.NoError(t, err)
	assert.Equal(t, 2, len(data.Configuration.Nodes))
	assert.Equal(t, "cib-bootstrap-options-cluster-name", data.Configuration.CrmConfig.ClusterProperties[3].Id)
	assert.Equal(t, "hana_cluster", data.Configuration.CrmConfig.ClusterProperties[3].Value)
	assert.Equal(t, "node01", data.Configuration.Nodes[0].Uname)
	assert.Equal(t, "node02", data.Configuration.Nodes[1].Uname)
	assert.Equal(t, 4, len(data.Configuration.Resources.Primitives))
	assert.Equal(t, 1, len(data.Configuration.Resources.Masters))
	assert.Equal(t, 1, len(data.Configuration.Resources.Clones))
	assert.Equal(t, "stonith-sbd", data.Configuration.Resources.Primitives[0].Id)
	assert.Equal(t, "stonith", data.Configuration.Resources.Primitives[0].Class)
	assert.Equal(t, "external/sbd", data.Configuration.Resources.Primitives[0].Type)
	assert.Equal(t, 1, len(data.Configuration.Resources.Primitives[0].InstanceAttributes))
	assert.Equal(t, "pcmk_delay_max", data.Configuration.Resources.Primitives[0].InstanceAttributes[0].Name)
	assert.Equal(t, "stonith-sbd-instance_attributes-pcmk_delay_max", data.Configuration.Resources.Primitives[0].InstanceAttributes[0].Id)
	assert.Equal(t, "30s", data.Configuration.Resources.Primitives[0].InstanceAttributes[0].Value)
	assert.Equal(t, "msl_SAPHana_PRD_HDB00", data.Configuration.Resources.Masters[0].Id)
	assert.Equal(t, 3, len(data.Configuration.Resources.Masters[0].MetaAttributes))
	assert.Equal(t, "rsc_SAPHana_PRD_HDB00", data.Configuration.Resources.Masters[0].Primitive.Id)
	assert.Equal(t, 5, len(data.Configuration.Resources.Masters[0].Primitive.Operations))
	assert.Equal(t, "rsc_SAPHana_PRD_HDB00-start-0", data.Configuration.Resources.Masters[0].Primitive.Operations[0].Id)
	assert.Equal(t, "start", data.Configuration.Resources.Masters[0].Primitive.Operations[0].Name)
	assert.Equal(t, "0", data.Configuration.Resources.Masters[0].Primitive.Operations[0].Interval)
	assert.Equal(t, "3600", data.Configuration.Resources.Masters[0].Primitive.Operations[0].Timeout)
	assert.Equal(t, "rsc_SAPHana_PRD_HDB00-stop-0", data.Configuration.Resources.Masters[0].Primitive.Operations[1].Id)
	assert.Equal(t, "stop", data.Configuration.Resources.Masters[0].Primitive.Operations[1].Name)
	assert.Equal(t, "0", data.Configuration.Resources.Masters[0].Primitive.Operations[1].Interval)
	assert.Equal(t, "3600", data.Configuration.Resources.Masters[0].Primitive.Operations[1].Timeout)
	assert.Equal(t, "rsc_SAPHana_PRD_HDB00-promote-0", data.Configuration.Resources.Masters[0].Primitive.Operations[2].Id)
	assert.Equal(t, "promote", data.Configuration.Resources.Masters[0].Primitive.Operations[2].Name)
	assert.Equal(t, "0", data.Configuration.Resources.Masters[0].Primitive.Operations[2].Interval)
	assert.Equal(t, "3600", data.Configuration.Resources.Masters[0].Primitive.Operations[2].Timeout)
	assert.Equal(t, "rsc_SAPHana_PRD_HDB00-monitor-60", data.Configuration.Resources.Masters[0].Primitive.Operations[3].Id)
	assert.Equal(t, "monitor", data.Configuration.Resources.Masters[0].Primitive.Operations[3].Name)
	assert.Equal(t, "Master", data.Configuration.Resources.Masters[0].Primitive.Operations[3].Role)
	assert.Equal(t, "60", data.Configuration.Resources.Masters[0].Primitive.Operations[3].Interval)
	assert.Equal(t, "700", data.Configuration.Resources.Masters[0].Primitive.Operations[3].Timeout)
	assert.Equal(t, "rsc_SAPHana_PRD_HDB00-monitor-61", data.Configuration.Resources.Masters[0].Primitive.Operations[4].Id)
	assert.Equal(t, "monitor", data.Configuration.Resources.Masters[0].Primitive.Operations[4].Name)
	assert.Equal(t, "Slave", data.Configuration.Resources.Masters[0].Primitive.Operations[4].Role)
	assert.Equal(t, "61", data.Configuration.Resources.Masters[0].Primitive.Operations[4].Interval)
	assert.Equal(t, "700", data.Configuration.Resources.Masters[0].Primitive.Operations[4].Timeout)
	assert.Equal(t, "test", data.Configuration.Resources.Primitives[2].Id)
	assert.Equal(t, "ocf", data.Configuration.Resources.Primitives[2].Class)
	assert.Equal(t, "heartbeat", data.Configuration.Resources.Primitives[2].Provider)
	assert.Equal(t, "Dummy", data.Configuration.Resources.Primitives[2].Type)

}
