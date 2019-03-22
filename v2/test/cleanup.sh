#!/bin/bash

echo "Stop and delete containers..."
docker rm -f havipv2-controller-0 \
  havipv2-controller-1 \
  havipv2-controller-2 \
  etcd-0 \
  etcd-1 \
  etcd-2 \
  havipv2-keepalived-0 \
  havipv2-keepalived-1 \
  havipv2-fake-notify-server-0

ABS_PATH="$(cd `dirname $0`; pwd)"

for d in etcd-ca etcd-data data; do
    if test -d $ABS_PATH/$d; then
        sudo rm -rf $ABS_PATH/$d
    fi
done
