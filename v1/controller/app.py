#!/ur/bin/python2.7

import ast
from flask import Flask
from flask import request
import json
import os
import requests
import netaddr


app = Flask(__name__)


def get_bindings():
    url = 'https://172.30.0.1:443/api/v1/namespaces/%s/services/%s' % (
        os.getenv('NAMESPACE'), os.getenv('HAVIP_BINDS_SVC'))
    token = open('/var/run/secrets/kubernetes.io/serviceaccount/token').read()
    headers = {'Accept': 'application/json',
               'Authorization': 'Bearer %s' % token}
    req = requests.get(url, headers=headers, verify=False)
    return req.json()['metadata'].get('annotations', {})


def get_ns_svc(binding):
    ns, svc = None, None
    if binding:
        if binding[:2] == 'ns':
            _, ns = binding.split(':')
        else:
            _, ns, svc = binding.split(':')
    return ns, svc


def list_ip_range(iprange):
    start, end = iprange.split('-')
    end = '%s.%s' % (start.rsplit('.', 1)[0], end)
    return [str(i) for i in netaddr.IPRange(start, end)]


def vip_in_range(vip, iprange, new_range):
    # iprange e.g. 10.0.0.10,10.0.0.100-105,10.0.0.200
    if '-' not in iprange:
        return False
    in_range = False
    extend_range = []
    for _iprange in iprange.split(','):
        if '-' not in _iprange:
            continue
        iplist = list_ip_range(_iprange)
        in_range = in_range or vip in iplist
        extend_range.extend(iplist)
    if extend_range:
        new_range['found'] = extend_range
    return in_range


def get_providers(vip):
    url = 'https://172.30.0.1:443/api/v1/namespaces/%s/services/%s' % (
        os.getenv('NAMESPACE'), os.getenv('HAVIP_PROVIDERS_SVC'))
    token = open('/var/run/secrets/kubernetes.io/serviceaccount/token').read()
    headers = {'Accept': 'application/json',
               'Authorization': 'Bearer %s' % token}
    req = requests.get(url, headers=headers, verify=False)
    providers = req.json()['metadata'].get('annotations', {}).get(vip)
    if not providers:
        _url = 'https://172.30.0.1:443/api/v1/namespaces/%s/services/%s' % (
            os.getenv('NAMESPACE'), os.getenv('HAVIP_HOSTS_SVC'))
        req = requests.get(_url, headers=headers, verify=False)
        hosts = req.json()['metadata'].get('annotations', {})
        iprange = {}
        providers = ','.join([
            host for host in hosts
            if vip in hosts[host] or vip_in_range(vip, hosts[host], iprange)])
        patch_hdrs = {'Accept': 'application/json',
                      'Authorization': 'Bearer %s' % token,
                      'Content-Type': 'application/strategic-merge-patch+json'}
        if iprange:
            vips_info = {_vip: providers for _vip in iprange['found']}
        else:
            vips_info = {vip: providers}
        patch_data = json.dumps({"metadata": {"annotations": vips_info}})
        requests.patch(
            url, headers=patch_hdrs, data=patch_data, verify=False)
    return providers


def get_gateway(vip):
    providers = get_providers(vip)
    if not providers:
        # No provider is holding that VIP, since CMP should know which VIPs
        # are available for cluster.
        return "", ""
    tun_mac, node_ip = "", ""
    for provider in providers.split(','):
        ip, port = provider.split('-')
        master_info = requests.get('http://%s:%s' % (ip, port)).json()
        if master_info:
            tun_mac = master_info['tun_mac']
            node_ip = master_info['node_ip']
            break
    return tun_mac, node_ip


