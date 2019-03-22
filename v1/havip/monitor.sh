#!/bin/bash

#  Includes.
source "$(dirname "${BASH_SOURCE[0]}")/lib/failover-functions.sh"


#
#  main():
#
setup_failover

nohup gunicorn -b 0.0.0.0:$((63000 + OPENSHIFT_HA_VRRP_ID)) -w2 --chdir /var/lib/ipfailover/keepalived/app/ status_api:app &

nohup sh -c "sleep 10; /var/lib/ipfailover/keepalived/app/svc_register.py up" &

start_failover_services

echo "`basename $0`: OpenShift IP Failover service terminated."

