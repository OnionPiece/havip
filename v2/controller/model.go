package model

import (
	"fmt"
	"go.etcd.io/etcd/clientv3"
	"strings"
)

const (
	/* binding constants */
	VIP       = "vip"
	NAMESPACE = "namespace"
	SHARED    = "shared"
	SERVICES  = "services"
	SERVICE   = "service"

	/* node constants */
	NODEIP = "node_ip"
	VRIDS  = "vrids"

	/* virtual router constants */
	VIPS           = "vips"
	VRID           = "vrid"
	INTERFACE      = "interface"
	ADVERTINTERVAL = "advert_interval"
	STARTVIP       = "start_vip"
	ENDVIP         = "end_vip"

	/* virtual router default constants */
	CHECKINTERVAL = "check_interval"
	CHECKTIMEOUT  = "check_timeout"
	CHECKRISE     = "check_rise"
	CHECKFALL     = "check_fall"
)

type BindingRequest struct {
	Vip       string `json:"vip"`
	Namespace string `json:"namespace"`
	Service   string `json:"service,omitempty"`
}

func (br *BindingRequest) Get(key string) (string, error) {
	switch key {
	case VIP:
		if br.Vip == "" {
			return "", fmt.Errorf("No vip assigned in binding request.")
		} else {
			return br.Vip, nil
		}
	case NAMESPACE:
		if br.Namespace == "" {
			return "", fmt.Errorf("No namespace assigned in binding request.")
		} else {
			return br.Namespace, nil
		}
	case SERVICE:
		return br.Service, nil
	default:
		return "", nil
	}
}

type BindingEntity struct {
	Vip       string `json:"vip"`
	Namespace string `json:"namespace"`
	Services  string `json:"services"`
	Shared    string `json:"shared"`
	Vrid      string `json:"vrid"`
}

func (br *BindingEntity) Set(key, value string) {
	switch key {
	case VIP:
		br.Vip = value
	case NAMESPACE:
		br.Namespace = value
	case SERVICES:
		br.Services = value
	case SHARED:
		br.Shared = value
	case VRID:
		br.Vrid = value
	}
}

func (br *BindingEntity) LoadFromKVs(vip string, from map[string]string) {
	br.Vip = vip
	keys := []string{NAMESPACE, SERVICES, SHARED, VRID}
	for _, key := range keys {
		br.Set(key, from[GetBindingKey(vip, key)])
	}
}

type VirtualRouter struct {
	Vrid           string   `json:"vrid,omitempty"`
	Interface      string   `json:"interface,omitempty"`
	AdvertInterval string   `json:"advert_interval,omitempty"`
	Vips           []string `json:"vips,omitempty"`
	StartVip       string   `json:"start_vip,omitempty"`
	EndVip         string   `json:"end_vip,omitempty"`
	StaleVips      []string
}

func (vr *VirtualRouter) Set(key, value string) {
	switch key {
	case VRID:
		vr.Vrid = value
	case INTERFACE:
		vr.Interface = value
	case ADVERTINTERVAL:
		vr.AdvertInterval = value
	case VIPS:
		vr.Vips = strings.Split(value, ",")
	}
}

func (vr *VirtualRouter) LoadFromKVs(vrid string, from map[string]string) {
	vr.Vrid = vrid
	keys := []string{INTERFACE, ADVERTINTERVAL, VIPS}
	for _, key := range keys {
		vr.Set(key, from[GetVirtualRouterKey(vrid, key)])
	}
}

type NodeInfo struct {
	Node   string   `json:"node,omitempty"`
	NodeIp string   `json:"node_ip,omitempty"`
	Vrids  []string `json:"vrids,omitempty"`
}

func (ni *NodeInfo) Set(key, value string) {
	switch key {
	case NODEIP:
		ni.NodeIp = value
	case VRIDS:
		ni.Vrids = strings.Split(value, ",")
	}
}

