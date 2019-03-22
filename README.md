#### CGW

CGW stands for centralized  gateway, which means Pods egress traffic will be SNAT-ed by Service first external IP, and for ingress traffic, they can be access Service via the same external IP.

Currently it's only for origin v3.9 with redhat/openshift-ovs-multitenant plugin. For redhat/openshift-ovs-networkpolicy plugin, technical logic is similar, but OVS egress traffic flows should be taken more care of.

#### havip v1

It's a demo, but still usable for small scale.

It has two parts, controllers and keepalived pods. Controllers are deployed via Deployments and exposed via Service and Route to accept APIs. This version doesn't have any datastore, it uses some nit Services and store data like binding, node info into Services annotations. So data will be finally stored into etcd cluster, but it's stupid to do this via OpenShift/K8S APIs.

Another bad thing is, origin ipfailover/keepalived is inherited, which means it doesn't have any maintainability in VIPs scaleup.

#### v2

Main difference from havip:

1. controller is writen by Golang, since no suitable python library is found as etcd v3 client;
2. store data into etcd cluster, which is shared by origin;
3. using confd to control keepalived configures.
