package model

import (
	"strconv"
	"testing"
)

func TestBindingRequestGet(t *testing.T) {
	possitiveTarget := &BindingRequest{
		Vip:       "10.0.0.2",
		Namespace: "ns1",
		Service:   "svc1",
	}
	possitiveCases := map[string]string{
		VIP:       "10.0.0.2",
		NAMESPACE: "ns1",
		SERVICE:   "svc1",
	}
	passiveTarget := &BindingRequest{
		Vip:       "",
		Namespace: "",
		Service:   "",
	}
	passiveCases := map[string]string{
		VIP:       "No vip assigned in binding request.",
		NAMESPACE: "No namespace assigned in binding request.",
		SERVICE:   "",
		"foo":     "",
	}
	for key, expectedRes := range possitiveCases {
		if value, err := possitiveTarget.Get(key); value != expectedRes && err != nil {
			t.Errorf("Expected value %s, error is nil, but get %s and %s.", expectedRes, value, err)
		}
	}
	for key, expectedRes := range passiveCases {
		if value, err := passiveTarget.Get(key); value != "" && err.Error() != expectedRes {
			t.Errorf("Expected value \"\", error with message %s, but get %s and %s.", expectedRes, value, err.Error())
		}
	}
}

func TestBindingEntitySet(t *testing.T) {
	var be BindingEntity
	beData := map[string]string{
		VIP:       "1.1.1.1",
		NAMESPACE: "ns1",
		SERVICES:  "svc1,svc2",
		SHARED:    "true",
		VRID:      "101",
	}
	for k, v := range beData {
		be.Set(k, v)
	}
	cases := map[string]string{
		be.Vip:       "1.1.1.1",
		be.Namespace: "ns1",
		be.Services:  "svc1,svc2",
		be.Shared:    "true",
		be.Vrid:      "101",
	}
	for observed, expected := range cases {
		if observed != expected {
			t.Errorf("Expected value %s, but get %s.", expected, observed)
		}
	}
}

func TestBindingEntityLoadFromKVs(t *testing.T) {
	var be BindingEntity
	beData := map[string]string{
		"/havip/binding/1.1.1.1/namespace": "ns1",
		"/havip/binding/1.1.1.1/services":  "svc1,svc2",
		"/havip/binding/1.1.1.1/shared":    "true",
		"/havip/binding/1.1.1.1/vrid":      "101",
	}
	be.LoadFromKVs("1.1.1.1", beData)
	cases := map[string]string{
		be.Vip:       "1.1.1.1",
		be.Namespace: "ns1",
		be.Services:  "svc1,svc2",
		be.Shared:    "true",
		be.Vrid:      "101",
	}
	for observed, expected := range cases {
		if observed != expected {
			t.Errorf("Expected value %s, but get %s.", expected, observed)
		}
	}
}

func TestVirtualRouterSet(t *testing.T) {
	var vr VirtualRouter
	vrData := map[string]string{
		VRID:           "101",
		INTERFACE:      "eth0",
		ADVERTINTERVAL: "2",
		VIPS:           "1.1.1.1,1.1.1.2,1.1.1.3",
	}
	for k, v := range vrData {
		vr.Set(k, v)
	}
	cases := map[string]string{
		vr.Vrid:           "101",
		vr.Interface:      "eth0",
		vr.AdvertInterval: "2",
	}
	for observed, expected := range cases {
		if observed != expected {
			t.Errorf("Expected value %s, but get %s.", expected, observed)
		}
	}
	expectedVips := []string{"1.1.1.1", "1.1.1.2", "1.1.1.3"}
	if len(vr.Vips) != len(expectedVips) {
		t.Error("Number of observed VirtualRouter VIPs not match expected.")
	}
	for idx, ob := range vr.Vips {
		if ob != expectedVips[idx] {
			t.Error("Not all observed VirtualRouter VIPs match expected.")
		}
	}
}

