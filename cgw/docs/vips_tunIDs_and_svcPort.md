# 关于 VIP & tun_id & Svc Port

## 1. VIP与Service多对多的问题

对OpenShift代码的修改，没有在VIP与Service的关联上做任何限制，一个VIP可以与多个Svc关联，甚至跨namespaces。

所以，CMP需要控制一个VIP属于哪个namespace，而该namespace下这个VIP可以与哪些Svc关联，则由用户自己把握，但需要提示，与VIP绑定的Svc的端口不能重复。

## 2. VIP数量上限的问题

由于同一个广播域中，VRRP的vrid只有255个，因此如果我们的集群都在同一个广播域中，则整个集群底层的VIP分组的最多只有255个。

如果集群是跨广播域的，则理论上可以突破255个分组，但依赖底层架构的东西会显得很蠢。

(什么是广播域？我们最常见的广播风暴是ARP引起的。一个节点发起的ARP请求，被接入交换机广播，所有这个交换机下面的节点都会收到ARP请求。当接入交换机通过级联或者堆叠的方式来连接其他的交换机时，其他交换机下面的节点也会收到ARP请求，也即广播域被扩展了）

一个分组中可以有多个VIP，它们承担相同的HA。例如，网络抖动、脑裂影响了某一个分组，那么这个分组的VIP都会受影响。

CGW v1的设计是，一个VIP分组属于一个namespace。因此，255的限制就变成了全平台只有255个租户可以使用VIP了。实际上，应该理解为只有255个HA。

C(23, 2) = 253, C(24, 2) = 276。在同一个广播域内，当有23、24个节点时，它们彼此两两结对组成HA，就可以比较均匀的负载253或255个HA。这里的均匀指的是VIP因为设备故障导致不通情况下的隔离域，例如网卡坏了，节点掉电等。以一个23节点的环境为例，极端情况下，A节点最多负载22个HA，B节点最多负载21个HA，后续节点以此类推。当A、B节点都故障时，只影响1个HA。

按照这样的设计，我们可以得出节点规模与支持VIP数量的关系，即C(节点数, 2) = HA数。

  | 节点数 | HA数 |
  |----------|--------|
  |  6       |   15  |
  |  7       |   21  |
  |  8       |   28  |
  |  9       |   36  |
  |  10     |   45  |
  |  15     |  105 |
  |  20     |  190 |

进一步地，HA应该不在与namespace绑定，而由平台提供。VIP仍绑定到namespace，但VIP具体挂在到哪个HA上，则有平台负责。这样，平台能提供的最大VIP数量就可以上去。

单个网卡可以挂在超过255个IP，Linux内核支持这样做，但是过多的IP会增减内存消耗并影响性能。

节点能负载的HA，以及HA能挂在的VIP应该结合起来考虑，避免性能影响。并且考虑到HA的主备的不可预测性，因此应当以最坏情况考虑。例如，10节点的集群，HA数为45，一个节点最大负载9个HA；假设我们定义一个节点最多挂载约250个VIP，那么单个HA最大挂在27个VIP，则整个集群可以负载1215个VIP。

单个节点最250个VIP限制的情况下：

  | 节点数 | HA 数 | 单个HA最大VIP数 | 集群最大VIP数 |
  |----|----|----|----|
  |6 |15 |50 |750|
  |7 |21 |41 |861|
  |8 |28 |35 |980|
  |9 |36 |31 |1116|
  |10 |45 |27 |1215|
  |15 |105 |17 |1785|
  |20 |190 |13 |2470|

计算方式:

    #!/usr/bin/python2.7

    # n is number of nodes, m is max VIP on node
    def foo(n, m):
        ha_sum = n * (n-1) / 2
        ha_single = n - 1
        ha_vip = m / ha_single
        vip_sum = ha_vip * ha_sum
        print n, ha_sum, ha_vip, vip_sum

    for i in (6,7,8,9,10,15,20):
        foo(i, 250)

## 3. tun_id

在底层实现中，使用OVS的一个内置寄存器(reg3)来存储Pods与Svc的关联关系，借此来判断当一个Pod要访问集群外时，它的流量是否需要被Svc配置的VIP所处理，即将egress流量转发到集中式网关(Centralized SNAT Gateway，即CGW)去使用VIP做SNAT。

OVS内置寄存器的为32位，可以提供2^32(4,294,967,296.0)大小的空间。按照每个Svc一个tun_id，可以支持4亿个Svc。

## 4. Svc Port

对于绑定了VIP的Svc，无论是独占还是共享VIP，绑定后，都可以通过VIP + Svc Port的方式直接访问Svc。

这种方式不同于NodePort方式，NodePort方式会在整个集群的所有节点进行端口绑定，所以以集群内任意节点的IP + NodePort 端口都可以访问到服务。

一些需要避免冲突情况:

  - 同一个namespace，共享同一个VIP的多个Svc需要彼此使用不同的Port，否则会出现冲突，造成多个Svc只有一个能访问到的情况；
  - 如果VIP所在的底层节点上，已经有服务（系统服务或者开启了hostNetwork的容器服务）跑在了某一端口，则绑定VIP的服务则不能使用该端口，否则会造成冲突，导致绑定VIP的服务无法被访问。

在不与底层服务端口冲突的情况下，绑定不同VIP的Svc的端口可以重复。

由于需要避免Port冲突，可能会需要用户修改Svc Port，有可能需要OpenShift Svc的Port和Target Port解除一致，例如Pod的内的服务listen在8080，即Target Port是8080，而Svc Port是8081。

我们当前部署的环境中，VIP会挂在三个Master节点，可以通过如下命令来查看当前有哪些端口被listen了，即有冲突的可能:

    netstat -pnlt | awk '{print $4}' | cut -d ':' -f 2 | uniq | sort -g

作为参考，可以使用2W～3W之间的端口来作为绑定VIP的Svc可以使用的端口。