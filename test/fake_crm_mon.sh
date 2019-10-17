#!/usr/bin/env bash

cat <<EOF
<?xml version="1.0"?>
<crm_mon version="2.0.0">
  <summary>
    <stack type="corosync" />
    <current_dc present="true" version="2.0.0+20181108.62ffcafbc-1.1-2.0.0+20181108.62ffcafbc" name="Hawk3-1" id="168430211" with_quorum="true" />
    <last_update time="Tue Jan 15 22:20:05 2019" />
    <last_change time="Tue Jan 15 22:19:59 2019" user="root" client="cibadmin" origin="Hawk3-2" />
    <nodes_configured number="2" />
    <resources_configured number="3" disabled="0" blocked="0" />
    <cluster_options stonith-enabled="false" symmetric-cluster="true" no-quorum-policy="stop" maintenance-mode="false" />
  </summary>
  <nodes>
    <node name="Hawk3-1" id="168430211" online="true" standby="false" standby_onfail="false" maintenance="false" pending="false" unclean="false" shutdown="false" expected_up="true" is_dc="true" resources_running="1" type="member" />
    <node name="Hawk3-2" id="168430212" online="true" standby="false" standby_onfail="false" maintenance="false" pending="false" unclean="false" shutdown="false" expected_up="true" is_dc="false" resources_running="1" type="member" />
  </nodes>
  <resources>
    <resource id="d1" resource_agent="ocf::heartbeat:Dummy" role="Started" active="true" orphaned="false" blocked="false" managed="true" failed="false" failure_ignored="false" nodes_running_on="1" >
      <node name="Hawk3-1" id="168430211" cached="false"/>
    </resource>
    <resource id="vip1" resource_agent="ocf::heartbeat:IPaddr2" role="Started" active="true" orphaned="false" blocked="false" managed="true" failed="false" failure_ignored="false" nodes_running_on="1" >
      <node name="Hawk3-2" id="168430212" cached="false"/>
    </resource>
  </resources>
  <node_attributes>
    <node name="Hawk3-1">
    </node>
    <node name="Hawk3-2">
    </node>
  </node_attributes>
  <node_history>
    <node name="Hawk3-2">
      <resource_history id="vip1" orphan="false" migration-threshold="3">
        <operation_history call="10" task="start" last-rc-change="Tue Jan 15 22:19:14 2019" last-run="Tue Jan 15 22:19:14 2019" exec-time="66ms" queue-time="0ms" rc="0" rc_text="ok" />
        <operation_history call="11" task="monitor" interval="10000ms" last-rc-change="Tue Jan 15 22:19:15 2019" exec-time="34ms" queue-time="0ms" rc="0" rc_text="ok" />
      </resource_history>
      <resource_history id="ddd" orphan="false" migration-threshold="3" fail-count="1000000" last-failure="Tue Jan 15 22:20:00 2019">
        <operation_history call="16" task="start" last-rc-change="Tue Jan 15 22:19:59 2019" last-run="Tue Jan 15 22:19:59 2019" exec-time="34ms" queue-time="0ms" rc="5" rc_text="not installed" />
        <operation_history call="17" task="stop" last-rc-change="Tue Jan 15 22:19:59 2019" last-run="Tue Jan 15 22:19:59 2019" exec-time="36ms" queue-time="0ms" rc="0" rc_text="ok" />
      </resource_history>
    </node>
    <node name="Hawk3-1">
      <resource_history id="d1" orphan="false" migration-threshold="3">
        <operation_history call="10" task="start" last-rc-change="Tue Jan 15 22:19:15 2019" last-run="Tue Jan 15 22:19:15 2019" exec-time="23ms" queue-time="0ms" rc="0" rc_text="ok" />
      </resource_history>
      <resource_history id="ddd" orphan="false" migration-threshold="3" fail-count="1000000" last-failure="Tue Jan 15 22:19:59 2019">
        <operation_history call="15" task="start" last-rc-change="Tue Jan 15 22:19:59 2019" last-run="Tue Jan 15 22:19:59 2019" exec-time="45ms" queue-time="0ms" rc="5" rc_text="not installed" />
        <operation_history call="16" task="stop" last-rc-change="Tue Jan 15 22:19:59 2019" last-run="Tue Jan 15 22:19:59 2019" exec-time="38ms" queue-time="0ms" rc="0" rc_text="ok" />
      </resource_history>
    </node>
  </node_history>
  <failures>
    <failure op_key="ddd_start_0" node="Hawk3-2" exitstatus="not installed" exitreason="Setup problem: couldn&apos;t find command: /usr/bin/safe_mysqld" exitcode="5" call="16" status="complete" last-rc-change="Tue Jan 15 22:19:59 2019" queued="0" exec="34" interval="0" task="start" />
    <failure op_key="ddd_start_0" node="Hawk3-1" exitstatus="not installed" exitreason="Setup problem: couldn&apos;t find command: /usr/bin/safe_mysqld" exitcode="5" call="15" status="complete" last-rc-change="Tue Jan 15 22:19:59 2019" queued="0" exec="45" interval="0" task="start" />
  </failures>
  <fence_history>
  </fence_history>
  <tickets>
  </tickets>
  <bans>
  </bans>
</crm_mon>
EOF