func (ni *NodeInfo) LoadFromKVs(node string, from map[string]string) {
	ni.Node = node
	keys := []string{NODEIP, VRIDS}
	for _, key := range keys {
		ni.Set(key, from[GetNodeKey(node, key)])
	}
}

type VirtualRouterDefault struct {
	Interface      string `json:"interface,omitempty"`
	AdvertInterval string `json:"advert_interval,omitempty"`
	CheckInterval  string `json:"check_interval,omitempty"`
	CheckTimeout   string `json:"check_timeout,omitempty"`
	CheckRise      string `json:"check_rise,omitempty"`
	CheckFall      string `json:"check_fall,omitempty"`
}

func (vrd *VirtualRouterDefault) Get(key string) string {
	switch key {
	case INTERFACE:
		return vrd.Interface
	case ADVERTINTERVAL:
		return vrd.AdvertInterval
	case CHECKINTERVAL:
		return vrd.CheckInterval
	case CHECKTIMEOUT:
		return vrd.CheckTimeout
	case CHECKRISE:
		return vrd.CheckRise
	case CHECKFALL:
		return vrd.CheckFall
	default:
		return ""
	}
}

func (vrd *VirtualRouterDefault) Set(key, value string) {
	switch key {
	case INTERFACE:
		vrd.Interface = value
	case ADVERTINTERVAL:
		vrd.AdvertInterval = value
	case CHECKINTERVAL:
		vrd.CheckInterval = value
	case CHECKTIMEOUT:
		vrd.CheckTimeout = value
	case CHECKRISE:
		vrd.CheckRise = value
	case CHECKFALL:
		vrd.CheckFall = value
	}
}

func (vrd *VirtualRouterDefault) LoadFromKVs(from map[string]string) {
	keys := []string{INTERFACE, ADVERTINTERVAL, CHECKINTERVAL, CHECKTIMEOUT, CHECKRISE, CHECKFALL}
	for _, key := range keys {
		vrd.Set(key, from[GetVirtualRouterDefaultKey(key)])
	}
}

func GetBindingKey(vip string, kind string) string {
	return fmt.Sprintf("/havip/binding/%s/%s", vip, kind)
}

func GetBindingTxnOp(vip, kind, value string) clientv3.Op {
	if value == "none" {
		return clientv3.OpDelete(GetBindingKey(vip, kind), clientv3.WithPrefix())
	} else {
		return clientv3.OpPut(GetBindingKey(vip, kind), value)
	}
}

func GetNodeKey(node, kind string) string {
	return fmt.Sprintf("/havip/node/%s/%s", node, kind)
}

func GetNodeTxnOp(node, kind, value string) clientv3.Op {
	if value == "none" {
		return clientv3.OpDelete(GetNodeKey(node, kind), clientv3.WithPrefix())
	} else {
		return clientv3.OpPut(GetNodeKey(node, kind), value)
	}
}

func GetVirtualRouterKey(vrid, kind string) string {
	return fmt.Sprintf("/havip/virtual_routers/%s/%s", vrid, kind)
}

func GetVirtualRouterTxnOp(vrid, kind, value string) clientv3.Op {
	if value == "none" {
		return clientv3.OpDelete(GetVirtualRouterKey(vrid, kind), clientv3.WithPrefix())
	} else {
		return clientv3.OpPut(GetVirtualRouterKey(vrid, kind), value)
	}
}

func GetVirtualRouterDefaultKey(kind string) string {
	return fmt.Sprintf("/havip/virtual_routers/default/%s", kind)
}

func GetVirtualRouterDefaultTxnOp(key, value string) clientv3.Op {
	if value == "none" {
		return clientv3.OpDelete(GetVirtualRouterDefaultKey(key), clientv3.WithPrefix())
	} else {
		return clientv3.OpPut(GetVirtualRouterDefaultKey(key), value)
	}
}

func GetVRouterGatewayKey(vrid string) string {
	return fmt.Sprintf("/havip/vRouterGateway/%s", vrid)
}
