#!/bin/bash

vrrp_exists=`test -f /var/run/vrrp.pid && pgrep -c -F /var/run/vrrp.pid || echo 0`
any_vrrp_instance=`grep -q vrrp_instance /etc/keepalived/keepalived.conf && echo 1 || echo 0`
state=$((vrrp_exists + any_vrrp_instance))
case $state in
    2)
        echo '[havip/v2] Reload keepalived vrrp subsystem...' > /proc/1/fd/1
        cat /var/run/vrrp.pid | xargs kill -s 1
        ;;
    1)
        echo '[havip/v2] Restart keepalived...' > /proc/1/fd/1
        cat /var/run/keepalived.pid | xargs kill -9
        keepalived -P -n -l 1>&2 2>/proc/1/fd/1
        ;;
    *)
        echo '[havip/v2] Nothing to do, since as a middle state, restart will be do when next updating come' > /proc/1/fd/1
        ;;
esac
