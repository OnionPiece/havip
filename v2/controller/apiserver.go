package main

import (
	"context"
	"encoding/json"
	"fmt"
	"go.etcd.io/etcd/clientv3"
	"log"
	"model"
	"net"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"
	"utils"
)

type etcdHandler struct {
	client *clientv3.Client
}

func (h *etcdHandler) bindVip(bindReq model.BindingRequest, w http.ResponseWriter) {
	vip, noVip := bindReq.Get(model.VIP)
	if noVip != nil {
		http.Error(w, noVip.Error(), http.StatusBadRequest)
		return
	}
	namespace, noNs := bindReq.Get(model.NAMESPACE)
	if noNs != nil {
		http.Error(w, noNs.Error(), http.StatusBadRequest)
		return
	}

	kvs, err := utils.GetKeyValues(h.client, model.GetBindingKey(vip, ""))
	if err != nil {
		http.Error(w, "Failed to get key-values for VIP, datastore error.", http.StatusInternalServerError)
		return
	}
	if len(kvs) == 0 {
		http.Error(w, "Invalid VIP, no VirtualRouter provides it.", http.StatusBadRequest)
		return
	}
	var be model.BindingEntity
	be.LoadFromKVs(vip, kvs)

	svc, _ := bindReq.Get(model.SERVICE)
	if be.Namespace != "" {
		if be.Namespace != namespace {
			http.Error(w, "VIP has been bound to another namespace", http.StatusForbidden)
			return
		}
		if be.Shared == "" {
			http.Error(w, "Failed to detect VIP shared level", http.StatusInternalServerError)
			return
		} else if be.Shared != "true" {
			http.Error(w, "Non-shared VIP cannot be bound with another service", http.StatusForbidden)
			return
		}
		if svc != "" {
			if be.Services != "" {
				for _, _svc := range strings.Split(be.Services, ",") {
					if _svc == svc {
						http.Error(w, "VIP has been bound to specified service, no need to bind twice", http.StatusBadRequest)
						return
					}
				}
				newSvcs := fmt.Sprintf("%s,%s", be.Services, svc)
				if err := utils.PutKeyValue(h.client, model.GetBindingKey(vip, model.SERVICES), newSvcs); err != nil {
					http.Error(w, "Failed to set service, datastore error.", http.StatusInternalServerError)
					return
				}
			} else {
				if err := utils.PutKeyValue(h.client, model.GetBindingKey(vip, model.SERVICES), svc); err != nil {
					http.Error(w, "Failed to set service, datastore error.", http.StatusInternalServerError)
					return
				}
			}
		}
	} else {
		ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
		opsNum := 2
		if svc != "" {
			opsNum = 3
		}
		ops := make([]clientv3.Op, opsNum)
		ops[0] = model.GetBindingTxnOp(vip, model.NAMESPACE, namespace)
		if svc != "" {
			ops[1] = model.GetBindingTxnOp(vip, model.SHARED, "false")
			ops[2] = model.GetBindingTxnOp(vip, model.SERVICES, svc)
		} else {
			ops[1] = model.GetBindingTxnOp(vip, model.SHARED, "true")
		}
		_, err := h.client.Txn(ctx).
			Then(ops...).
			Commit()
		cancel()
		if err != nil {
			http.Error(w, "Failed to bind vip, datastore error.", http.StatusInternalServerError)
			return
		}
	}
	if svc != "" {
		gw_info, err1 := utils.GetValue(h.client, model.GetVRouterGatewayKey(be.Vrid))
		if err1 != nil {
			http.Error(w, "Datastore updated successfully, but service patch request failed, since failed to fetch VirtualRouter gateway info.", http.StatusInternalServerError)
			return
		}
		if gw_info == "" {
			http.Error(w, "Datastore updated successfully, but service patch request failed, since VirtualRouter gateway info not found.", http.StatusInternalServerError)
			return
		}
		err2 := utils.PatchHaVipForService(namespace, svc, vip, gw_info)
		if err2 != nil {
			http.Error(w, fmt.Sprintf("Datastore updated successfully, but service patch request failed, since: %s", err2), http.StatusInternalServerError)
			return
		}
	}
	w.Write([]byte("Bind API received."))
}

func (h *etcdHandler) vipBindHandleFunc(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Unsupported method", http.StatusMethodNotAllowed)
		return
	}

	var bindReq model.BindingRequest
	err := json.NewDecoder(r.Body).Decode(&bindReq)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	h.bindVip(bindReq, w)
}

