# OVS Centralized SNAT gateway datapath

## NOTE

Based on ovs-multitenant plugin, there are something different from ovs-multitenant and ovs-networkpolicy.

CGW stands for Centralized Gateway.

### About MASQUERADE

For ingress traffic from outside of cluster, they will be marked with 0x1, no matter via NodePort or externalIPs. And 0x1 mark will stands for do MASQUERADE.

When egress traffic from Pod to Svc or external, source will be matched, and to do MASQUERADE.

## Ingress datapath

### Iptables

#### nat-PREROUTING

##### original

    -A PREROUTING -j KUBE-SERVICES
        -A KUBE-SERVICES -d 192.168.0.11/32 -p tcp -m tcp --dport 8080 -j KUBE-MARK-MASQ
            # Mark will be set to 0x1 (like 0001)
            -A KUBE-MARK-MASQ -j MARK --set-xmark 0x1/0x1
        -A KUBE-SERVICES -d 192.168.0.11/32 -p tcp -m tcp --dport 8080 -m physdev ! --physdev-is-in -m addrtype ! --src-type LOCAL -j KUBE-SVC-YH37FP6DX24EPAZM
        -A KUBE-SERVICES -d 192.168.0.11/32 -p tcp -m tcp --dport 8080 -m addrtype --dst-type LOCAL -j KUBE-SVC-YH37FP6DX24EPAZM

The above rules will be added on all nodes.

In svc endpoint chain, DNAT will be done.

##### with CGW

    -A PREROUTING -j KUBE-SERVICES
        -A KUBE-SERVICES -d 192.168.0.11/32 -p tcp -m tcp --dport 8080 -j KUBE-MARK-CSNAT
            # Mark ingress traffic from outside to cluster wth be set to 0x2 (like 0010)
            -A KUBE-MARK-CSNAT ! -s 10.128.0.0/18 -j MARK --set-xmark 0x2/0x2
            # ingress from Pod in cluster via client VIP
            -A KUBE-MARK-CSNAT -j MARK --set-xmark 0x4/0x4
        -A KUBE-SERVICES -d 192.168.0.11/32 -p tcp -m tcp --dport 8080 -j KUBE-MARK-MASQ
            # Mark will be set to 0x3 (like 0011) or 0x5 (like 0101)
            -A KUBE-MARK-MASQ -j MARK --set-xmark 0x1/0x1
        -A KUBE-SERVICES -d 192.168.0.11/32 -p tcp -m tcp --dport 8080 -m physdev ! --physdev-is-in -m addrtype ! --src-type LOCAL -j KUBE-SVC-YH37FP6DX24EPAZM
        -A KUBE-SERVICES -d 192.168.0.11/32 -p tcp -m tcp --dport 8080 -m addrtype --dst-type LOCAL -j KUBE-SVC-YH37FP6DX24EPAZM

Only master node hosting VIP will install the above rules.

#### filter-FORWARD

    -A FORWARD -j KUBE-FORWARD
        -A KUBE-FORWARD -m mark --mark 0x1/0x1 -j ACCEPT

#### nat-POSTROUTING

##### original

NOTE: The following two chains **OPENSHIFT-MASQUERADE and KUBE-POSTROUTING could be in different order**

    -A POSTROUTING -j OPENSHIFT-MASQUERADE
        -A OPENSHIFT-MASQUERADE -s 10.128.0.0/18 -j MASQUERADE
    -A POSTROUTING -j KUBE-POSTROUTING
        -A KUBE-POSTROUTING -m mark --mark 0x1/0x1 -j MASQUERADE

##### with CGW

    -A POSTROUTING -j KUBE-MARK-CSNAT-INGRESS2
        -A KUBE-MARK-CSNAT-INGRESS2 -s 10.128.0.78/32 ! -d 10.128.0.0/18 -p tcp -m tcp -j SNAT --to-source 192.168.0.11
        -A KUBE-MARK-CSNAT-INGRESS2 -s 10.128.0.78/32 -p tcp -m tcp -m mark --mark 0x4/0x4 -j SNAT --to-source 192.168.0.11
        -A KUBE-MARK-CSNAT-INGRESS2 -s 10.128.2.116/32 ! -d 10.128.0.0/18 -p tcp -m tcp -j SNAT --to-source 192.168.0.11
        -A KUBE-MARK-CSNAT-INGRESS2 -s 10.128.2.116/32 -p tcp -m tcp -m mark --mark 0x4/0x4 -j SNAT --to-source 192.168.0.11
    -A POSTROUTING -j KUBE-POSTROUTING
        -A KUBE-POSTROUTING -m mark --mark 0x2/0x2 -j ACCEPT
        -A KUBE-POSTROUTING -m mark --mark 0x1/0x1 -j MASQUERADE
    -A POSTROUTING -j OPENSHIFT-MASQUERADE
        -A OPENSHIFT-MASQUERADE -s 10.128.0.0/18 -m mark --mark 0x2/0x2 -j ACCEPT
        -A OPENSHIFT-MASQUERADE -s 10.128.0.0/18 -j MASQUERADE

