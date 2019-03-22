## Admin instructions

0. 编辑各个计算节点的node-yaml.config, 确认cluster-cidr被配置，配置需要重启origin-node生效:

        proxyArguments:
           cluster-cidr:
              - 10.128.0.0/18

1. 为节点打上适当的标签，用于合理的网络和资源规划。参考[vips_and_tunIDs.md] 中VIP数量上限的问题。

        oc label node test-node2 ipfo=zone1

    以下用户操作中将以ipfo=zone1作为label来为keepalived选择运行节点。

2. 创建专门用于部署havip服务的project:

        oc new-project havip

    以下步骤需要切换至havip namespace内完成
    
3. 为该项目内的default SA赋予privileged权限，以满足部署keepalived容器的需要:
    
        oc adm policy add-scc-to-user privileged -z default

4. 为该项目内的default SA赋予平台管理员的权限:

        oc adm policy add-cluster-role-to-user admin system:serviceaccount:havip:default
        or
        oc adm policy add-cluster-role-to-user admin -z default

5. 为项目内的default SA赋予system:sdn-reader的权限:

        oc adm policy add-cluster-role-to-user system:sdn-reader system:serviceaccount:havip:default
        or
        oc adm policy add-cluster-role-to-user system:sdn-reader -z default

6. 创建havip内部服务

        oc create svc clusterip havip-binds --tcp=8080:8080
        oc create svc clusterip havip-hosts --tcp=8080:8080
        oc create svc clusterip havip-providers --tcp=8080:8080

    这三个服务，将用于后续部署havip-controller和havip-keepalived时填写相应的环境变量。

7. 通过DC yaml创建havip-controller服务，创建svc和route暴露服务:

        oc create -f havip-controller.yaml
        oc expose dc havip-controller
        oc expose svc havip-controller

    havip-controller.yaml中的环境变量 **HAVIP_BINDS_SVC**, **HAVIP_HOSTS_SVC**, **HAVIP_PROVIDERS_SVC** 分别对应前一步骤中创建的三个服务。
    
8. 通过DC yaml创建havip-keepalived服务:

        oc create -f havip-keepalived.yaml

    havip-keepalived.yaml中配置参数说明:
    
      - OPENSHIFT_HA_CHECK_INTERVAL: Keepalived执行check脚本的周期；
      - OPENSHIFT_HA_VRRP_ID: 通过DC起的不同havip-keepalived服务需要被配置不同的VRRP ID，有效值[1, 255]；
      - VIP_INTERFACE: VIP所挂在的网卡；
      - HAVIP_CONTROLLER_ENDPOINT: 即havip-controller服务的svc的域名+端口；
      - **HAVIP_HOSTS_SVC**: 如前述步骤中创建的havip-hosts 服务；
      - OPENSHIFT_HA_VIRTUAL_IPS: 根据实际规划，可以一次写入多个，用逗号分割;
      - nodeSelector: 在实际部署中，需要使用步骤ii中为不同的节点组打的labels，使得HA VIP能有比较好的故障隔离域;
      - ports.containerPort & ports.hostPort: 值的计算方式为63000 + OPENSHIFT_HA_VRRP_ID的值。

## For tenant users

通过DC部署服务时，需要为Pods打上tun_id annotation, e.g.:

    spec:
      template:
        metadata:
          annotations:
            tun_id: aabb

对应的svc也需要打上相应的annotation, e.g.:

    metadata:
      annotations:
        tun_id: aabb

需要满足一个dc与一个svc对应，并且**tun_id的值是整个平台全局唯一的**，不同的dc和与之对应的svc的tun_id值不能和其他的dc及svc的重复。

tun_id值的计算方式是:

    #!/usr/bin/python2.7
    import random
    import math
    max = math.pow(2, 32)  // max = 4,294,967,296.0
    print hex(random.randint(1, max)) // => 0xdbe3dc9b
    tun_id = hex(random.randint(1, max))[2:]  // e.g. dbe3dc9b

关于tun_id地址空间的大小，参考[vips_and_tunIDs.md]。

## APIs

注: HAVIP_CONTROLLER_HOST 即havip-controller服务通过route暴露出来的地址。

Try [scripts/get_api.sh].

1. 将VIP绑定到某个namespace，例如archerprd是租户的namespace，192.168.0.11是要绑定的VIP:

        curl -X POST http://$HAVIP_CONTROLLER_HOST/vip -d '{"namespace": "archerprd", "vip": "192.168.0.11"}'

2. 在将VIP绑定到namespace后，将namespace下的某个service通过VIP进行暴露，例如pyflask是archerprd这个namespace下的service:

        curl -X POST http://$HAVIP_CONTROLLER_HOST/vip -d '{"namespace": "archerprd", "vip": "192.168.0.11", "service": "pyflask"}'

    **注1: 如果选择将VIP绑定的namespace下的单个service，而不是绑定到namespace后再由各个services共享，则需要跳过绑定namespace API，直接调用此API。**
    
    注2: 同一namespace下多个services通过同一个VIP进行暴露时，多个services的端口不能相同。

3. 强制刷新某个VIP在底层的关联配置:

        curl -X POST http://$HAVIP_CONTROLLER_HOST/vip -d '{"vip": "192.168.0.11", "notify_only": True}'

4. 解除VIP对某个service的暴露:

        curl -X DELETE http://$HAVIP_CONTROLLER_HOST/vip/<VIP>/<NAMESPACE>/<SVC>

5. 解除VIP对某个namespace的绑定:

        curl -X DELETE http://$HAVIP_CONTROLLER_HOST/vip/<VIP>/<NAMESPACE>

    注1: 如果VIP是直接绑定到单个service的，而非绑定到namespace，则只需要进行VIP到service的解绑。
    注2: 如果VIP绑定的namespace下有多个services通过VIP进行了暴露，则在VIP与namespace解绑前，需要将services都先与VIP先解绑。
