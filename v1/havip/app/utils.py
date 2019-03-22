import netaddr
import netifaces
import os


def get_node_ip():
    ha_intf = os.getenv('OPENSHIFT_HA_NETWORK_INTERFACE')
    if ha_intf:
        return [
            addr for addr in netifaces.ifaddresses(ha_intf)[netifaces.AF_INET]
            if addr['netmask'] != '255.255.255.255'][0]['addr']
    else:
        return os.popen("ip r get 8.8.8.8 | awk '{print $7}'").read().strip()


def get_tun_mac():
    return os.popen(
        "ip l show tun0 | awk '/ether/{print $2}'").read().strip()


def is_master():
    ha_vips = os.getenv('OPENSHIFT_HA_VIRTUAL_IPS')
    if '-' in ha_vips:
        vips = []
        for vip in ha_vips.split(','):
            if '-' in vip:
                start, end = vip.split('-')
                end = '%s.%s' % (start.rsplit('.', 1)[0], end)
                vips.extend([str(ip) for ip in netaddr.IPRange(start, end)])
            else:
                vips.append(vip)
    else:
        vips = ha_vips.split(',')
    vip_intf = os.getenv('OPENSHIFT_HA_NETWORK_INTERFACE') or (
        netifaces.gateways()['default'][netifaces.AF_INET][1])
    intf_addrs = set([addr['addr'] for addr in netifaces.ifaddresses(
        vip_intf)[netifaces.AF_INET]])
    return set(vips) - intf_addrs == set()