func (h *etcdHandler) unbindVip(be model.BindingEntity, vip, namespace, service string, w http.ResponseWriter, r *http.Request) {
	if be.Namespace != namespace {
		http.Error(w, "VIP is not bound to specified namespace", http.StatusForbidden)
		return
	}
	cmps := []clientv3.Cmp{}
	ops := make([]clientv3.Op, 3)
	elseOps := []clientv3.Op{}
	if service == "" {
		if be.Services != "" {
			http.Error(w, "Cannot unbound VIP, since some services in specified namespace are bound to VIP", http.StatusConflict)
			return
		}
		ops[0] = model.GetBindingTxnOp(vip, model.NAMESPACE, "none")
		ops[1] = model.GetBindingTxnOp(vip, model.SHARED, "none")
		ops = ops[:2]
	} else {
		if be.Services == "" {
			http.Error(w, "VIP in not bound to specified service, nothing to do", http.StatusNotFound)
			return
		} else {
			svcIndex := -1
			newSvcs := strings.Split(be.Services, ",")
			for index, _svc := range newSvcs {
				if _svc == service {
					svcIndex = index
				}
			}
			if svcIndex >= 0 {
				newSvcs[svcIndex] = newSvcs[len(newSvcs)-1]
				newSvcs = newSvcs[:len(newSvcs)-1]
				if len(newSvcs) == 0 {
					cmps = append(cmps, clientv3.Compare(clientv3.Value(model.GetBindingKey(vip, model.SHARED)), "=", "false"))
					ops[0] = model.GetBindingTxnOp(vip, model.SERVICES, "none")
					ops[1] = model.GetBindingTxnOp(vip, model.SHARED, "none")
					ops[2] = model.GetBindingTxnOp(vip, model.NAMESPACE, "none")
					elseOps = append(elseOps, model.GetBindingTxnOp(vip, model.SERVICES, "none"))
				} else {
					ops[0] = model.GetBindingTxnOp(vip, model.SERVICES, strings.Join(newSvcs, ","))
					ops = ops[:1]
				}
			} else {
				http.Error(w, "VIP in not bound to specified service, nothing to do", http.StatusNotFound)
				return
			}
		}
	}
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	_, err := h.client.Txn(ctx).If(cmps...).Then(ops...).Else(elseOps...).Commit()
	cancel()
	if err != nil {
		http.Error(w, "Failed to unbind vip, datastore error.", http.StatusInternalServerError)
		return
	}
	if service != "" {
		err1 := utils.PatchHaVipForService(namespace, service, "", "")
		if err1 != nil {
			http.Error(w, fmt.Sprintf("Unbound on datastore updating successfully, but service patch request failed, since: %s", err1), http.StatusInternalServerError)
			return
		}
	}
	w.Write([]byte("Unbound API received"))
}

func (h *etcdHandler) vipGetOrUnbindHandleFunc(w http.ResponseWriter, r *http.Request) {
	items := strings.Split(strings.TrimPrefix(r.URL.Path, "/"), "/")
	if !((r.Method == http.MethodGet && len(items) == 2) || (r.Method == http.MethodDelete && len(items) >= 3 && len(items) <= 4)) {
		http.Error(w, "Invalid URL or unsupported method", http.StatusMethodNotAllowed)
		return
	}
	if addr := net.ParseIP(items[1]); addr == nil {
		http.Error(w, "Invalid URL path", http.StatusMethodNotAllowed)
		return
	}
	vip := items[1]
	kvs, err := utils.GetKeyValues(h.client, model.GetBindingKey(vip, ""))
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to get binding info for vip %s, since %s", items[1], err.Error()), http.StatusInternalServerError)
		return
	}
	var be model.BindingEntity
	be.LoadFromKVs(vip, kvs)
	if r.Method == http.MethodGet {
		jsonBytes, err := json.Marshal(be)
		if err != nil {
			http.Error(w, "Failed to convert vip binding info result to json", http.StatusInternalServerError)
			return
		}
		w.Header().Set("Content-Type", "application/json; charset=utf-8")
		w.Write([]byte(jsonBytes))
	} else if len(kvs) == 0 {
		http.Error(w, "No binding records found for given VIP, nothing to do", http.StatusNotFound)
	} else {
		namespace := items[2]
		service := ""
		if len(items) == 4 {
			service = items[3]
		}
		h.unbindVip(be, vip, namespace, service, w, r)
	}
}

