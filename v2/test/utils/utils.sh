#!/bin/bash

NET="havipv2-test-net"
NET_ID=10.0.0
NET2="havipv2-test-net2"
NET2_CIDR="172.28.0.0/14"
NOTIFY_SERVER_IP="172.30.0.1"
CA="/etcd-ca/ca.crt"
#CERT="/etcd-ca/etcd0.example.com.crt"
CERT="/etcd-ca/etcd-server.crt"
#KEY="/etcd-ca/etcd0.example.com.key"
KEY="/etcd-ca/etcd-server.key"

CONTROLLER="havipv2-controller"
HAVIP="havipv2"
NOTIFY_SERVER="fake-notify-server"
ETCD="havipv2-etcd-cluster"

function try_create_network(){
    name=$1
    cidr=$2
    docker network ls | grep -q $name
    if [[ $? -ne 0 ]]; then
        echo "Create test docker network $name..."
        docker network create --subnet $cidr $name
    else
        echo "Test docker network $name already, no need to create again."
    fi
}

function setup_network(){
    try_create_network $NET "$NET_ID.0/24"
    try_create_network $NET2 $NET2_CIDR
    echo "Insert workaround iptables rules to make $NET and $NET2 are reachable from each other."
    for id in `docker network ls | egrep "($NET|$NET2)" | awk '{print $1}'`; do
        sudo iptables -I DOCKER-ISOLATION-STAGE-2 -o br-$id -j ACCEPT
    done
}

function build_image(){
    path=""
    name=$1
    if [[ $name == $CONTROLLER ]]; then
        path="${ABS_PATH%/*}/controller"
    elif [[ $name == $HAVIP ]]; then
        path="${ABS_PATH%/*}/havip"
    elif [[ $name == $NOTIFY_SERVER ]]; then
        path="${ABS_PATH}/fake_notify_server"
    else
        echo "Unknonw image $name to build."
        exit 1
    fi
    echo "Build latest $name image..."
    pushd $path
    docker build . -t $name 2>&1 1>/proc/$$/fd/1
    if [[ $? -ne 0 ]]; then
        echo "Failed to build $name."
        popd
        exit 1
    fi
    popd
}

function setup_etcd_cluster(){
    expectedUps=3
    observedUps=`docker ps -f name=${ETCD}-* | wc -l`
    observedUps=$((observedUps-1))
    if [[ $expectedUps -ne $observedUps ]]; then
        echo "Setup etcd cluster containers..."
        # Refer:
        #   - https://coreos.com/etcd/docs/latest/v2/docker_guide.html
        #   - https://coreos.com/etcd/docs/latest/op-guide/container.html
        #   - https://github.com/kelseyhightower/etcd-production-setup
        #   - https://docs.docker.com/engine/reference/run/#cpu-period-constraint
        #   - https://www.kernel.org/doc/Documentation/scheduler/sched-bwc.txt
        for id in {1..3}; do
            docker run -d --network=$NET --ip=${NET_ID}.$((id+1)) --name ${ETCD}-${id} \
                --cpu-period=100000 --cpu-quota=20000 \
                --volume=${ABS_PATH}/etcd-ca:/etcd-ca \
                --volume=${ABS_PATH}/data/etcd${id}:/etcd-data \
                quay.io/coreos/etcd:v3.3 etcd --name etcd${id} --data-dir=/etcd-data \
                -advertise-client-urls https://$NET_ID.$((id+1)):2379 \
                -listen-client-urls https://0.0.0.0:2379 \
                -initial-advertise-peer-urls https://$NET_ID.$((id+1)):2380 \
                -listen-peer-urls https://0.0.0.0:2380 \
                -initial-cluster-token my-etcd-token \
                -initial-cluster etcd1=https://$NET_ID.2:2380,etcd2=https://$NET_ID.3:2380,etcd3=https://$NET_ID.4:2380 \
                -initial-cluster-state new \
                --ca-file $CA \
                --cert-file $CERT \
                --key-file $KEY \
                --peer-ca-file $CA \
                --peer-cert-file $CERT \
                --peer-key-file $KEY
        done
        observedUps=`docker ps -f name=${ETCD}-* | wc -l`
        observedUps=$((observedUps-1))
        if [[ $expectedUps -ne $observedUps ]]; then
            echo "No all etcd containers are UP. Setup failed."
            exit 1
        else
            echo "Verify all etcd containers running."
        fi
    else
        echo "All etcd containers are running."
    fi
}
