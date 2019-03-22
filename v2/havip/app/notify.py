#!/usr/bin/python2.7

import base64
import const
import json
import os
import requests
import utils


# $1 = "GROUP"|"INSTANCE"
# $2 = name of the group or instance
# $3 = target state of transition
#     ("MASTER"|"BACKUP"|"FAULT")
if os.sys.argv[1] == "GROUP" or os.sys.argv[3] != "MASTER":
    os.sys.exit(0)

inst_id = os.sys.argv[2].split('_', 1)[1]
gw_info = "%s,%s" % (utils.get_node_ip(), utils.get_tun_mac())
gi_encoded = base64.b64encode(gw_info)
gw_key = base64.b64encode('/havip/vRouterGateway/%s' % inst_id)

data = json.dumps({"key": gw_key, "value": gi_encoded})
ok = False
for i in range(3):
    for ep in os.getenv('ETCD_ENDPOINTS').split(','):
        url = '%s/v3beta/kv/put' % ep
        try:
            requests.post(url, data=data, verify=os.getenv('ETCD_CA'),
                          cert=(os.getenv('ETCD_CERT'), os.getenv('ETCD_KEY')))
        except Exception:
            utils.log("Notification: exception raised when post to %s for "
                      "vrrp_instance with vrid %s" % (ep, inst_id))
            time.sleep(1)
        else:
            ok = True
            break
    if ok:
        break

with open(const.GATEWAY_INFO_PATH, 'w+') as f:
    f.write(gw_info)

master_vrids = inst_id
write_master_vrids = True
if os.path.exists(const.MASTER_VRID_PATH):
    current_master_vrids = open(const.MASTER_VRID_PATH).read().strip()
    if current_master_vrids:
        if inst_id in current_master_vrids:
            write_master_vrids = False
        else:
            master_vrids = current_master_vrids + ',' + inst_id
if write_master_vrids:
    with open(const.MASTER_VRID_PATH, 'w+') as f:
        f.write(master_vrids)