func (h *etcdHandler) updateVirtualRouterDefault(vrd model.VirtualRouterDefault, w http.ResponseWriter) bool {
	updated := true
	vrdKeys := []string{model.INTERFACE, model.ADVERTINTERVAL, model.CHECKTIMEOUT, model.CHECKINTERVAL, model.CHECKRISE, model.CHECKFALL}
	ops := []clientv3.Op{}
	for _, key := range vrdKeys {
		if value := vrd.Get(key); value != "" {
			ops = append(ops, model.GetVirtualRouterDefaultTxnOp(key, value))
		}
	}
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	_, err := h.client.Txn(ctx).Then(ops...).Commit()
	cancel()
	if err != nil {
		http.Error(w, "Failed to update provider defaults, datastore error.", http.StatusInternalServerError)
		return !updated
	}
	return updated
}

func (h *etcdHandler) virtualRouterDefaultHandleFunc(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodPost {
		var vrd model.VirtualRouterDefault
		err := json.NewDecoder(r.Body).Decode(&vrd)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		if !h.updateVirtualRouterDefault(vrd, w) {
			return
		}
		w.Write([]byte("Provider defaults updating API received."))
	} else if r.Method == http.MethodGet {
		kvs, err := utils.GetKeyValues(h.client, model.GetVirtualRouterDefaultKey(""))
		if err != nil {
			http.Error(w, "Failed to fetch provider default info, datastore error", http.StatusInternalServerError)
			return
		} else if len(kvs) == 0 {
			http.Error(w, "Provider default info not found", http.StatusNotFound)
			return
		}
		w.Header().Set("Content-Type", "application/json; charset=utf-8")
		var vrd model.VirtualRouterDefault
		vrd.LoadFromKVs(kvs)
		jsonBytes, err1 := json.Marshal(vrd)
		if err1 != nil {
			http.Error(w, "Failed to encode provider default info to json", http.StatusInternalServerError)
			return
		}
		w.Write(jsonBytes)
	} else {
		http.Error(w, "Unsupported method", http.StatusMethodNotAllowed)
		return
	}
}

func (h *etcdHandler) validateVipForUpdate(vr *model.VirtualRouter, w http.ResponseWriter) bool {
	valid := true
	newVips := map[string]int{}
	if len(vr.Vips) == 0 {
		if vr.StartVip != "" && vr.EndVip != "" {
			if parsedVips, err := utils.GetIPsFromRange(vr.StartVip, vr.EndVip); err != nil {
				http.Error(w, err.Error(), http.StatusBadRequest)
				return !valid
			} else {
				vr.Vips = parsedVips
				for _, vip := range parsedVips {
					newVips[vip] = 1
				}
			}
		}
	} else {
		for _, vip := range vr.Vips {
			if addr := net.ParseIP(vip); addr == nil {
				http.Error(w, "Invalid IP address format found, fail to update virtual_router.", http.StatusBadRequest)
				return !valid
			}
			newVips[vip] = 1
		}
	}
	if len(newVips) == 0 {
		http.Error(w, "Invalid to update virtual_router, since no VIPs assigned.", http.StatusInternalServerError)
		return !valid
	}
	existingVips, err := utils.GetValue(h.client, model.GetVirtualRouterKey(vr.Vrid, model.VIPS))
	if err != nil {
		http.Error(w, "Cannot detect whether virtual_router can be updated, since datastore error.", http.StatusInternalServerError)
		return !valid
	}
	if len(existingVips) == 0 {
		return valid
	}
	staleVips := []string{}
	for _, vip := range strings.Split(existingVips, ",") {
		if _, ok := newVips[vip]; !ok {
			if ns, err := utils.GetValue(h.client, model.GetBindingKey(vip, model.NAMESPACE)); err == nil && ns != "" {
				http.Error(w, fmt.Sprintf("Cannot delete VIP %s, it's in used", vip), http.StatusInternalServerError)
				return !valid
			}
			staleVips = append(staleVips, vip)
		}
	}
	if len(staleVips) != 0 {
		vr.StaleVips = staleVips
		log.Output(2, fmt.Sprintf("Found %d stale VIPs will be deleted.", len(staleVips)))
	}
	return valid
}