### OVS

##### original

Forwarding SDN subnet gateway IP to Pod IP. On pod host:

    table=0, priority=200, in_port=1, ip, nw_dst=10.128.0.0/18, actions=move:NXM_NX_TUN_ID[0..31]->NXM_NX_REG0[],goto_table:10

##### with CGW

Forwarding external IP to Pod IP. On pod host:

    # P2P, P2S traffic
    table=0, priority=200, in_port=1, ip, nw_src=10.128.0.0/18, nw_dst=10.128.0.0/18, actions=move:NXM_NX_TUN_ID[0..31]->NXM_NX_REG0[],goto_table:10
    # ingress traffic on CSNAT gateway node
    table=0, priority=190, in_port=1, ip, nw_dst=10.128.0.0/18, actions=move:NXM_NX_TUN_ID[0..31]->NXM_NX_REG0[],goto_table:10

## Egress datapath

### OVS

##### original

Forwarding Pod IP to External IP, locally, directly.

    table=0, priority=100, ip actions=goto_table:20
    table=20, priority=100, ip, in_port=9, nw_src=10.128.0.96 actions=load:0xd8d268->NXM_NX_REG0[],goto_table:21
    table=21, priority=0 actions=goto_table:30
    table=30, priority=0,ip actions=goto_table:100
    table=100, priority=0 actions=goto_table:101
    table=101, priority=0 actions=output:2

##### with CGW

Mark traffic egress from Pod with specified CSNAT gateway ID (e.g. 0xaabb):

    table=20, priority=100, ip, in_port=6,nw_src=10.128.0.112 actions=load:0x3f6450->NXM_NX_REG0[],load:0xaabb->NXM_NX_REG3[],goto_table:21

On non CSNAT gateway node:

    table=100, priority=10,ip,reg3=0xaabb actions=set_field:c2:f3:c9:c8:36:60->eth_dst,move:NXM_NX_TUN_ID[0..31]->NXM_NX_REG0[],set_field:192.168.0.20->tun_dst,output:1

where c2:f3:c9:c8:36:60 is tun0 mac, 192.168.0.20 is node main interface IP (used as VXLAN tunnel IP).

On CSNAT gateway node:

    table=0, priority=180,ip,in_port=1,nw_src=10.128.0.0/18 actions=move:NXM_NX_TUN_ID[0..31]->NXM_NX_REG0[],goto_table:100

### iptables

#### nat-PREROUTING

No need for egress.

#### filter-FORWARD

    -A FORWARD -j KUBE-FORWARD
        # the following packets will match the two rules
        -A KUBE-FORWARD -s 10.128.0.0/18 -m conntrack --ctstate RELATED,ESTABLISHED -j ACCEPT
        -A KUBE-FORWARD -d 10.128.0.0/18 -m conntrack --ctstate RELATED,ESTABLISHED -j ACCEPT
    -A FORWARD -j OPENSHIFT-FIREWALL-FORWARD
        -A OPENSHIFT-FIREWALL-FORWARD -s 10.128.0.0/18 -j ACCEPT

#### nat-POSTROUTING

##### original

    -A POSTROUTING -j OPENSHIFT-MASQUERADE
        -A OPENSHIFT-MASQUERADE -s 10.128.0.0/18 -j OPENSHIFT-MASQUERADE-2
            -A OPENSHIFT-MASQUERADE-2 -j MASQUERADE

##### with CGW

    -A POSTROUTING -j KUBE-MARK-CSNAT-INGRESS2
        -A KUBE-MARK-CSNAT-INGRESS2 -s 10.128.0.78/32 ! -d 10.128.0.0/18 -p tcp -m tcp -j SNAT --to-source 192.168.0.11
        -A KUBE-MARK-CSNAT-INGRESS2 -s 10.128.0.78/32 -p tcp -m tcp -m mark --mark 0x4/0x4 -j SNAT --to-source 192.168.0.11
        -A KUBE-MARK-CSNAT-INGRESS2 -s 10.128.2.116/32 ! -d 10.128.0.0/18 -p tcp -m tcp -j SNAT --to-source 192.168.0.11
        -A KUBE-MARK-CSNAT-INGRESS2 -s 10.128.2.116/32 -p tcp -m tcp -m mark --mark 0x4/0x4 -j SNAT --to-source 192.168.0.11
    -A POSTROUTING -j KUBE-POSTROUTING
    -A POSTROUTING -j OPENSHIFT-MASQUERADE