func TestVirtualRouterLoadFromKVs(t *testing.T) {
	var vr VirtualRouter
	vrData := map[string]string{
		"/havip/virtual_routers/101/interface":       "eth0",
		"/havip/virtual_routers/101/advert_interval": "2",
		"/havip/virtual_routers/101/vips":            "1.1.1.1,1.1.1.2,1.1.1.3",
	}
	vr.LoadFromKVs("101", vrData)
	cases := map[string]string{
		vr.Vrid:           "101",
		vr.Interface:      "eth0",
		vr.AdvertInterval: "2",
	}
	for observed, expected := range cases {
		if observed != expected {
			t.Errorf("Expected value %s, but get %s.", expected, observed)
		}
	}
	expectedVips := []string{"1.1.1.1", "1.1.1.2", "1.1.1.3"}
	if len(vr.Vips) != len(expectedVips) {
		t.Error("Number of observed VirtualRouter VIPs not match expected.")
	}
	for idx, ob := range vr.Vips {
		if ob != expectedVips[idx] {
			t.Error("Not all observed VirtualRouter VIPs match expected.")
		}
	}
}

func TestNodeInfoSet(t *testing.T) {
	var ni NodeInfo
	ni.Set(NODEIP, "10.0.0.10")
	ni.Set(VRIDS, "100,101,102")
	if ni.NodeIp != "10.0.0.10" {
		t.Error("Observed NodeInfo NodeIp not match expected")
	}
	expectedVrids := []string{"100", "101", "102"}
	if len(ni.Vrids) != len(expectedVrids) {
		t.Error("Number of observed NodeInfo Vrids not match expected.")
	}
	for idx, ob := range ni.Vrids {
		if ob != expectedVrids[idx] {
			t.Error("Not all observed NodeInfo Vrids match expected.")
		}
	}
}

func TestNodeInfoLoadFromKVs(t *testing.T) {
	var ni NodeInfo
	niData := map[string]string{
		"/havip/node/node1/node_ip": "10.0.0.10",
		"/havip/node/node1/vrids":   "100,101,102",
	}
	ni.LoadFromKVs("node1", niData)
	if ni.Node != "node1" {
		t.Error("Observed NodeInfo Node not match expected")
	}
	if ni.NodeIp != "10.0.0.10" {
		t.Error("Observed NodeInfo NodeIp not match expected")
	}
	expectedVrids := []string{"100", "101", "102"}
	if len(ni.Vrids) != len(expectedVrids) {
		t.Error("Number of observed NodeInfo Vrids not match expected.")
	}
	for idx, ob := range ni.Vrids {
		if ob != expectedVrids[idx] {
			t.Error("Not all observed NodeInfo Vrids match expected.")
		}
	}
}

func TestVirtualRouterDefaultGet(t *testing.T) {
	vrd := &VirtualRouterDefault{
		Interface:      "0",
		AdvertInterval: "1",
		CheckInterval:  "2",
		CheckTimeout:   "3",
		CheckRise:      "4",
		CheckFall:      "5",
	}
	cases := []string{INTERFACE, ADVERTINTERVAL, CHECKINTERVAL, CHECKTIMEOUT, CHECKRISE, CHECKFALL}
	for idx, k := range cases {
		if vrd.Get(k) != strconv.Itoa(idx) {
			t.Errorf("Observed VirtualRouterDefault attr %s not match expected.", k)
		}
	}
	if vrd.Get("foo") != "" {
		t.Error("Observed VirtualRouterDefault invalid attr not match expected.")
	}
}

func TestVirtualRouterDefaultSet(t *testing.T) {
	var vrd VirtualRouterDefault
	vrdData := map[string]string{
		INTERFACE:      "0",
		ADVERTINTERVAL: "1",
		CHECKINTERVAL:  "2",
		CHECKTIMEOUT:   "3",
		CHECKRISE:      "4",
		CHECKFALL:      "5",
	}
	for k, v := range vrdData {
		vrd.Set(k, v)
	}
	cases := map[string]string{
		vrd.Interface:      "0",
		vrd.AdvertInterval: "1",
		vrd.CheckInterval:  "2",
		vrd.CheckTimeout:   "3",
		vrd.CheckRise:      "4",
		vrd.CheckFall:      "5",
	}
	for observed, expected := range cases {
		if observed != expected {
			t.Errorf("Observed VirtualRouterDefault attr not match expected.")
		}
	}
}