func (h *etcdHandler) updateVirtualRouter(vr model.VirtualRouter, w http.ResponseWriter) bool {
	updated := true
	if !h.validateVipForUpdate(&vr, w) {
		return !updated
	}
	ops := []clientv3.Op{}
	if vr.Interface != "" {
		ops = append(ops, model.GetVirtualRouterTxnOp(vr.Vrid, model.INTERFACE, vr.Interface))
	}
	if vr.AdvertInterval != "" {
		ops = append(ops, model.GetVirtualRouterTxnOp(vr.Vrid, model.ADVERTINTERVAL, vr.AdvertInterval))
	}
	if len(vr.Vips) != 0 {
		ops = append(ops, model.GetVirtualRouterTxnOp(vr.Vrid, model.VIPS, strings.Join(vr.Vips, ",")))
		for _, vip := range vr.Vips {
			ops = append(ops, model.GetBindingTxnOp(vip, model.VRID, vr.Vrid))
		}
	}
	for _, vip := range vr.StaleVips {
		ops = append(ops, model.GetBindingTxnOp(vip, model.VIP, "none"))
	}
	log.Output(2, fmt.Sprintf("%d transaction operations found for VirtualRouter updating request", len(ops)))
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	_, err := h.client.Txn(ctx).Then(ops...).Commit()
	cancel()
	if err != nil {
		http.Error(w, "Failed to update virtual_router, datastore error.", http.StatusInternalServerError)
		return !updated
	}
	return updated
}

func (h *etcdHandler) virtualRouterSettingHandleFunc(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Unsupported method", http.StatusMethodNotAllowed)
		return
	}
	var vr model.VirtualRouter
	err := json.NewDecoder(r.Body).Decode(&vr)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	log.Output(2, fmt.Sprintf("Received VirtualRouter update request: vrid=%s, vips=%v, start_vip=%s, end_vip=%s, advert_interval=%s, interface=%s", vr.Vrid, vr.Vips, vr.StartVip, vr.EndVip, vr.AdvertInterval, vr.Interface))
	if !h.updateVirtualRouter(vr, w) {
		return
	}
	w.Write([]byte("Virtual_router updating API received."))
}

func (h *etcdHandler) virtualRouterGetOrDeleteHandleFunc(w http.ResponseWriter, r *http.Request) {
	if !(r.Method == http.MethodGet || r.Method == http.MethodDelete) {
		http.Error(w, "Unsupported method", http.StatusMethodNotAllowed)
		return
	}
	items := strings.Split(strings.TrimPrefix(r.URL.Path, "/"), "/")
	if !(len(items) == 2 && items[1] != "") {
		http.Error(w, "Invalid URL", http.StatusBadRequest)
		return
	}
	vrid := items[1]
	vridInt, notInt := strconv.Atoi(vrid)
	if notInt != nil {
		http.Error(w, "Invalid URL", http.StatusBadRequest)
		return
	}
	if !(0 < vridInt && vridInt < 256) {
		http.Error(w, "Invalid URL", http.StatusBadRequest)
		return
	}
	kvs, err := utils.GetKeyValues(h.client, model.GetVirtualRouterKey(vrid, ""))
	if err != nil {
		http.Error(w, "Failed to fetch provider node, datastore error", http.StatusInternalServerError)
		return
	} else if len(kvs) == 0 {
		http.Error(w, "Virtual router not found", http.StatusNotFound)
		return
	}
	var vr model.VirtualRouter
	vr.LoadFromKVs(vrid, kvs)
	if r.Method == http.MethodGet {
		jsonBytes, err1 := json.Marshal(vr)
		if err1 != nil {
			http.Error(w, "Failed to encode virtual router to json", http.StatusInternalServerError)
			return
		}
		w.Header().Set("Content-Type", "application/json; charset=utf-8")
		w.Write(jsonBytes)
	} else if r.Method == http.MethodDelete {
		for _, vip := range vr.Vips {
			if ns, err1 := utils.GetValue(h.client, model.GetBindingKey(vip, model.NAMESPACE)); err1 == nil && ns != "" {
				http.Error(w, fmt.Sprintf("Cannot delete VirtualRouter, since VIP %s is in used", vip), http.StatusInternalServerError)
				return
			}
		}
		ops := make([]clientv3.Op, 1+len(vr.Vips))
		ops = ops[:0]
		ops = append(ops, model.GetVirtualRouterTxnOp(vrid, "", "none"))
		for _, vip := range vr.Vips {
			ops = append(ops, model.GetBindingTxnOp(vip, model.VIP, "none"))
		}
		ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
		_, err3 := h.client.Txn(ctx).Then(ops...).Commit()
		cancel()
		if err3 != nil {
			http.Error(w, "Failed to delete virtual router, datastore error", http.StatusInternalServerError)
			return
		}
		w.Write([]byte("Virtual router deletion API received."))
	}
}

