#!/usr/bin/env bash

cat <<EOF
<?xml version="1.0"?>
<crm_mon version="2.0.0">
    <summary>
        <stack type="corosync" />
        <current_dc present="true" version="1.1.18+20180430.b12c320f5-3.15.1-b12c320f5" name="hana01" id="1084783375" with_quorum="true" />
        <last_update time="Fri Oct 18 11:48:54 2019" />
        <last_change time="Fri Oct 18 11:48:22 2019" user="root" client="crm_attribute" origin="hana01" />
        <nodes_configured number="2" />
        <resources_configured number="6" disabled="0" blocked="0" />
        <cluster_options stonith-enabled="true" symmetric-cluster="true" no-quorum-policy="stop" maintenance-mode="false" />
    </summary>
    <nodes>
        <node name="hana01" id="1084783375" online="true" standby="false" standby_onfail="false" maintenance="false" pending="false" unclean="false" shutdown="false" expected_up="true" is_dc="true" resources_running="4" type="member" >
            <resource id="rsc_SAPHana_PRD_HDB00" resource_agent="ocf::suse:SAPHana" role="Master" active="true" orphaned="false" blocked="false" managed="true" failed="false" failure_ignored="false" nodes_running_on="1" />
            <resource id="rsc_ip_PRD_HDB00" resource_agent="ocf::heartbeat:IPaddr2" role="Started" active="true" orphaned="false" blocked="false" managed="true" failed="false" failure_ignored="false" nodes_running_on="1" />
            <resource id="stonith-sbd" resource_agent="stonith:external/sbd" role="Started" active="true" orphaned="false" blocked="false" managed="true" failed="false" failure_ignored="false" nodes_running_on="1" />
            <resource id="rsc_SAPHanaTopology_PRD_HDB00" resource_agent="ocf::suse:SAPHanaTopology" role="Started" active="true" orphaned="false" blocked="false" managed="true" failed="false" failure_ignored="false" nodes_running_on="1" />
        </node>
        <node name="hana02" id="1084783376" online="true" standby="false" standby_onfail="false" maintenance="false" pending="false" unclean="false" shutdown="false" expected_up="true" is_dc="false" resources_running="2" type="member" >
            <resource id="rsc_SAPHana_PRD_HDB00" resource_agent="ocf::suse:SAPHana" role="Slave" active="true" orphaned="false" blocked="false" managed="true" failed="false" failure_ignored="false" nodes_running_on="1" />
            <resource id="rsc_SAPHanaTopology_PRD_HDB00" resource_agent="ocf::suse:SAPHanaTopology" role="Started" active="true" orphaned="false" blocked="false" managed="true" failed="false" failure_ignored="false" nodes_running_on="1" />
        </node>
    </nodes>
    <node_attributes>
        <node name="hana01">
            <attribute name="hana_prd_clone_state" value="PROMOTED" />
            <attribute name="hana_prd_op_mode" value="logreplay" />
            <attribute name="hana_prd_remoteHost" value="hana02" />
            <attribute name="hana_prd_roles" value="4:P:master1:master:worker:master" />
            <attribute name="hana_prd_site" value="PRIMARY_SITE_NAME" />
            <attribute name="hana_prd_srmode" value="sync" />
            <attribute name="hana_prd_sync_state" value="PRIM" />
            <attribute name="hana_prd_version" value="2.00.040.00.1553674765" />
            <attribute name="hana_prd_vhost" value="hana01" />
            <attribute name="lpa_prd_lpt" value="1571392102" />
            <attribute name="master-rsc_SAPHana_PRD_HDB00" value="150" />
        </node>
        <node name="hana02">
            <attribute name="hana_prd_clone_state" value="DEMOTED" />
            <attribute name="hana_prd_op_mode" value="logreplay" />
            <attribute name="hana_prd_remoteHost" value="hana01" />
            <attribute name="hana_prd_roles" value="4:S:master1:master:worker:master" />
            <attribute name="hana_prd_site" value="SECONDARY_SITE_NAME" />
            <attribute name="hana_prd_srmode" value="sync" />
            <attribute name="hana_prd_sync_state" value="SOK" />
            <attribute name="hana_prd_version" value="2.00.040.00.1553674765" />
            <attribute name="hana_prd_vhost" value="hana02" />
            <attribute name="lpa_prd_lpt" value="30" />
            <attribute name="master-rsc_SAPHana_PRD_HDB00" value="100" />
        </node>
    </node_attributes>
    <node_history>
        <node name="hana01">
            <resource_history id="rsc_SAPHana_PRD_HDB00" orphan="false" migration-threshold="5000" fail-count="1000000" last-failure="Wed Oct 23 12:37:22 2019">
                <operation_history call="15" task="probe" last-rc-change="Thu Oct 10 12:57:33 2019" last-run="Thu Oct 10 12:57:33 2019" exec-time="4140ms" queue-time="0ms" rc="0" rc_text="ok" />
                <operation_history call="31" task="promote" last-rc-change="Thu Oct 10 12:57:57 2019" last-run="Thu Oct 10 12:57:57 2019" exec-time="2015ms" queue-time="0ms" rc="0" rc_text="ok" />
                <operation_history call="32" task="monitor" interval="60000ms" last-rc-change="Thu Oct 10 12:58:03 2019" exec-time="3589ms" queue-time="0ms" rc="8" rc_text="master" />
            </resource_history>
            <resource_history id="rsc_ip_PRD_HDB00" orphan="false" migration-threshold="5000" fail-count="2" last-failure="Wed Oct 23 12:37:22 2019">
                <operation_history call="21" task="start" last-rc-change="Thu Oct 10 12:57:33 2019" last-run="Thu Oct 10 12:57:33 2019" exec-time="130ms" queue-time="0ms" rc="0" rc_text="ok" />
                <operation_history call="22" task="monitor" interval="10000ms" last-rc-change="Thu Oct 10 12:57:33 2019" exec-time="78ms" queue-time="0ms" rc="0" rc_text="ok" />
            </resource_history>
            <resource_history id="stonith-sbd" orphan="false" migration-threshold="5000">
                <operation_history call="6" task="start" last-rc-change="Thu Oct 10 12:57:31 2019" last-run="Thu Oct 10 12:57:31 2019" exec-time="2201ms" queue-time="0ms" rc="0" rc_text="ok" />
            </resource_history>
            <resource_history id="rsc_SAPHanaTopology_PRD_HDB00" orphan="false" migration-threshold="1">
                <operation_history call="24" task="start" last-rc-change="Thu Oct 10 12:57:39 2019" last-run="Thu Oct 10 12:57:39 2019" exec-time="4538ms" queue-time="0ms" rc="0" rc_text="ok" />
                <operation_history call="26" task="monitor" interval="10000ms" last-rc-change="Thu Oct 10 12:57:46 2019" exec-time="4220ms" queue-time="0ms" rc="0" rc_text="ok" />
            </resource_history>
        </node>
        <node name="hana02">
            <resource_history id="rsc_SAPHana_PRD_HDB00" orphan="false" migration-threshold="50" fail-count="300" last-failure="Wed Oct 23 12:37:22 2019">
                <operation_history call="22" task="start" last-rc-change="Thu Oct 17 15:22:40 2019" last-run="Thu Oct 17 15:22:40 2019" exec-time="44083ms" queue-time="0ms" rc="0" rc_text="ok" />
                <operation_history call="23" task="monitor" interval="61000ms" last-rc-change="Thu Oct 17 15:23:24 2019" exec-time="2605ms" queue-time="0ms" rc="0" rc_text="ok" />
            </resource_history>
            <resource_history id="rsc_SAPHanaTopology_PRD_HDB00" orphan="false" migration-threshold="3">
                <operation_history call="20" task="start" last-rc-change="Thu Oct 17 15:22:37 2019" last-run="Thu Oct 17 15:22:37 2019" exec-time="2905ms" queue-time="0ms" rc="0" rc_text="ok" />
                <operation_history call="21" task="monitor" interval="10000ms" last-rc-change="Thu Oct 17 15:22:40 2019" exec-time="3347ms" queue-time="0ms" rc="0" rc_text="ok" />
            </resource_history>
        </node>
    </node_history>
    <tickets>
    </tickets>
    <bans>
        <ban id="cli-ban-msl_SAPHana_PRD_HDB00-on-hana01" resource="msl_SAPHana_PRD_HDB00" node="hana01" weight="-1000000" master_only="false" />
    </bans>
</crm_mon>
EOF