def notify(vip, ns, svc, tun_mac=None, node_ip=None, bind=True,
           with_enabled=True):
    if bind:
        if not tun_mac or not node_ip:
            tun_mac, node_ip = get_gateway(vip)
        if not tun_mac or not node_ip:
            return
        annotations = {'tun_mac': tun_mac, 'node_ip': node_ip}
        if with_enabled:
            annotations['cgw_enabled'] = "true"
        externalIPs = [vip]
    else:
        annotations = {'tun_mac': '', 'node_ip': '', 'cgw_enabled': "false"}
        externalIPs = []

    url = 'https://172.30.0.1:443/api/v1/namespaces/%s/services' % ns
    token = open('/var/run/secrets/kubernetes.io/serviceaccount/token').read()
    if svc:
        svcs = [svc]
    else:
        headers = {
            'Accept': 'application/json', 'Authorization': 'Bearer %s' % token}
        r = requests.get(url, headers=headers, verify=False)
        svcs = [
            _svc['metadata']['name']
            for _svc in r.json()['items']
            if vip in _svc['spec'].get('externalIPs', []) and (
                _svc['metadata'].get('annotations', {}).get(
                    'cgw_enabled') == 'true')]
    patch_hdrs = {'Accept': 'application/json',
                  'Authorization': 'Bearer %s' % token,
                  'Content-Type': 'application/strategic-merge-patch+json'}
    patch_data = json.dumps({
        "metadata": {"annotations": annotations},
        "spec": {"externalIPs": externalIPs}})
    for _svc in svcs:
        requests.patch(
            url + '/' + _svc,
            headers=patch_hdrs, data=patch_data, verify=False)


def get_bind_data(prev_ns, prev_svc, ns, svc):
    if not prev_ns and ns and not svc:
        return 'ns:%s' % ns
    elif not prev_ns and ns and svc:
        return 'sv:%s:%s' % (ns, svc)
    elif prev_ns and prev_ns == ns and not prev_svc and svc:
        # True for valid condition
        return True
    return False


def get_unbind_data(prev_ns, prev_svc, ns, svc):
    if prev_ns and prev_ns == ns:
        if prev_svc:
            if prev_svc == svc:
                return ""
        else:
            if svc:
                # True for valid condition
                return True
            else:
                return ""
    return False


def update_binding(vip, ns, svc=None, bind=True):
    prev_ns, prev_svc = get_ns_svc(get_bindings().get(vip))
    bind_info = None
    if bind:
        # NOTE: controller will not check if VIP is hosting by any providers,
        # since it should be CMP responsibility to know which VIP are valid.
        bind_info = get_bind_data(prev_ns, prev_svc, ns, svc)
    else:
        bind_info = get_unbind_data(prev_ns, prev_svc, ns, svc)
    if bind_info in (True, False):
        return bind_info

    url = 'https://172.30.0.1:443/api/v1/namespaces/%s/services/%s' % (
        os.getenv('NAMESPACE'), os.getenv('HAVIP_BINDS_SVC'))
    token = open('/var/run/secrets/kubernetes.io/serviceaccount/token').read()
    headers = {'Accept': 'application/json',
               'Authorization': 'Bearer %s' % token,
               'Content-Type': 'application/strategic-merge-patch+json'}
    data = json.dumps({"metadata": {"annotations": {vip: bind_info}}})
    requests.patch(url, headers=headers, data=data, verify=False)
    return True


@app.route('/vip', methods=['POST'])
def handle_vip():
    req = ast.literal_eval(request.data)
    vip = req['vip']
    if req.get('notify_only'):
        ns, svc = get_ns_svc(get_bindings().get(vip))
        if ns:
            notify(vip, ns, svc)
    else:
        ns = req['namespace']
        svc = req.get('service', None)
        if update_binding(vip, ns, svc):
            if svc:
                notify(vip, ns, svc)
    return 'OK'


@app.route('/vip/<vip>/<ns>', methods=['DELETE'])
@app.route('/vip/<vip>/<ns>/<svc>', methods=['DELETE'])
def unbind_vip(vip, ns, svc=None):
    if update_binding(vip, ns, svc, False):
        if svc:
            notify(vip, ns, svc, bind=False)
    return 'OK'


def notify_new_master(vips, tun_mac, node_ip):
    def _notify(vip):
        ns, svc = get_ns_svc(vip_bindings.get(vip))
        if ns:
            notify(vip, ns, svc, tun_mac, node_ip, with_enabled=False)

    vip_bindings = get_bindings()
    for vip in vips:
        if '-' in vip:
            for _vip in list_ip_range(vip):
                _notify(_vip)
        else:
            _notify(vip)
    return None, None


@app.route('/master_notify', methods=['POST'])
def master_notify():
    req = ast.literal_eval(request.data)
    vips = req['vips']
    tun_mac = req['tun_mac']
    node_ip = req['node_ip']
    notify_new_master(vips, tun_mac, node_ip)
    return "Notifactions for given VIPs are done"


if __name__ == "__main__":
    app.run(host='0.0.0.0', port=8081)
