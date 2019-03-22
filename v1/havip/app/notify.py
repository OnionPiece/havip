#!/usr/bin/python2.7

import json
import os
import requests
import utils


if os.sys.argv[3] != "MASTER":
    os.sys.exit(0)

node_ip = utils.get_node_ip()
tun_mac = utils.get_tun_mac()
vips = os.getenv('OPENSHIFT_HA_VIRTUAL_IPS').split(',')
data = json.dumps({
    'vips': vips, 'tun_mac': tun_mac, 'node_ip': node_ip})
url = 'http://%s/master_notify' % os.getenv('HAVIP_CONTROLLER_ENDPOINT')
requests.post(url, data=data)
