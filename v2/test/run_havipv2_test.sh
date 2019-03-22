#!/bin/bash

ABS_PATH="$(cd `dirname $0`; pwd)"
CTL_PATH=$ABS_PATH/../../controller/test/
AD_INT=2 # adverttime interval for keepalived

bash $CTL_PATH/test_apis.sh clean

$ABS_PATH/get_containers.sh
if [[ $? -ne 0 ]]; then
    echo "havipv2 containers failed to start, cannot do tests."
    exit 1
fi

TASK_ID=0

function run_api(){
    cmd=$1
    shift
    $CTL_PATH/get_api.sh $cmd $@ -e 10.0.0.10:8080
    echo ""
}

node1=`docker ps --no-trunc | awk '/havipv2-1/{print $1}'`
node2=`docker ps --no-trunc | awk '/havipv2-2/{print $1}'`

function validate_keepalived() {
    TASK_ID=$((TASK_ID+1))
    expected=$1
    msg=$2
    for node in $node1 $node2; do
        observed=`docker exec -it $node sh -c "pgrep keepalived | wc -l" | cut -c 1`
        if [[ $expected == "not running" && "$observed" == "0" ]]; then
            echo "#$TASK_ID $msg: havipv2 container ${node:0:10} keepalived match $expected status" >> /proc/$$/fd/1
        elif [[ $expected == "running" && "$observed" != "0" ]]; then
            echo "#$TASK_ID $msg: havipv2 container ${node:0:10} keepalived match $expected status" >> /proc/$$/fd/1
        else
            echo "#$TASK_ID $msg: havipv2 container ${node:0:10} keepalived doesn't match $expected status" >> /proc/$$/fd/1
            exit 1
        fi
    done
}

validate_keepalived "not running" "test havipv2 keepalived not running when etcd is empty"

run_api setvrd -i eth0 -a $AD_INT -ct 3 -ci 3 -cr 3 -cf 3
validate_keepalived "not running" "test havipv2 keepalived not running when only provider/default added"

VRID1=100
VRID2=101
run_api setvr -i $VRID1 -vs 10.0.0.21 -ve 10.0.0.25
run_api setvr -i $VRID2 -vs 10.0.0.26 -ve 10.0.0.30
run_api setn -n node1 -i 10.0.0.11 -v 100,101
run_api setn -n node2 -i 10.0.0.12 -v 100,101
validate_keepalived "running" "test havipv2 keepalived running after provider info added"

node_ip=""
function validate_master() {
    TASK_ID=$((TASK_ID+1))
    expected_master_ip=$1
    vrid=$2
    expected_vips=$3
    msg=$4
    gw_info=`$CTL_PATH/test_apis.sh dump | grep "/havip/vRouterGateway/$vrid" -A 1 | grep "10.0.0"`
    node_ip=`echo $gw_info | cut -d ',' -f 1`
    if [[ $expected_master_ip != "" ]]; then
        if [[ $expected_master_ip != $node_ip ]]; then
            echo "#$TASK_ID $msg: validate keepalived master ip, not match expected $expected_master_ip" >> /proc/$$/fd/1
            exit 1
        fi
    fi
    node=$node1
    if [[ $node_ip != "10.0.0.11" ]]; then
        node=$node2
    fi
    observed=`docker exec -it $node sh -c "hostname -I"`
    echo $observed | grep -q "$node_ip.*$expected_vips"
    if [[ $? -ne 0 ]]; then
        echo "#$TASK_ID $msg: validate keepalived master, master doesn't have all VIPs" >> /proc/$$/fd/1
        exit 1
    fi
    echo "#$TASK_ID $msg: master IP and VIPs match expected" >> /proc/$$/fd/1
}
sleep_time=$((AD_INT*3+2))
echo "sleep $sleep_time seconds(advert_interval * 3 + 2) to wait one node become master..."
sleep $sleep_time
validate_master "" "100" "10.0.0.21.*10.0.0.22.*10.0.0.23.*10.0.0.24.*10.0.0.25" "test havipv2 keepalived nodes init status for virtual_router 100"
validate_master "" "101" "10.0.0.26*.10.0.0.27.*10.0.0.28.*10.0.0.29.*10.0.0.30" "test havipv2 keepalived nodes init status for virtual_router 101"

cur_master=$node1
new_master_ip=10.0.0.12
if [[ $node_ip != "10.0.0.11" ]]; then
    cur_master=$node2
    new_master_ip=10.0.0.11
fi

echo "do HA migration, and sleep $sleep_time to wait ready"
docker exec -it $cur_master sh -c "cat /var/run/keepalived.pid | xargs kill -9"
sleep $sleep_time
validate_master $new_master_ip "100" "10.0.0.21.*10.0.0.22.*10.0.0.23.*10.0.0.24.*10.0.0.25" "test havipv2 keepalived nodes status after doing harmless migration for virtual_router 100"
validate_master $new_master_ip "101" "10.0.0.26*.10.0.0.27.*10.0.0.28.*10.0.0.29.*10.0.0.30" "test havipv2 keepalived nodes status after doing harmless migration for virtual_router 101"

echo "All test passed."
