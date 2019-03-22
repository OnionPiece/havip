## Refers:

Except golang doc, also ref:

  - https://gist.github.com/michaljemala/d6f4e01c4834bf47a9c4
  - https://gist.github.com/aodin/9493190
  - https://github.com/etcd-io/etcd/pull/10044#issuecomment-417125341
  - https://stackoverflow.com/questions/15672556/handling-json-post-request-in-go
  - https://github.com/etcd-io/etcd/blob/master/clientv3/naming/grpc_test.go#L126
  - https://github.com/etcd-io/etcd/blob/master/clientv3/example_kv_test.go#L238-L245
  - https://github.com/etcd-io/etcd/blob/master/clientv3/integration/black_hole_test.go#L130-L133
  - https://github.com/etcd-io/etcd/blob/master/clientv3/clientv3util/example_key_test.go#L59-L62
  - https://godoc.org/github.com/coreos/etcd/clientv3
  - ... more may miss to list.

## APIs:

Check test/get_apis.sh and test/test_apis.sh to see APIs details.

## Datastore:

Only for etcdv3.

### VIPS Bindings

Prefix: /havip/binding/VIP/ , e.g. /havip/binding/10.0.0.3/ where 10.0.0.3 is a VIP.

| Key | Default | Value example | Description |
|--|--|--|--|
|namespace|nil| - | K8S which VIP bound to.|
|shared|nil|"true" or "false"| Is VIP shared to multiple services in namespace or dedicated to a service.|
|services|nil|svc1 or svc1,svc2,svc3| Which service(s) in namespace will  use this VIP.|
|vrid|nil|100| VRID of virtual router which provide this VIP.|

### Virtual Router Default

Prefix: /havip/virtual_router/default/

| Key | Default | Value example | Description |
|--|--|--|--|
|interface|nil|eth0| Which interface keepalived will use to setup VIPs and send VRRPs.|
|advert_interval|nil|2| Interval will be used by keepalived to send VRRP advertisement.|
|check_interval|nil|2| Interval will be used by keepalived to invoke check script.|
|check_timeout|nil|3| Timeout used by check script.|
|check_rise|nil|3| Keepalived check script rise.|
|check_fall|nil|3| Keepalived check script fall.|

### Virtual Router

Prefix: /havip/virtual_router/VRID/ , e.g. /havip/virtual_router/100/ ,  VRID should be in [1, 255].

|Key | Default | Value example | Description |
|--|--|--|--|
|vips|nil|10.0.0.3,10.0.0.4| Vips will be supplied by virtual router.|
|interface|/havip/virtual_router/-<br>default/interface|eth0| On which interface will VIPs be added on.|
|advert_interval|/havip/virtual_router/-<br>default/advert_interval|3|  Interval will be used by keepalived to send VRRP advertisement.|


### Node Info

Prefix: /havip/node/NODENAME/ , e.g. /havip/node/node1/ where node1 is a node running havip/keepalived and provide VIPs.

|Key | Default | Value example | Description |
|--|--|--|--|
|node_ip|nil|1.2.3.4| IP of node, used as Centralized SNAT Gateway tunnel Port.|
|vrids|nil|100,101,102| VRIDs of virtual routers running on the node.|

### Virtual Router Gateway

Prefix: /havip/vRouterGateway/ , e.g. /havip/vRouterGateway/101 where 101 is VRID.

|Key | Default | Value example | Description |
|--|--|--|--|
|$\{prefix}/VRID| nil | 10.0.0.10,ab:cd:ef:12:34:56| For VirtualRouter whos VRID is 101, currently it's on node whos node IP is 10.0.0.10 and tun0 MAC is ab:cd:ef:12:34:56.|
