# Keepalived 与 Controller 生命周期

## Keepalived

  - 初始化:

    - 进行服务注册，将自己的node IP，status API port，负责的VIPs注册到HAVIP_HOSTS_SVC的annotations中；
    - 启动API服务，监听status API port端口。如果当前是Master，则在被调用时，返回自己的node IP以及tun0 MAC；否则返回空。

  - 周期性执行check.py脚本：

    - 如果是backup，则直接返回成功；
    - 如果tun0 MAC或者node IP发生了变化，则调用controller的master_notify API，通告VIPs对应的最新的tun0 MAC和node IP；

  - 当自己升级为Master时，调用notify脚本，调用controller的master_notify API，通告VIPs对应的最新的tun0 MAC和node IP；
  
## Controller

  - 启动API server:
  
    - 监听master_notify接口，当有keepalived通告VIPs对应的新的tun0 MAC和node IP时，该接口调用K8S API来修改受影响的租户svc的annotations和externalIPs；
    - 监听vip接口，当有租户Namespace与VIP发生绑定，或者租户的service与VIP发生绑定时，或者解绑时，该接口调用K8S API更新HAVIP_BIND_SVC的annotations，调用K8S API更新受影响的租户svc的annotations和externalIPs。
