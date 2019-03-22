#!/usr/bin/python2.7

import json
import netifaces
import os
import requests
import utils


node_ip = utils.get_node_ip()
port = 63000 + int(os.getenv('OPENSHIFT_HA_VRRP_ID'))
endpoint = '%s-%d' % (node_ip, port)
url = 'https://172.30.0.1:443/api/v1/namespaces/%s/services/%s' % (
    os.getenv('NAMESPACE'), os.getenv('HAVIP_HOSTS_SVC'))
token = open('/var/run/secrets/kubernetes.io/serviceaccount/token').read()
headers = {
    'Accept': 'application/json', 'Authorization': 'Bearer %s' % token}
patch_hdrs = {'Accept': 'application/json',
              'Authorization': 'Bearer %s' % token,
              'Content-Type': 'application/strategic-merge-patch+json'}
if len(os.sys.argv) == 1 or os.sys.argv[1] == 'up':
    ep_info = os.getenv('OPENSHIFT_HA_VIRTUAL_IPS')
else:
    ep_info = ""
patch_data = json.dumps({"metadata": {"annotations": {endpoint: ep_info}}})
requests.patch(url, headers=patch_hdrs, data=patch_data, verify=False)