func TestVirtualRouterDefaultLoadFromKVs(t *testing.T) {
	var vrd VirtualRouterDefault
	vrdData := map[string]string{
		"/havip/virtual_routers/default/interface":       "0",
		"/havip/virtual_routers/default/advert_interval": "1",
		"/havip/virtual_routers/default/check_interval":  "2",
		"/havip/virtual_routers/default/check_timeout":   "3",
		"/havip/virtual_routers/default/check_rise":      "4",
		"/havip/virtual_routers/default/check_fall":      "5",
	}
	vrd.LoadFromKVs(vrdData)
	cases := map[string]string{
		vrd.Interface:      "0",
		vrd.AdvertInterval: "1",
		vrd.CheckInterval:  "2",
		vrd.CheckTimeout:   "3",
		vrd.CheckRise:      "4",
		vrd.CheckFall:      "5",
	}
	for observed, expected := range cases {
		if observed != expected {
			t.Errorf("Observed VirtualRouterDefault attr not match expected.")
		}
	}
}

func TestGetBindingKey(t *testing.T) {
	cases := map[string]string{
		"":    "/havip/binding/foo/",
		"bar": "/havip/binding/foo/bar",
	}
	for key, expectedRes := range cases {
		if res := GetBindingKey("foo", key); res != expectedRes {
			t.Errorf("Expected string %s, but it was %s instead.", expectedRes, res)
		}
	}
}

func TestGetNodeKey(t *testing.T) {
	cases := map[string]string{
		"":    "/havip/node/foo/",
		"bar": "/havip/node/foo/bar",
	}
	for key, expectedRes := range cases {
		if res := GetNodeKey("foo", key); res != expectedRes {
			t.Errorf("Expected string %s, but it was %s instead.", expectedRes, res)
		}
	}
}

func TestGetVirtualRouterKey(t *testing.T) {
	cases := map[string]string{
		"":    "/havip/virtual_routers/foo/",
		"bar": "/havip/virtual_routers/foo/bar",
	}
	for key, expectedRes := range cases {
		if res := GetVirtualRouterKey("foo", key); res != expectedRes {
			t.Errorf("Expected string %s, but it was %s instead.", expectedRes, res)
		}
	}
}

func TestGetVirtualRouterDefaultKey(t *testing.T) {
	cases := map[string]string{
		"":    "/havip/virtual_routers/default/",
		"foo": "/havip/virtual_routers/default/foo",
	}
	for key, expectedRes := range cases {
		if res := GetVirtualRouterDefaultKey(key); res != expectedRes {
			t.Errorf("Expected string %s, but it was %s instead.", expectedRes, res)
		}
	}
}

func TestGetVRouterGatewayKey(t *testing.T) {
	cases := map[string]string{
		"":    "/havip/vRouterGateway/",
		"foo": "/havip/vRouterGateway/foo",
	}
	for key, expectedRes := range cases {
		if res := GetVRouterGatewayKey(key); res != expectedRes {
			t.Errorf("Expected string %s, but it was %s instead.", expectedRes, res)
		}
	}
}

func TestGetBindingTxnOp(t *testing.T) {
	delOp1 := GetBindingTxnOp("1.1.1.1", "", "none")
	if !delOp1.IsDelete() {
		t.Error("Observed GetBindingTxnOp obj1 opType not match expected.")
	}
	if string(delOp1.KeyBytes()) != "/havip/binding/1.1.1.1/" {
		t.Error("Observed GetBindingTxnOp obj1 key not match expected.")
	}
	delOp2 := GetBindingTxnOp("1.1.1.1", NAMESPACE, "none")
	if !delOp2.IsDelete() {
		t.Error("Observed GetBindingTxnOp obj2 opType not match expected.")
	}
	if string(delOp2.KeyBytes()) != "/havip/binding/1.1.1.1/namespace" {
		t.Error("Observed GetBindingTxnOp obj2 key not match expected.")
	}
	putOp := GetBindingTxnOp("1.1.1.1", NAMESPACE, "foo")
	if !putOp.IsPut() {
		t.Error("Observed GetBindingTxnOp obj3 opType not match expected.")
	}
	if string(putOp.KeyBytes()) != "/havip/binding/1.1.1.1/namespace" {
		t.Error("Observed GetBindingTxnOp obj3 key not match expected.")
	}
	if string(putOp.ValueBytes()) != "foo" {
		t.Error("Observed GetBindingTxnOp obj3 value not match expected.")
	}
}

