#!/usr/bin/python2.7

import json
import os
import requests
import time
import utils


vips = os.getenv('OPENSHIFT_HA_VIRTUAL_IPS').split(',')
if not utils.is_master():
    # No need to do further check for backup
    os.sys.exit(0)


# origin-node restart will cause tun0 MAC changed, need to check whether
# that happened during keepalived pod running.
node_ip = utils.get_node_ip()
tun_mac = utils.get_tun_mac()
gw_changed = False
LAST_NOTIFY_PATH = '/etc/keepalived/last_notify'
if os.path.exists(LAST_NOTIFY_PATH):
    with open(LAST_NOTIFY_PATH) as f:
        last_notify = json.load(f)

    last_notify_ip = last_notify['node_ip']
    node_ip_changed = node_ip != last_notify_ip

    last_notify_mac = last_notify['tun_mac']
    tun_mac_changed = tun_mac != last_notify_mac

    if not (node_ip_changed or tun_mac_changed):
        # Neither tun0 MAC nor node IP changed, during keepalived pod running
        # skip checking
        os.sys.exit(0)


# tun0 MAC or node IP changed, need to notify controller
data = json.dumps({
    'vips': vips, 'tun_mac': tun_mac, 'node_ip': node_ip})
url = 'http://%s/master_notify' % os.getenv('HAVIP_CONTROLLER_ENDPOINT')
max_retry = 3
for i in range(max_retry):
    try:
        requests.post(url, data=data)
    except Exception:
        time.sleep(1)
    else:
        break


# record new tun0 MAC and node IP
with open(LAST_NOTIFY_PATH, 'w+') as f:
    f.write(json.dumps({'tun_mac': tun_mac, 'node_ip': node_ip}))

os.sys.exit(0)
