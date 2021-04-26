#!/usr/bin/env bash

cat <<EOF
<?xml version="1.0"?>
<crm_mon version="2.0.0">
    <summary>
        <stack type="corosync" />
        <current_dc present="true" version="1.1.18+20180430.b12c320f5-3.15.1-b12c320f5" name="node01" id="1084783375" with_quorum="true" />
        <last_update time="Fri Oct 18 11:48:54 2019" />
        <last_change time="Fri Oct 18 11:48:22 2019" user="root" client="crm_attribute" origin="node01" />
        <nodes_configured number="2" />
        <resources_configured number="8" disabled="1" blocked="0" />
        <cluster_options stonith-enabled="true" symmetric-cluster="true" no-quorum-policy="stop" maintenance-mode="false" />
    </summary>
    <nodes>
        <node name="node01" id="1084783375" online="true" standby="false" standby_onfail="false" maintenance="false" pending="false" unclean="false" shutdown="false" expected_up="true" is_dc="true" resources_running="7" type="member" />
        <node name="node02" id="1084783376" online="true" standby="false" standby_onfail="false" maintenance="false" pending="false" unclean="false" shutdown="false" expected_up="true" is_dc="false" resources_running="5" type="member" />
    </nodes>
    <resources>
        <resource id="test-stop" resource_agent="ocf::heartbeat:Dummy" role="Stopped" target_role="Stopped" active="false" orphaned="false" blocked="false" managed="true" failed="false" failure_ignored="false" nodes_running_on="0" />
        <resource id="test" resource_agent="ocf::heartbeat:Dummy" role="Started" target_role="Started" active="true" orphaned="false" blocked="false" managed="true" failed="false" failure_ignored="false" nodes_running_on="1">
            <node name="node02" id="1084783376" cached="false"/>
        </resource>
        <resource id="stonith-sbd" resource_agent="stonith:external/sbd" role="Started" active="true" orphaned="false" blocked="false" managed="true" failed="false" failure_ignored="false" nodes_running_on="1" >
            <node name="node01" id="1084783375" cached="false"/>
        </resource>
        <resource id="rsc_ip_PRD_HDB00" resource_agent="ocf::heartbeat:IPaddr2" role="Started" active="true" orphaned="false" blocked="false" managed="true" failed="false" failure_ignored="false" nodes_running_on="1" >
            <node name="node01" id="1084783375" cached="false"/>
        </resource>
        <clone id="msl_SAPHana_PRD_HDB00" multi_state="true" unique="false" managed="true" failed="false" failure_ignored="false" >
            <resource id="rsc_SAPHana_PRD_HDB00" resource_agent="ocf::suse:SAPHana" role="Master" active="true" orphaned="false" blocked="false" managed="true" failed="false" failure_ignored="false" nodes_running_on="1" >
                <node name="node01" id="1084783375" cached="false"/>
            </resource>
            <resource id="rsc_SAPHana_PRD_HDB00" resource_agent="ocf::suse:SAPHana" role="Slave" active="true" orphaned="false" blocked="false" managed="true" failed="false" failure_ignored="false" nodes_running_on="1" pending="Monitoring" >
                <node name="node02" id="1084783376" cached="false"/>
            </resource>
        </clone>
        <clone id="cln_SAPHanaTopology_PRD_HDB00" multi_state="false" unique="false" managed="true" failed="false" failure_ignored="false" >
            <resource id="rsc_SAPHanaTopology_PRD_HDB00" resource_agent="ocf::suse:SAPHanaTopology" role="Started" active="true" orphaned="false" blocked="false" managed="true" failed="false" failure_ignored="false" nodes_running_on="1" >
                <node name="node01" id="1084783375" cached="false"/>
            </resource>
            <resource id="rsc_SAPHanaTopology_PRD_HDB00" resource_agent="ocf::suse:SAPHanaTopology" role="Started" active="true" orphaned="false" blocked="false" managed="true" failed="false" failure_ignored="false" nodes_running_on="1" >
                <node name="node02" id="1084783376" cached="false"/>
            </resource>
        </clone>
        <clone id="c-clusterfs" multi_state="false" unique="false" managed="true" failed="false" failure_ignored="false">
            <resource id="clusterfs" resource_agent="ocf::heartbeat:Filesystem" role="Started" active="true" orphaned="false" blocked="false" managed="true" failed="false" failure_ignored="false" nodes_running_on="1">
                <node name="node01" id="1084783225" cached="true"/>
            </resource>
            <resource id="clusterfs" resource_agent="ocf::heartbeat:Filesystem" role="Started" active="true" orphaned="false" blocked="false" managed="true" failed="false" failure_ignored="false" nodes_running_on="1">
                <node name="node02" id="1084783226" cached="true"/>
            </resource>
            <resource id="clusterfs" resource_agent="ocf::heartbeat:Filesystem" role="Stopped" active="false" orphaned="false" blocked="false" managed="true" failed="false" failure_ignored="false" nodes_running_on="0"/>
            <resource id="clusterfs" resource_agent="ocf::heartbeat:Filesystem" role="Stopped" active="false" orphaned="false" blocked="false" managed="true" failed="false" failure_ignored="false" nodes_running_on="0"/>
        </clone>
        <group id="grp_HA1_ASCS00" number_resources="3" >
             <resource id="rsc_ip_HA1_ASCS00" resource_agent="ocf::heartbeat:IPaddr2" role="Started" active="true" orphaned="false" blocked="false" managed="true" failed="false" failure_ignored="false" nodes_running_on="1" >
                 <node name="node01" id="1084783375" cached="false"/>
             </resource>
             <resource id="rsc_fs_HA1_ASCS00" resource_agent="ocf::heartbeat:Filesystem" role="Started" active="true" orphaned="false" blocked="false" managed="true" failed="false" failure_ignored="false" nodes_running_on="1" >
                 <node name="node01" id="1084783375" cached="false"/>
             </resource>
             <resource id="rsc_sap_HA1_ASCS00" resource_agent="ocf::heartbeat:SAPInstance" role="Started" active="true" orphaned="false" blocked="false" managed="true" failed="false" failure_ignored="false" nodes_running_on="1" >
                 <node name="node01" id="1084783375" cached="false"/>
             </resource>
        </group>
        <group id="grp_HA1_ERS10" number_resources="3" >
             <resource id="rsc_ip_HA1_ERS10" resource_agent="ocf::heartbeat:IPaddr2" role="Started" active="true" orphaned="false" blocked="false" managed="true" failed="false" failure_ignored="false" nodes_running_on="1" >
                 <node name="node02" id="1084783376" cached="false"/>
             </resource>
             <resource id="rsc_fs_HA1_ERS10" resource_agent="ocf::heartbeat:Filesystem" role="Started" active="true" orphaned="false" blocked="false" managed="true" failed="false" failure_ignored="false" nodes_running_on="1" >
                 <node name="node02" id="1084783376" cached="false"/>
             </resource>
             <resource id="rsc_sap_HA1_ERS10" resource_agent="ocf::heartbeat:SAPInstance" role="Started" active="true" orphaned="false" blocked="false" managed="true" failed="false" failure_ignored="false" nodes_running_on="1" >
                 <node name="node02" id="1084783376" cached="false"/>
             </resource>
        </group>
    </resources>
    <node_attributes>
        <node name="node01">
            <attribute name="hana_prd_clone_state" value="PROMOTED" />
            <attribute name="hana_prd_op_mode" value="logreplay" />
            <attribute name="hana_prd_remoteHost" value="node02" />
            <attribute name="hana_prd_roles" value="4:P:master1:master:worker:master" />
            <attribute name="hana_prd_site" value="PRIMARY_SITE_NAME" />
            <attribute name="hana_prd_srmode" value="sync" />
            <attribute name="hana_prd_sync_state" value="PRIM" />
            <attribute name="hana_prd_version" value="2.00.040.00.1553674765" />
            <attribute name="hana_prd_vhost" value="node01" />
            <attribute name="lpa_prd_lpt" value="1571392102" />
            <attribute name="master-rsc_SAPHana_PRD_HDB00" value="150" />
        </node>
        <node name="node02">
            <attribute name="hana_prd_clone_state" value="DEMOTED" />
            <attribute name="hana_prd_op_mode" value="logreplay" />
            <attribute name="hana_prd_remoteHost" value="node01" />
            <attribute name="hana_prd_roles" value="4:S:master1:master:worker:master" />
            <attribute name="hana_prd_site" value="SECONDARY_SITE_NAME" />
            <attribute name="hana_prd_srmode" value="sync" />
            <attribute name="hana_prd_sync_state" value="SOK" />
            <attribute name="hana_prd_version" value="2.00.040.00.1553674765" />
            <attribute name="hana_prd_vhost" value="node02" />
            <attribute name="lpa_prd_lpt" value="30" />
            <attribute name="master-rsc_SAPHana_PRD_HDB00" value="100" />
        </node>
    </node_attributes>
    <node_history>
        <node name="node01">
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
        <node name="node02">
            <resource_history id="rsc_SAPHana_PRD_HDB00" orphan="false" migration-threshold="50" fail-count="300" last-failure="Wed Oct 23 12:37:22 2019">
                <operation_history call="22" task="start" last-rc-change="Thu Oct 17 15:22:40 2019" last-run="Thu Oct 17 15:22:40 2019" exec-time="44083ms" queue-time="0ms" rc="0" rc_text="ok" />
                <operation_history call="23" task="monitor" interval="61000ms" last-rc-change="Thu Oct 17 15:23:24 2019" exec-time="2605ms" queue-time="0ms" rc="0" rc_text="ok" />
            </resource_history>
            <resource_history id="rsc_SAPHanaTopology_PRD_HDB00" orphan="false" migration-threshold="3">
                <operation_history call="20" task="start" last-rc-change="Thu Oct 17 15:22:37 2019" last-run="Thu Oct 17 15:22:37 2019" exec-time="2905ms" queue-time="0ms" rc="0" rc_text="ok" />
                <operation_history call="21" task="monitor" interval="10000ms" last-rc-change="Thu Oct 17 15:22:40 2019" exec-time="3347ms" queue-time="0ms" rc="0" rc_text="ok" />
            </resource_history>
            <resource_history id="test" orphan="false" migration-threshold="5000">
                <operation_history call="29" task="start" last-rc-change="Mon Feb 24 09:45:49 2020" last-run="Mon Feb 24 09:45:49 2020" exec-time="11ms" queue-time="0ms" rc="0" rc_text="ok" />
            </resource_history>
            <resource_history id="test-stop" orphan="false" migration-threshold="5000">
                <operation_history call="35" task="stop" last-rc-change="Mon Feb 24 09:46:58 2020" last-run="Mon Feb 24 09:46:58 2020" exec-time="12ms" queue-time="0ms" rc="0" rc_text="ok" />
            </resource_history>
        </node>
    </node_history>
    <tickets>
    </tickets>
    <bans>
    </bans>
</crm_mon>
EOF
