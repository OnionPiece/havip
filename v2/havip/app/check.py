#!/usr/bin/python2.7

import const
import netifaces
import os
import utils


if not os.path.exists(const.MASTER_VRID_PATH):
    # when notify is called with MASTER status, the file will be created
    # and writen with master vrrp_instance virtual_router_ids.
    os.sys.exit(0)

master_vrids = open(const.MASTER_VRID_PATH).read().strip()
if not master_vrids:
    # not master, no need to do further check
    os.sys.exit(0)
master_vrids = master_vrids.split(',')

# Notify script should be called for the following reasons:
#   - tun0 MAC address changed, may be caused by service origin-node restart
#   - node IP address changed, may be caused by node networking reconfigured
if os.path.exists(const.GATEWAY_INFO_PATH):
    # the same to master_vrids.dat, it should be created when notify script
    # is called
    last_gateway_info = open(const.GATEWAY_INFO_PATH).read().strip()
    new_gateway_info = '%s,%s' % (utils.get_node_ip(), utils.get_tun_mac())
    gateway_changed = new_gateway_info != last_gateway_info

    if gateway_changed:
        for vrid in master_vrids:
            os.system(
                '/etc/keepalived/app/notify.py "INSTANCE" "%s" "MASTER"' % (
                    'instance_%s' % vrid))
