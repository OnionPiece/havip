package main

import (
	"context"
	"fmt"
	"go.etcd.io/etcd/clientv3"
	"go.etcd.io/etcd/mvcc/mvccpb"
	"log"
	"model"
	"os"
	"strings"
	"utils"
)

func notifyGatewayInfo(client *clientv3.Client, vrid, gwInfo string) {
	vips, err := utils.GetValue(client, model.GetVirtualRouterKey(vrid, model.VIPS))
	if err != nil {
		log.Output(2, fmt.Sprintf("[Watchd] Failed to fetch VirtualRouter VIPs for VRID: %s, since: %v", vrid, err))
		return
	}
	for _, vip := range strings.Split(vips, ",") {
		kvs, err1 := utils.GetKeyValues(client, model.GetBindingKey(vip, ""))
		if err1 != nil {
			log.Output(2, fmt.Sprintf("[Watchd] Failed to fetch binding info for VIP %s, since: %v", vip, err1))
			continue
		}
		if len(kvs) != 4 {
			/* VIP is unbound, or bound to a namespace but not to any services. */
			continue
		}
		var be model.BindingEntity
		be.LoadFromKVs(vip, kvs)
		for _, svc := range strings.Split(be.Services, ",") {
			err2 := utils.PatchHaVipForService(be.Namespace, svc, vip, gwInfo)
			if err2 != nil {
				log.Output(2, fmt.Sprintf("[Watchd] Patch request for %s.%s(ns.svc) failed, since: %v", be.Namespace, svc, err2))
			}
		}
	}
}

func main() {
	log.SetOutput(os.Stdout)
	cli := utils.GetEtcdClient()
	defer cli.Close()

	gwKey := model.GetVRouterGatewayKey("")
	vridOffset := len(gwKey)

	rch := cli.Watch(context.Background(), gwKey, clientv3.WithPrefix())
	for wresp := range rch {
		for _, ev := range wresp.Events {
			if ev.Type == mvccpb.PUT {
				notifyGatewayInfo(cli, string(ev.Kv.Key[vridOffset:]), string(ev.Kv.Value))
				log.Output(2, fmt.Sprintf("[Watchd] Recevied VirtualRouter gateway update: VRID: %s, info: %s", ev.Kv.Key[vridOffset:], ev.Kv.Value))
			}
		}
	}
}
