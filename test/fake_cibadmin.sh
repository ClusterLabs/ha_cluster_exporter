#!/usr/bin/env bash

cat <<EOF
<cib crm_feature_set="3.1.0" validate-with="pacemaker-3.0" epoch="6881" num_updates="0" admin_epoch="0" cib-last-written="Mon Nov 18 17:48:21 2019" update-origin="node01" update-client="crm_attribute" update-user="root" have-quorum="1" dc-uuid="1084783375">
  <configuration>
    <crm_config>
      <cluster_property_set id="cib-bootstrap-options">
        <nvpair id="cib-bootstrap-options-have-watchdog" name="have-watchdog" value="true"/>
        <nvpair id="cib-bootstrap-options-dc-version" name="dc-version" value="1.1.18+20180430.b12c320f5-3.15.1-b12c320f5"/>
        <nvpair id="cib-bootstrap-options-cluster-infrastructure" name="cluster-infrastructure" value="corosync"/>
        <nvpair id="cib-bootstrap-options-cluster-name" name="cluster-name" value="hana_cluster"/>
        <nvpair name="stonith-enabled" value="true" id="cib-bootstrap-options-stonith-enabled"/>
        <nvpair name="placement-strategy" value="balanced" id="cib-bootstrap-options-placement-strategy"/>
      </cluster_property_set>
    </crm_config>
    <nodes>
      <node id="1084783375" uname="node01">
        <instance_attributes id="nodes-1084783375">
          <nvpair id="nodes-1084783375-lpa_prd_lpt" name="lpa_prd_lpt" value="1574095701"/>
          <nvpair id="nodes-1084783375-hana_prd_vhost" name="hana_prd_vhost" value="node01"/>
          <nvpair id="nodes-1084783375-hana_prd_site" name="hana_prd_site" value="PRIMARY_SITE_NAME"/>
          <nvpair id="nodes-1084783375-hana_prd_op_mode" name="hana_prd_op_mode" value="logreplay"/>
          <nvpair id="nodes-1084783375-hana_prd_srmode" name="hana_prd_srmode" value="sync"/>
          <nvpair id="nodes-1084783375-hana_prd_remoteHost" name="hana_prd_remoteHost" value="node02"/>
        </instance_attributes>
      </node>
      <node id="1084783376" uname="node02">
        <instance_attributes id="nodes-1084783376">
          <nvpair id="nodes-1084783376-lpa_prd_lpt" name="lpa_prd_lpt" value="30"/>
          <nvpair id="nodes-1084783376-hana_prd_op_mode" name="hana_prd_op_mode" value="logreplay"/>
          <nvpair id="nodes-1084783376-hana_prd_vhost" name="hana_prd_vhost" value="node02"/>
          <nvpair id="nodes-1084783376-hana_prd_remoteHost" name="hana_prd_remoteHost" value="node01"/>
          <nvpair id="nodes-1084783376-hana_prd_site" name="hana_prd_site" value="SECONDARY_SITE_NAME"/>
          <nvpair id="nodes-1084783376-hana_prd_srmode" name="hana_prd_srmode" value="sync"/>
        </instance_attributes>
      </node>
    </nodes>
    <resources>
      <primitive id="stonith-sbd" class="stonith" type="external/sbd">
        <instance_attributes id="stonith-sbd-instance_attributes">
          <nvpair name="pcmk_delay_max" value="30s" id="stonith-sbd-instance_attributes-pcmk_delay_max"/>
        </instance_attributes>
      </primitive>
      <primitive id="rsc_ip_PRD_HDB00" class="ocf" provider="heartbeat" type="IPaddr2">
        <!--#-->
        <!--# production HANA-->
        <!--#-->
        <instance_attributes id="rsc_ip_PRD_HDB00-instance_attributes">
          <nvpair name="ip" value="192.168.123.200" id="rsc_ip_PRD_HDB00-instance_attributes-ip"/>
          <nvpair name="cidr_netmask" value="24" id="rsc_ip_PRD_HDB00-instance_attributes-cidr_netmask"/>
          <nvpair name="nic" value="eth1" id="rsc_ip_PRD_HDB00-instance_attributes-nic"/>
        </instance_attributes>
        <operations>
          <op name="start" timeout="20" interval="0" id="rsc_ip_PRD_HDB00-start-0"/>
          <op name="stop" timeout="20" interval="0" id="rsc_ip_PRD_HDB00-stop-0"/>
          <op name="monitor" interval="10" timeout="20" id="rsc_ip_PRD_HDB00-monitor-10"/>
        </operations>
      </primitive>
      <master id="msl_SAPHana_PRD_HDB00">
        <meta_attributes id="msl_SAPHana_PRD_HDB00-meta_attributes">
          <nvpair name="clone-max" value="2" id="msl_SAPHana_PRD_HDB00-meta_attributes-clone-max"/>
          <nvpair name="clone-node-max" value="1" id="msl_SAPHana_PRD_HDB00-meta_attributes-clone-node-max"/>
          <nvpair name="interleave" value="true" id="msl_SAPHana_PRD_HDB00-meta_attributes-interleave"/>
        </meta_attributes>
        <primitive id="rsc_SAPHana_PRD_HDB00" class="ocf" provider="suse" type="SAPHana">
          <instance_attributes id="rsc_SAPHana_PRD_HDB00-instance_attributes">
            <nvpair name="SID" value="PRD" id="rsc_SAPHana_PRD_HDB00-instance_attributes-SID"/>
            <nvpair name="InstanceNumber" value="00" id="rsc_SAPHana_PRD_HDB00-instance_attributes-InstanceNumber"/>
            <nvpair name="PREFER_SITE_TAKEOVER" value="True" id="rsc_SAPHana_PRD_HDB00-instance_attributes-PREFER_SITE_TAKEOVER"/>
            <nvpair name="AUTOMATED_REGISTER" value="False" id="rsc_SAPHana_PRD_HDB00-instance_attributes-AUTOMATED_REGISTER"/>
            <nvpair name="DUPLICATE_PRIMARY_TIMEOUT" value="7200" id="rsc_SAPHana_PRD_HDB00-instance_attributes-DUPLICATE_PRIMARY_TIMEOUT"/>
          </instance_attributes>
          <operations>
            <op name="start" interval="0" timeout="3600" id="rsc_SAPHana_PRD_HDB00-start-0"/>
            <op name="stop" interval="0" timeout="3600" id="rsc_SAPHana_PRD_HDB00-stop-0"/>
            <op name="promote" interval="0" timeout="3600" id="rsc_SAPHana_PRD_HDB00-promote-0"/>
            <op name="monitor" interval="60" role="Master" timeout="700" id="rsc_SAPHana_PRD_HDB00-monitor-60"/>
            <op name="monitor" interval="61" role="Slave" timeout="700" id="rsc_SAPHana_PRD_HDB00-monitor-61"/>
          </operations>
        </primitive>
      </master>
      <clone id="cln_SAPHanaTopology_PRD_HDB00">
        <meta_attributes id="cln_SAPHanaTopology_PRD_HDB00-meta_attributes">
          <nvpair name="is-managed" value="true" id="cln_SAPHanaTopology_PRD_HDB00-meta_attributes-is-managed"/>
          <nvpair name="clone-node-max" value="1" id="cln_SAPHanaTopology_PRD_HDB00-meta_attributes-clone-node-max"/>
          <nvpair name="interleave" value="true" id="cln_SAPHanaTopology_PRD_HDB00-meta_attributes-interleave"/>
        </meta_attributes>
        <primitive id="rsc_SAPHanaTopology_PRD_HDB00" class="ocf" provider="suse" type="SAPHanaTopology">
          <instance_attributes id="rsc_SAPHanaTopology_PRD_HDB00-instance_attributes">
            <nvpair name="SID" value="PRD" id="rsc_SAPHanaTopology_PRD_HDB00-instance_attributes-SID"/>
            <nvpair name="InstanceNumber" value="00" id="rsc_SAPHanaTopology_PRD_HDB00-instance_attributes-InstanceNumber"/>
          </instance_attributes>
          <operations>
            <op name="monitor" interval="10" timeout="600" id="rsc_SAPHanaTopology_PRD_HDB00-monitor-10"/>
            <op name="start" interval="0" timeout="600" id="rsc_SAPHanaTopology_PRD_HDB00-start-0"/>
            <op name="stop" interval="0" timeout="300" id="rsc_SAPHanaTopology_PRD_HDB00-stop-0"/>
          </operations>
        </primitive>
      </clone>
      <primitive id="test" class="ocf" provider="heartbeat" type="Dummy"/>
      <primitive id="test-stop" class="ocf" provider="heartbeat" type="Dummy">
        <meta_attributes id="test-stop-meta_attributes">
          <nvpair id="test-stop-meta_attributes-target-role" name="target-role" value="Stopped"/>
        </meta_attributes>
      </primitive>
    </resources>
    <constraints>
      <rsc_colocation id="col_saphana_ip_PRD_HDB00" score="2000" rsc="rsc_ip_PRD_HDB00" rsc-role="Started" with-rsc="msl_SAPHana_PRD_HDB00" with-rsc-role="Master"/>
      <rsc_order id="ord_SAPHana_PRD_HDB00" kind="Optional" first="cln_SAPHanaTopology_PRD_HDB00" then="msl_SAPHana_PRD_HDB00"/>
      <rsc_location id="cli-prefer-msl_SAPHana_PRD_HDB00" rsc="msl_SAPHana_PRD_HDB00" role="Started" node="node01" score="INFINITY"/>
      <rsc_location id="cli-prefer-cln_SAPHanaTopology_PRD_HDB00" rsc="cln_SAPHanaTopology_PRD_HDB00" role="Started" node="node01" score="INFINITY"/>
      <rsc_location id="cli-ban-msl_SAPHana_PRD_HDB00-on-node01" rsc="msl_SAPHana_PRD_HDB00" role="Started" node="node01" score="-INFINITY"/>
      <rsc_location id="test" rsc="test" role="Started" node="node02" score="666"/>
    </constraints>
    <rsc_defaults>
      <meta_attributes id="rsc-options">
        <nvpair name="resource-stickiness" value="1000" id="rsc-options-resource-stickiness"/>
        <nvpair name="migration-threshold" value="5000" id="rsc-options-migration-threshold"/>
      </meta_attributes>
    </rsc_defaults>
    <op_defaults>
      <meta_attributes id="op-options">
        <nvpair name="timeout" value="600" id="op-options-timeout"/>
        <nvpair name="record-pending" value="true" id="op-options-record-pending"/>
      </meta_attributes>
    </op_defaults>
  </configuration>
  <status>
    <node_state id="1084783375" uname="node01" in_ccm="true" crmd="online" crm-debug-origin="do_update_resource" join="member" expected="member">
      <transient_attributes id="1084783375">
        <instance_attributes id="status-1084783375">
          <nvpair id="status-1084783375-master-rsc_SAPHana_PRD_HDB00" name="master-rsc_SAPHana_PRD_HDB00" value="150"/>
          <nvpair id="status-1084783375-hana_prd_version" name="hana_prd_version" value="2.00.040.00.1553674765"/>
          <nvpair id="status-1084783375-hana_prd_clone_state" name="hana_prd_clone_state" value="PROMOTED"/>
          <nvpair id="status-1084783375-hana_prd_sync_state" name="hana_prd_sync_state" value="PRIM"/>
          <nvpair id="status-1084783375-hana_prd_roles" name="hana_prd_roles" value="4:P:master1:master:worker:master"/>
        </instance_attributes>
      </transient_attributes>
      <lrm id="1084783375">
        <lrm_resources>
          <lrm_resource id="rsc_SAPHana_PRD_HDB00" type="SAPHana" class="ocf" provider="suse">
            <lrm_rsc_op id="rsc_SAPHana_PRD_HDB00_last_failure_0" operation_key="rsc_SAPHana_PRD_HDB00_monitor_0" operation="monitor" crm-debug-origin="build_active_RAs" crm_feature_set="3.1.0" transition-key="3:3:7:70ea6528-73ad-48be-9eb7-583ee933f216" transition-magic="0:0;3:3:7:70ea6528-73ad-48be-9eb7-583ee933f216" exit-reason="" on_node="node01" call-id="15" rc-code="0" op-status="0" interval="0" last-run="1573663876" last-rc-change="1573663876" exec-time="3450" queue-time="0" op-digest="ff4ff123bc6f906497ef0ef5e44dffd1"/>
            <lrm_rsc_op id="rsc_SAPHana_PRD_HDB00_last_0" operation_key="rsc_SAPHana_PRD_HDB00_promote_0" operation="promote" crm-debug-origin="do_update_resource" crm_feature_set="3.1.0" transition-key="12:8:0:70ea6528-73ad-48be-9eb7-583ee933f216" transition-magic="0:0;12:8:0:70ea6528-73ad-48be-9eb7-583ee933f216" exit-reason="" on_node="node01" call-id="31" rc-code="0" op-status="0" interval="0" last-run="1573663898" last-rc-change="1573663898" exec-time="2257" queue-time="0" op-digest="ff4ff123bc6f906497ef0ef5e44dffd1" op-force-restart=" INSTANCE_PROFILE " op-restart-digest="f2317cad3d54cec5d7d7aa7d0bf35cf8"/>
            <lrm_rsc_op id="rsc_SAPHana_PRD_HDB00_monitor_60000" operation_key="rsc_SAPHana_PRD_HDB00_monitor_60000" operation="monitor" crm-debug-origin="do_update_resource" crm_feature_set="3.1.0" transition-key="14:9:8:70ea6528-73ad-48be-9eb7-583ee933f216" transition-magic="0:8;14:9:8:70ea6528-73ad-48be-9eb7-583ee933f216" exit-reason="" on_node="node01" call-id="32" rc-code="8" op-status="0" interval="60000" last-rc-change="1573663906" exec-time="3586" queue-time="0" op-digest="05b857e482ebd46019d347fd55ebbcdb"/>
          </lrm_resource>
          <lrm_resource id="rsc_ip_PRD_HDB00" type="IPaddr2" class="ocf" provider="heartbeat">
            <lrm_rsc_op id="rsc_ip_PRD_HDB00_last_0" operation_key="rsc_ip_PRD_HDB00_start_0" operation="start" crm-debug-origin="build_active_RAs" crm_feature_set="3.1.0" transition-key="7:3:0:70ea6528-73ad-48be-9eb7-583ee933f216" transition-magic="0:0;7:3:0:70ea6528-73ad-48be-9eb7-583ee933f216" exit-reason="" on_node="node01" call-id="21" rc-code="0" op-status="0" interval="0" last-run="1573663876" last-rc-change="1573663876" exec-time="136" queue-time="0" op-digest="a6da6959be1e15c2f9f5e88476e82ba4"/>
            <lrm_rsc_op id="rsc_ip_PRD_HDB00_monitor_10000" operation_key="rsc_ip_PRD_HDB00_monitor_10000" operation="monitor" crm-debug-origin="build_active_RAs" crm_feature_set="3.1.0" transition-key="8:3:0:70ea6528-73ad-48be-9eb7-583ee933f216" transition-magic="0:0;8:3:0:70ea6528-73ad-48be-9eb7-583ee933f216" exit-reason="" on_node="node01" call-id="22" rc-code="0" op-status="0" interval="10000" last-rc-change="1573663876" exec-time="85" queue-time="0" op-digest="c7df6e2194c50ed86aa98b66e909fe11"/>
          </lrm_resource>
          <lrm_resource id="stonith-sbd" type="external/sbd" class="stonith">
            <lrm_rsc_op id="stonith-sbd_last_0" operation_key="stonith-sbd_start_0" operation="start" crm-debug-origin="build_active_RAs" crm_feature_set="3.1.0" transition-key="3:2:0:70ea6528-73ad-48be-9eb7-583ee933f216" transition-magic="0:0;3:2:0:70ea6528-73ad-48be-9eb7-583ee933f216" exit-reason="" on_node="node01" call-id="6" rc-code="0" op-status="0" interval="0" last-run="1573663874" last-rc-change="1573663874" exec-time="2238" queue-time="0" op-digest="265be3215da5e5037d35e7fe1bcc5ae0"/>
          </lrm_resource>
          <lrm_resource id="rsc_SAPHanaTopology_PRD_HDB00" type="SAPHanaTopology" class="ocf" provider="suse">
            <lrm_rsc_op id="rsc_SAPHanaTopology_PRD_HDB00_last_0" operation_key="rsc_SAPHanaTopology_PRD_HDB00_start_0" operation="start" crm-debug-origin="build_active_RAs" crm_feature_set="3.1.0" transition-key="19:4:0:70ea6528-73ad-48be-9eb7-583ee933f216" transition-magic="0:0;19:4:0:70ea6528-73ad-48be-9eb7-583ee933f216" exit-reason="" on_node="node01" call-id="24" rc-code="0" op-status="0" interval="0" last-run="1573663881" last-rc-change="1573663881" exec-time="4355" queue-time="0" op-digest="2d8d79c3726afb91c33d406d5af79b53" op-force-restart="" op-restart-digest="f2317cad3d54cec5d7d7aa7d0bf35cf8"/>
            <lrm_rsc_op id="rsc_SAPHanaTopology_PRD_HDB00_monitor_10000" operation_key="rsc_SAPHanaTopology_PRD_HDB00_monitor_10000" operation="monitor" crm-debug-origin="build_active_RAs" crm_feature_set="3.1.0" transition-key="22:5:0:70ea6528-73ad-48be-9eb7-583ee933f216" transition-magic="0:0;22:5:0:70ea6528-73ad-48be-9eb7-583ee933f216" exit-reason="" on_node="node01" call-id="26" rc-code="0" op-status="0" interval="10000" last-rc-change="1573663885" exec-time="4949" queue-time="0" op-digest="64db68ca3e12e0d41eb98ce63b9610d2"/>
          </lrm_resource>
          <lrm_resource id="test" type="Dummy" class="ocf" provider="heartbeat">
            <lrm_rsc_op id="test_last_0" operation_key="test_start_0" operation="start" crm-debug-origin="do_update_resource" crm_feature_set="3.1.0" transition-key="8:6863:0:70ea6528-73ad-48be-9eb7-583ee933f216" transition-magic="0:0;8:6863:0:70ea6528-73ad-48be-9eb7-583ee933f216" exit-reason="" on_node="node01" call-id="37" rc-code="0" op-status="0" interval="0" last-run="1574095329" last-rc-change="1574095329" exec-time="10" queue-time="0" op-digest="f2317cad3d54cec5d7d7aa7d0bf35cf8" op-force-restart=" state " op-restart-digest="f2317cad3d54cec5d7d7aa7d0bf35cf8"/>
          </lrm_resource>
          <lrm_resource id="test-stop" type="Dummy" class="ocf" provider="heartbeat">
            <lrm_rsc_op id="test-stop_last_0" operation_key="test-stop_monitor_0" operation="monitor" crm-debug-origin="do_update_resource" crm_feature_set="3.1.0" transition-key="7:13662:7:5a2e7427-7cbd-4bd9-8e8c-fd633866c4a9" transition-magic="0:7;7:13662:7:5a2e7427-7cbd-4bd9-8e8c-fd633866c4a9" exit-reason="" on_node="stefanotorresi2-node01" call-id="40" rc-code="7" op-status="0" interval="0" last-run="1582534010" last-rc-change="1582534010" exec-time="9" queue-time="0" op-digest="f2317cad3d54cec5d7d7aa7d0bf35cf8" op-force-restart=" state " op-restart-digest="f2317cad3d54cec5d7d7aa7d0bf35cf8"/>
          </lrm_resource>
        </lrm_resources>
      </lrm>
    </node_state>
    <node_state id="1084783376" in_ccm="true" crmd="online" crm-debug-origin="do_update_resource" uname="node02" join="member" expected="member">
      <lrm id="1084783376">
        <lrm_resources>
          <lrm_resource id="stonith-sbd" type="external/sbd" class="stonith">
            <lrm_rsc_op id="stonith-sbd_last_0" operation_key="stonith-sbd_monitor_0" operation="monitor" crm-debug-origin="do_update_resource" crm_feature_set="3.1.0" transition-key="5:6:7:70ea6528-73ad-48be-9eb7-583ee933f216" transition-magic="0:7;5:6:7:70ea6528-73ad-48be-9eb7-583ee933f216" exit-reason="" on_node="node02" call-id="5" rc-code="7" op-status="0" interval="0" last-run="1573663890" last-rc-change="1573663890" exec-time="1" queue-time="0" op-digest="265be3215da5e5037d35e7fe1bcc5ae0"/>
          </lrm_resource>
          <lrm_resource id="rsc_ip_PRD_HDB00" type="IPaddr2" class="ocf" provider="heartbeat">
            <lrm_rsc_op id="rsc_ip_PRD_HDB00_last_0" operation_key="rsc_ip_PRD_HDB00_monitor_0" operation="monitor" crm-debug-origin="do_update_resource" crm_feature_set="3.1.0" transition-key="6:6:7:70ea6528-73ad-48be-9eb7-583ee933f216" transition-magic="0:7;6:6:7:70ea6528-73ad-48be-9eb7-583ee933f216" exit-reason="" on_node="node02" call-id="9" rc-code="7" op-status="0" interval="0" last-run="1573663890" last-rc-change="1573663890" exec-time="56" queue-time="0" op-digest="a6da6959be1e15c2f9f5e88476e82ba4"/>
          </lrm_resource>
          <lrm_resource id="rsc_SAPHana_PRD_HDB00" type="SAPHana" class="ocf" provider="suse">
            <lrm_rsc_op id="rsc_SAPHana_PRD_HDB00_last_0" operation_key="rsc_SAPHana_PRD_HDB00_monitor_0" operation="monitor" crm-debug-origin="do_update_resource" crm_feature_set="3.1.0" transition-key="7:6:7:70ea6528-73ad-48be-9eb7-583ee933f216" transition-magic="0:0;7:6:7:70ea6528-73ad-48be-9eb7-583ee933f216" exit-reason="" on_node="node02" call-id="14" rc-code="0" op-status="0" interval="0" last-run="1573663890" last-rc-change="1573663890" exec-time="3515" queue-time="0" op-digest="ff4ff123bc6f906497ef0ef5e44dffd1" op-force-restart=" INSTANCE_PROFILE " op-restart-digest="f2317cad3d54cec5d7d7aa7d0bf35cf8"/>
            <lrm_rsc_op id="rsc_SAPHana_PRD_HDB00_last_failure_0" operation_key="rsc_SAPHana_PRD_HDB00_monitor_0" operation="monitor" crm-debug-origin="do_update_resource" crm_feature_set="3.1.0" transition-key="7:6:7:70ea6528-73ad-48be-9eb7-583ee933f216" transition-magic="0:0;7:6:7:70ea6528-73ad-48be-9eb7-583ee933f216" exit-reason="" on_node="node02" call-id="14" rc-code="0" op-status="0" interval="0" last-run="1573663890" last-rc-change="1573663890" exec-time="3515" queue-time="0" op-digest="ff4ff123bc6f906497ef0ef5e44dffd1"/>
            <lrm_rsc_op id="rsc_SAPHana_PRD_HDB00_monitor_61000" operation_key="rsc_SAPHana_PRD_HDB00_monitor_61000" operation="monitor" crm-debug-origin="do_update_resource" crm_feature_set="3.1.0" transition-key="13:7:0:70ea6528-73ad-48be-9eb7-583ee933f216" transition-magic="0:0;13:7:0:70ea6528-73ad-48be-9eb7-583ee933f216" exit-reason="" on_node="node02" call-id="20" rc-code="0" op-status="0" interval="61000" last-rc-change="1573663895" exec-time="3225" queue-time="0" op-digest="05b857e482ebd46019d347fd55ebbcdb"/>
          </lrm_resource>
          <lrm_resource id="rsc_SAPHanaTopology_PRD_HDB00" type="SAPHanaTopology" class="ocf" provider="suse">
            <lrm_rsc_op id="rsc_SAPHanaTopology_PRD_HDB00_last_0" operation_key="rsc_SAPHanaTopology_PRD_HDB00_start_0" operation="start" crm-debug-origin="do_update_resource" crm_feature_set="3.1.0" transition-key="24:7:0:70ea6528-73ad-48be-9eb7-583ee933f216" transition-magic="0:0;24:7:0:70ea6528-73ad-48be-9eb7-583ee933f216" exit-reason="" on_node="node02" call-id="21" rc-code="0" op-status="0" interval="0" last-run="1573663895" last-rc-change="1573663895" exec-time="3650" queue-time="0" op-digest="2d8d79c3726afb91c33d406d5af79b53" op-force-restart="" op-restart-digest="f2317cad3d54cec5d7d7aa7d0bf35cf8"/>
            <lrm_rsc_op id="rsc_SAPHanaTopology_PRD_HDB00_monitor_10000" operation_key="rsc_SAPHanaTopology_PRD_HDB00_monitor_10000" operation="monitor" crm-debug-origin="do_update_resource" crm_feature_set="3.1.0" transition-key="28:8:0:70ea6528-73ad-48be-9eb7-583ee933f216" transition-magic="0:0;28:8:0:70ea6528-73ad-48be-9eb7-583ee933f216" exit-reason="" on_node="node02" call-id="22" rc-code="0" op-status="0" interval="10000" last-rc-change="1573663898" exec-time="3978" queue-time="0" op-digest="64db68ca3e12e0d41eb98ce63b9610d2"/>
          </lrm_resource>
          <lrm_resource id="test" type="Dummy" class="ocf" provider="heartbeat">
            <lrm_rsc_op id="test_last_0" operation_key="test_stop_0" operation="stop" crm-debug-origin="do_update_resource" crm_feature_set="3.1.0" transition-key="7:6863:0:70ea6528-73ad-48be-9eb7-583ee933f216" transition-magic="0:0;7:6863:0:70ea6528-73ad-48be-9eb7-583ee933f216" exit-reason="" on_node="node02" call-id="28" rc-code="0" op-status="0" interval="0" last-run="1574095329" last-rc-change="1574095329" exec-time="12" queue-time="0" op-digest="f2317cad3d54cec5d7d7aa7d0bf35cf8" op-force-restart=" state " op-restart-digest="f2317cad3d54cec5d7d7aa7d0bf35cf8"/>
          </lrm_resource>
          <lrm_resource id="test-stop" type="Dummy" class="ocf" provider="heartbeat">
            <lrm_rsc_op id="test-stop_last_0" operation_key="test-stop_stop_0" operation="stop" crm-debug-origin="do_update_resource" crm_feature_set="3.1.0" transition-key="35:13663:0:5a2e7427-7cbd-4bd9-8e8c-fd633866c4a9" transition-magic="0:0;35:13663:0:5a2e7427-7cbd-4bd9-8e8c-fd633866c4a9" exit-reason="" on_node="stefanotorresi2-node02" call-id="35" rc-code="0" op-status="0" interval="0" last-run="1582534018" last-rc-change="1582534018" exec-time="12" queue-time="0" op-digest="f2317cad3d54cec5d7d7aa7d0bf35cf8" op-force-restart=" state " op-restart-digest="f2317cad3d54cec5d7d7aa7d0bf35cf8"/>
          </lrm_resource>
        </lrm_resources>
      </lrm>
      <transient_attributes id="1084783376">
        <instance_attributes id="status-1084783376">
          <nvpair id="status-1084783376-hana_prd_clone_state" name="hana_prd_clone_state" value="DEMOTED"/>
          <nvpair id="status-1084783376-master-rsc_SAPHana_PRD_HDB00" name="master-rsc_SAPHana_PRD_HDB00" value="100"/>
          <nvpair id="status-1084783376-hana_prd_version" name="hana_prd_version" value="2.00.040.00.1553674765"/>
          <nvpair id="status-1084783376-hana_prd_roles" name="hana_prd_roles" value="4:S:master1:master:worker:master"/>
          <nvpair id="status-1084783376-hana_prd_sync_state" name="hana_prd_sync_state" value="SOK"/>
        </instance_attributes>
      </transient_attributes>
    </node_state>
  </status>
</cib>
EOF
