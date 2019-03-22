#!/usr/bin/python2.7

import json
import utils


def app(environ, start_response):
    data = {}
    if utils.is_master():
        node_ip = utils.get_node_ip()
        tun_mac = utils.get_tun_mac()
        data.update({'node_ip': node_ip, 'tun_mac': tun_mac})
    data = json.dumps(data)
    start_response("200 OK", [
        ("Content-Type", "application/json"),
        ("Content-Length", str(len(data)))
    ])
    return iter([data])
