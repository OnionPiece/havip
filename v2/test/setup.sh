#!/bin/bash

ABS_PATH="$(cd `dirname $0`; pwd)"
ETCD_CA=/tmp/havipv2-test/etcd-ca
ETCD_DATA=/tmp/havipv2-test/etcd-data
DATA=/tmp/havipv2-test/data
CONTROLLER_PATH=$ABS_PATH/../controller
HAVIP_PATH=$ABS_PATH/../havip
NOTIFY_SERVER_PATH=$ABS_PATH/fake_notify_server
TESTR_PATH=$ABS_PATH/testr


if test -d $ETCD_CA; then
    rm -rf $ETCD_CA
fi
mkdir -p $ETCD_CA

if test -d $DATA; then
    rm -rf $DATA
fi
mkdir -p $DATA
echo "token" > $DATA/token

if test -d $ETCD_DATA; then
    sudo rm -rf $ETCD_DATA
fi
mkdir -p $ETCD_DATA/{etcd1,etcd2,etcd3}

echo "Start deployer container, to deploy test env."
docker run -i --rm \
    -v $ABS_PATH/topo.yaml:/deploy.yaml \
    -v /var/run/docker.sock:/var/run/docker.sock \
    -v $ETCD_CA:/etcd-ca \
    -v $CONTROLLER_PATH:/controller \
    -v $NOTIFY_SERVER_PATH:/fake_notify_server \
    -v $TESTR_PATH:/testr \
    -v $HAVIP_PATH:/havip --name havipv2-deployer mydeployer
