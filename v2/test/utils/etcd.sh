#/bin/bash

ETCDCTL=/usr/local/bin/etcdctl3

ABS_PATH="$(cd `dirname $0`; pwd)"
ETCDCTL_ENDPOINTS="https://10.0.0.2:2379,https://10.0.0.3:2379,https://10.0.0.4:2379"
ETCDCTL_CA_FILE=$ABS_PATH/../etcd-ca/ca.crt
ETCDCTL_CERT_FILE=$ABS_PATH/../etcd-ca/etcd-client.crt
ETCDCTL_KEY_FILE=$ABS_PATH/../etcd-ca/etcd-client.key
API_ENDPOINT=http://10.0.0.10:8080

function validate_ctl(){
    ls $ETCDCTL 1>&2 2>/dev/null
    if [[ $? -ne 0 ]]; then
        echo "etcdctl v3 CLI not found, try to modify $ABS_PATH/etch.sh ETCDCTL to correct path."
        exit 1
    fi
}

function run_ctl(){
    ETCDCTL_API=3 $ETCDCTL --endpoints=$ETCDCTL_ENDPOINTS --cacert=$ETCDCTL_CA_FILE --cert=$ETCDCTL_CERT_FILE --key=$ETCDCTL_KEY_FILE $@
}

function etcd_cleanup() {
    run_ctl "del --prefix /havip/ > /dev/null"
}

function etcd_dump() {
    run_ctl "get --prefix $1"
}

function etcd_validate_value() {
    key=$1
    expectedValue=$2
    observedValue=`run_ctl get --print-value-only $key`
    test "$expectedValue" == "$observedValue" && echo "0" || echo "1"
}
