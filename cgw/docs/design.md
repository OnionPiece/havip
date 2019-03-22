# Centralized SNAT gateway design

## datapath

### egress

对于绑定了VIP的Svc后面的Pods，当它们要访问集群外时，流量会被转发到挂在VIP的节点（即集中式网关节点，CGW节点），再在CGW节点上做SNAT(to VIP)后转发出去。

这类Pods的流量会在OVS流表中的入口流量检查表(table 20)中打上特定的tag（向reg3写入tun_id）。在集群出口表（table 100）中，需要匹配reg3的标识，才能明确流量是否转发到CGW节点，以及对应的CGW节点在哪。在从VXLAN隧道转发出去前需要修改数据包的目的MAC地址，以确保在CGW节点上被tun0网口接受。

在CGW节点上，在OVS流表的入口流量检查表（table 0）中，通过匹配数据包的源和目的IP地址，发现数据包是egress流量，进而直接转发到集群出口表（table 100），然后从tun0流出OVS网桥。

CGW节点的iptables nat表POSTROUTING链中，将匹配到数据包的源IP地址，然后使用VIP做SNAT。

### ingress

外部流量到达CGW节点后，在iptables nat表PREROUTING链下的KUBE-SERVICES链中，通过数据包的目标IP匹配到与Svc有关的一组处理。首先，这样的流量会被打上0x2的标记；之后，通过Svc的简单负载均衡，会找到某一个Pod来做DNAT。

在iptables nat表POSTROUTING链下的OPENSHIFT-MASQUERADE或者KUBE-POSTROUTING子链中，标记0x2会被匹配，使得流量直接被接受，而不做SNAT。

之后数据包会进入OVS，后续处理一如既往。

### hairpin

例如，Svc A和Svc B都绑定了VIP，两个VIP当前都挂在同一个CGW节点上，现在Svc A后面的Pod a使用VIP来访问Svc B，Svc B下面有一个Pod b。

来自a的访问流量到达CGW节点后，经过DNAT，目标IP的地址变为b的IP，并且流量被打上0x4的标记。在iptables nat表POSTROUTING链下的OPENSHIFT-MASQUERADE或者KUBE-POSTROUTING子链中，源IP属于cluster CIDR，且具有0x4标记的流量被匹配，做SNAT，源IP变为A的VIP。

访问流量变为A的VIP访问b的IP，进入OVS，后续处理一如既往。

应答流量到达CGW节点后，利用内核的conntrack，自动进行DNAT和SNAT，目标IP变为a的IP，源IP变为B的VIP。

## HA & control plane

### keepalived

通过起keepalived Pod来提供HA。

keepalived的check script检查VIPs，节点的主IP以及tun0的MAC是否变化来判断当前Master是否发生变化，如果变化了，则将节点的主IP和tun0的MAC通知给控制面。

notify script则在Backup升级为Master时进行同样的通知。

同时keepalived Pod内会起一个web server，提供Master状态检查接口，供控制面检查Master的状态。不同DC启动keepalived会listen在不同的端口上，在Pod启动时，初始化程序会将节点IP以及listen的端口注册到控制面的服务中。

### controller

控制面通过DC来启动一组web server Pods，并且使用多个Svc来间接的将相关数据存储到etcd中（通过修改Svc的annotations实现）。

这些Svc包括:

  - hosts Svc: 用于keepalived Pods注册所在节点的IP和状态服务的监听端口；
  - binds Svc: 用于存储VIP与namespce或svc的绑定关系；
  - providers Svc: controller在运行期间，将动态从hosts Svc中提取数据，翻译成VIP与keepalived状态服务endpoints的映射关系，存储在该Svc中。

controller在运行期间，会监听vip和master_notify两个接口:

  - vip:

    - 当收到VIP绑定请求时，检查并更新binds Svc的annotations。如需要，则更新租户Svc的annotation，以指明CGW的IP以及tun0的MAC。CGW的IP和tun0的MAC获取，需要先从prodivers Svc中找到VIP所对应的keepalived状态检查endpoints，然后轮训访问，以获取最新的Master的数据。如果providers Svc中查询不到VIP所对应的keepalived endpoints，则查询hosts Svc，生成后更新到providers Svc以便之后使用；
    - 当收到VIP解绑请求时，处理过程类似，但不会获取CGW的IP和tun0的MAC，会直接使用空字符串""去刷新租户Svc annotation；
    - 当收到notify only请求时，会获取CGW的IP和tun0的MAC，获取VIP绑定的Svc，然后去更新Svc annotation。

  - master_notify: 该接口被调用，说明底层keepalived发生了主备切换，VIP漂移，相应的CGW的IP和tun0的MAC发生变化，VIP所绑定的Svc需要被更新。处理过程中会查询binds Svc。

#### binding

binds Svc中只会记录首次绑定。例如，如果一个VIP绑定了namespace，则binds中只会对namespace进行记录，而之后namespace中的多个Svc要共用VIP时，则不会再记录。

因此，如果某个namespace中的Svc要独占一个VIP，则需要在绑定时同时指定namespace和Svc。

解绑是逻辑类似，即对于Svc独占的情况，一次解绑即可完成，而共享的情况，则需要namespace下的Svc都解绑后，在进行namespace的解绑。