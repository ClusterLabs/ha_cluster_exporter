#!/usr/bin/env bash

cat <<EOF
<xpath-query>
  <rsc_location id="cli-prefer-msl_SAPHana_PRD_HDB00" rsc="msl_SAPHana_PRD_HDB00" role="Started" node="damadog-hana01" score="INFINITY"/>
  <rsc_location id="cli-prefer-cln_SAPHanaTopology_PRD_HDB00" rsc="cln_SAPHanaTopology_PRD_HDB00" role="Started" node="damadog-hana01" score="INFINITY"/>
</xpath-query>
EOF