/* It's not so necessary to validate VIP in using cases for node operation,
   since virtual routers may need migration, for cases like:
     - node need maintain,
     - node load is high.
*/
func (h *etcdHandler) updateNode(ni model.NodeInfo, w http.ResponseWriter) bool {
	updated := true
	ops := []clientv3.Op{}
	if ni.NodeIp != "" {
		ops = append(ops, model.GetNodeTxnOp(ni.Node, model.NODEIP, ni.NodeIp))
	}
	if len(ni.Vrids) != 0 {
		ops = append(ops, model.GetNodeTxnOp(ni.Node, model.VRIDS, strings.Join(ni.Vrids, ",")))
	}
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	_, err := h.client.Txn(ctx).Then(ops...).Commit()
	cancel()
	if err != nil {
		http.Error(w, "Failed to update node, datastore error.", http.StatusInternalServerError)
		return !updated
	}
	return updated
}

func (h *etcdHandler) nodeSettingHandleFunc(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Unsupported method", http.StatusMethodNotAllowed)
		return
	}
	var ni model.NodeInfo
	err := json.NewDecoder(r.Body).Decode(&ni)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	if ni.Node == "" {
		http.Error(w, "No node(hostname) assigned in request.", http.StatusBadRequest)
		return
	}
	if !h.updateNode(ni, w) {
		return
	}
	w.Write([]byte("Node updating API received."))
}

func (h *etcdHandler) nodeGetOrDeleteHandleFunc(w http.ResponseWriter, r *http.Request) {
	if !(r.Method == http.MethodGet || r.Method == http.MethodDelete) {
		http.Error(w, "Unsupported method", http.StatusMethodNotAllowed)
		return
	}
	items := strings.Split(strings.TrimPrefix(r.URL.Path, "/"), "/")
	if !(len(items) == 2 && items[1] != "") {
		http.Error(w, "Invalid URL", http.StatusBadRequest)
		return
	}
	node := items[1]
	kvs, err := utils.GetKeyValues(h.client, model.GetNodeKey(node, ""))
	if err != nil {
		http.Error(w, "Failed to fetch node, datastore error", http.StatusInternalServerError)
		return
	} else if len(kvs) == 0 {
		http.Error(w, "Node not found", http.StatusNotFound)
		return
	}
	if r.Method == http.MethodGet {
		var ni model.NodeInfo
		ni.LoadFromKVs(node, kvs)
		jsonBytes, err1 := json.Marshal(ni)
		if err1 != nil {
			http.Error(w, "Failed to encode node info to json", http.StatusInternalServerError)
			return
		}
		w.Header().Set("Content-Type", "application/json; charset=utf-8")
		w.Write(jsonBytes)
	} else if r.Method == http.MethodDelete {
		err1 := utils.DeleteKeys(h.client, model.GetNodeKey(node, ""))
		if err1 != nil {
			http.Error(w, "Failed to delete provider node, datastore error", http.StatusInternalServerError)
			return
		}
		w.Write([]byte("Provider deletion API received."))
	}
}