func TestGetNodeTxnOp(t *testing.T) {
	delOp1 := GetNodeTxnOp("node1", "", "none")
	if !delOp1.IsDelete() {
		t.Error("Observed GetNodeTxnOp obj1 opType not match expected.")
	}
	if string(delOp1.KeyBytes()) != "/havip/node/node1/" {
		t.Error("Observed GetNodeTxnOp obj1 key not match expected.")
	}
	delOp2 := GetNodeTxnOp("node1", NODEIP, "none")
	if !delOp2.IsDelete() {
		t.Error("Observed GetNodeTxnOp obj2 opType not match expected.")
	}
	if string(delOp2.KeyBytes()) != "/havip/node/node1/node_ip" {
		t.Error("Observed GetNodeTxnOp obj2 key not match expected.")
	}
	putOp := GetNodeTxnOp("node1", NODEIP, "1.1.1.1")
	if !putOp.IsPut() {
		t.Error("Observed GetNodeTxnOp obj3 opType not match expected.")
	}
	if string(putOp.KeyBytes()) != "/havip/node/node1/node_ip" {
		t.Error("Observed GetNodeTxnOp obj3 key not match expected.")
	}
	if string(putOp.ValueBytes()) != "1.1.1.1" {
		t.Error("Observed GetNodeTxnOp obj3 value not match expected.")
	}
}

func TestGetVirtualRouterTxnOp(t *testing.T) {
	delOp1 := GetVirtualRouterTxnOp("101", "", "none")
	if !delOp1.IsDelete() {
		t.Error("Observed GetVirtualRouterTxnOp obj1 opType not match expected.")
	}
	if string(delOp1.KeyBytes()) != "/havip/virtual_routers/101/" {
		t.Error("Observed GetVirtualRouterTxnOp obj1 key not match expected.")
	}
	delOp2 := GetVirtualRouterTxnOp("101", INTERFACE, "none")
	if !delOp2.IsDelete() {
		t.Error("Observed GetVirtualRouterTxnOp obj2 opType not match expected.")
	}
	if string(delOp2.KeyBytes()) != "/havip/virtual_routers/101/interface" {
		t.Error("Observed GetVirtualRouterTxnOp obj2 key not match expected.")
	}
	putOp := GetVirtualRouterTxnOp("101", INTERFACE, "eth0")
	if !putOp.IsPut() {
		t.Error("Observed GetVirtualRouterTxnOp obj3 opType not match expected.")
	}
	if string(putOp.KeyBytes()) != "/havip/virtual_routers/101/interface" {
		t.Error("Observed GetVirtualRouterTxnOp obj3 key not match expected.")
	}
	if string(putOp.ValueBytes()) != "eth0" {
		t.Error("Observed GetVirtualRouterTxnOp obj3 value not match expected.")
	}
}
func TestGetVirtualRouterDefaultTxnOp(t *testing.T) {
	delOp1 := GetVirtualRouterDefaultTxnOp(CHECKRISE, "none")
	if !delOp1.IsDelete() {
		t.Error("Observed GetVirtualRouterDefaultTxnOp obj1 opType not match expected.")
	}
	if string(delOp1.KeyBytes()) != "/havip/virtual_routers/default/check_rise" {
		t.Error("Observed GetVirtualRouterDefaultTxnOp obj1 key not match expected.")
	}
	putOp := GetVirtualRouterDefaultTxnOp(CHECKRISE, "3")
	if !putOp.IsPut() {
		t.Error("Observed GetVirtualRouterDefaultTxnOp obj2 opType not match expected.")
	}
	if string(putOp.KeyBytes()) != "/havip/virtual_routers/default/check_rise" {
		t.Error("Observed GetVirtualRouterDefaultTxnOp obj2 key not match expected.")
	}
	if string(putOp.ValueBytes()) != "3" {
		t.Error("Observed GetVirtualRouterDefaultTxnOp obj2 value not match expected.")
	}
}