func (h *etcdHandler) vipHelperHandleFunc(w http.ResponseWriter, r *http.Request) {
	ops := []clientv3.OpOption{clientv3.WithPrefix()}
	if r.RequestURI == "/vips" {
		ops = append(ops, clientv3.WithKeysOnly())
	} else {
		ops = append(ops, clientv3.WithSort(clientv3.SortByKey, clientv3.SortAscend))
	}
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	res, err := h.client.Get(ctx, "/havip/binding/", ops...)
	cancel()
	if err != nil {
		http.Error(w, "Failed to get binding info, datastore error", http.StatusInternalServerError)
		return
	}
	if len(res.Kvs) == 0 {
		w.Write([]byte("No VIP found."))
		return
	} else {
		if r.RequestURI == "/vips" {
			keys := make([]string, len(res.Kvs))
			keys = keys[:0]
			for _, kv := range res.Kvs {
				k := string(kv.Key)
				if strings.HasSuffix(k, model.VRID) {
					keys = append(keys, strings.Split(k, "/")[3])
				}
			}
			ret := map[string][]string{"vips": keys}
			jsonBytes, err := json.Marshal(ret)
			if err != nil {
				http.Error(w, "Failed to encode vips to json", http.StatusInternalServerError)
				return
			}
			w.Write(jsonBytes)
		} else {
			lastVipWithNs := ""
			vipNotUsed := ""
			for _, kv := range res.Kvs {
				k := string(kv.Key)
				if strings.HasSuffix(k, model.NAMESPACE) {
					lastVipWithNs = strings.Split(k, "/")[3]
				} else if strings.HasSuffix(k, model.VRID) {
					vipNotUsed = strings.Split(k, "/")[3]
					if vipNotUsed != lastVipWithNs {
						break
					}
				}
			}
			ret := map[string]string{"vip": vipNotUsed}
			jsonBytes, err := json.Marshal(ret)
			if err != nil {
				http.Error(w, "Failed to encode vip to json", http.StatusInternalServerError)
				return
			}
			w.Write(jsonBytes)
		}
	}
}

func (h *etcdHandler) getNodesHelperHandleFunc(w http.ResponseWriter, r *http.Request) {
	kvs, err := utils.GetKeyValues(h.client, model.GetNodeKey("", ""))
	if err != nil {
		http.Error(w, "Failed to get nodes, datastore error", http.StatusInternalServerError)
		return
	}
	nodes := map[string]map[string]string{}
	for k, v := range kvs {
		paths := strings.Split(k, "/")
		nodeName := paths[3]
		subKey := paths[4]
		if _, ok := nodes[nodeName]; !ok {
			nodes[nodeName] = map[string]string{}
		}
		nodes[nodeName][subKey] = v
	}
	if len(nodes) == 0 {
		w.Write([]byte("No node found."))
	} else {
		jsonBytes, err := json.Marshal(nodes)
		if err != nil {
			http.Error(w, "Failed to encode nodes to json", http.StatusInternalServerError)
			return
		}
		w.Write(jsonBytes)
	}
}

func (h *etcdHandler) getVRsHelperHandleFunc(w http.ResponseWriter, r *http.Request) {
	kvs, err := utils.GetKeyValues(h.client, model.GetVirtualRouterKey("", ""))
	if err != nil {
		http.Error(w, "Failed to get VirtualRouters, datastore error", http.StatusInternalServerError)
		return
	}
	vRouters := map[string]string{}
	for k, v := range kvs {
		if strings.HasSuffix(k, "vips") {
			vrid := strings.Split(k, "/")[3]
			vRouters[vrid] = v
		}
	}
	if len(vRouters) == 0 {
		w.Write([]byte("No VirtualRouters found."))
	} else {
		jsonBytes, err := json.Marshal(vRouters)
		if err != nil {
			http.Error(w, "Failed to encode VirtualRouters to json", http.StatusInternalServerError)
			return
		}
		w.Write(jsonBytes)
	}
}

func main() {
	log.SetOutput(os.Stdout)
	cli := utils.GetEtcdClient()
	defer cli.Close()

	handler := &etcdHandler{client: cli}

	http.HandleFunc("/vip", handler.vipBindHandleFunc)
	http.HandleFunc("/vip/", handler.vipGetOrUnbindHandleFunc)
	http.HandleFunc("/virtual_router_default", handler.virtualRouterDefaultHandleFunc)
	http.HandleFunc("/virtual_router", handler.virtualRouterSettingHandleFunc)
	http.HandleFunc("/virtual_router/", handler.virtualRouterGetOrDeleteHandleFunc)
	http.HandleFunc("/node", handler.nodeSettingHandleFunc)
	http.HandleFunc("/node/", handler.nodeGetOrDeleteHandleFunc)
	http.HandleFunc("/vips", handler.vipHelperHandleFunc)
	http.HandleFunc("/get_one_vip", handler.vipHelperHandleFunc)
	http.HandleFunc("/nodes", handler.getNodesHelperHandleFunc)
	http.HandleFunc("/virtual_routers", handler.getVRsHelperHandleFunc)
	if err := http.ListenAndServe(":8080", nil); err != nil {
		panic(err)
	}
}
