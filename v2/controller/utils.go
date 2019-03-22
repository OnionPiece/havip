package utils

import (
	"bytes"
	"context"
	"crypto/tls"
	"crypto/x509"
	"encoding/json"
	"flag"
	"fmt"
	"go.etcd.io/etcd/clientv3"
	"io/ioutil"
	"log"
	"net"
	"net/http"
	"os"
	"strings"
	"time"
)

func getEtcdTLSConfig() *tls.Config {
	certFile := flag.String("cert", os.Getenv("ETCD_CERT"), "A PEM eoncoded certificate file.")
	keyFile := flag.String("key", os.Getenv("ETCD_KEY"), "A PEM encoded private key file.")
	caFile := flag.String("CA", os.Getenv("ETCD_CA"), "A PEM eoncoded CA's certificate file.")

	cert, err := tls.LoadX509KeyPair(*certFile, *keyFile)
	if err != nil {
		log.Fatal("Fail to load etcd-client cert and key, since: ", err)
	}

	caCert, err := ioutil.ReadFile(*caFile)
	if err != nil {
		log.Fatal("Fail to load etcd ca file, since: ", err)
	}
	rootCAs := x509.NewCertPool()
	rootCAs.AppendCertsFromPEM(caCert)

	cfg := &tls.Config{
		Certificates: []tls.Certificate{cert},
		RootCAs:      rootCAs,
	}
	return cfg
}

func GetEtcdClient() *clientv3.Client {
	cli, err := clientv3.New(clientv3.Config{
		Endpoints:   strings.Split(os.Getenv("ETCD_ENDPOINTS"), ","),
		DialTimeout: 5 * time.Second,
		TLS:         getEtcdTLSConfig(),
	})
	if err != nil {
		log.Fatal("Fail to create etcd clientv3, since: %v", err)
	}
	return cli
}

func GetValue(cli *clientv3.Client, k string) (string, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	res, err := cli.Get(ctx, k)
	cancel()
	if err != nil {
		return "", err
	}
	if len(res.Kvs) == 0 {
		return "", nil
	} else {
		return string(res.Kvs[0].Value), nil
	}
}

func GetKeyValues(cli *clientv3.Client, k string) (map[string]string, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	res, err := cli.Get(ctx, k, clientv3.WithPrefix())
	cancel()
	if err != nil {
		return nil, err
	}
	kvs := make(map[string]string)
	for _, kv := range res.Kvs {
		kvs[string(kv.Key[:])] = string(kv.Value[:])
	}
	return kvs, nil
}

func PutKeyValue(cli *clientv3.Client, k, v string) error {
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	_, err := cli.Put(ctx, k, v)
	cancel()
	return err
}

func DeleteKey(cli *clientv3.Client, k string) error {
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	_, err := cli.Delete(ctx, k)
	cancel()
	return err
}

func DeleteKeys(cli *clientv3.Client, prefix string) error {
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	_, err := cli.Delete(ctx, prefix, clientv3.WithPrefix())
	cancel()
	return err
}

func GetIPsFromRange(start, end string) ([]string, error) {
	if start > end {
		swap := start
		start = end
		end = swap
	}
	firstIp := net.ParseIP(start)
	if firstIp == nil {
		return []string{}, fmt.Errorf("Invalid IP format as IP range start")
	}
	lastIp := net.ParseIP(end)
	if lastIp == nil {
		return []string{}, fmt.Errorf("Invalid IP format as IP range end")
	}
	if firstIp[12] != lastIp[12] || firstIp[13] != lastIp[13] {
		return []string{}, fmt.Errorf("Not for that large range.")
	}
	a := firstIp[12]
	b := firstIp[13]
	c := firstIp[14]
	d := firstIp[15]
	tail := byte(255)
	size := 256*int(lastIp[14]-c) + int(lastIp[15]-d) + 1
	ipStrList := make([]string, size)
	ipStrList = ipStrList[:0]
	for c != lastIp[14] || d != lastIp[15] {
		ipStrList = append(ipStrList, fmt.Sprintf("%s", net.IPv4(a, b, c, d)))
		if d == tail {
			c = c + 1
			d = 0
			continue
		}
		d = d + 1
	}
	ipStrList = append(ipStrList, end)
	return ipStrList, nil
}

type HaVipPatchData struct {
	Metadata map[string]map[string]string `json:"metadata"`
	Spec     map[string][]string          `json:"spec"`
}

func PatchHaVipForService(namespace, service, vip, gwInfo string) error {
	tr := &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
	}
	client := &http.Client{Transport: tr}

	url := fmt.Sprintf("https://172.30.0.1:443/api/v1/namespaces/%s/services/%s", namespace, service)
	tokenBytes, err := ioutil.ReadFile("/var/run/secrets/kubernetes.io/serviceaccount/token")
	if err != nil {
		return fmt.Errorf("Failed to load token file, since: %v", err)
	}
	token := fmt.Sprintf("Bearer %s", tokenBytes)
	token = strings.TrimSuffix(token, "\n")

	tunMac := ""
	nodeIp := ""
	cgwEnabled := "false"
	if gwInfo != "" {
		gwInfos := strings.Split(gwInfo, ",")
		nodeIp = gwInfos[0]
		tunMac = gwInfos[1]
		cgwEnabled = "true"
	}
	spec := map[string][]string{}
	if vip != "" {
		spec["externalIPs"] = []string{vip}
	}
	pd := &HaVipPatchData{
		Metadata: map[string]map[string]string{"annotations": {"tun_mac": tunMac, "node_ip": nodeIp, "cgw_enabled": cgwEnabled}},
		Spec:     spec,
	}
	jsonBytes, err1 := json.Marshal(pd)
	if err1 != nil {
		return fmt.Errorf("Failed to convert patch data to json, since %v", err1)
	}
	req, _ := http.NewRequest("PATCH", url, bytes.NewBuffer(jsonBytes))
	req.Header.Set("Authorization", token)
	req.Header.Set("Content-Type", "application/strategic-merge-patch+json")
	req.Header.Set("Accept", "application/json, */*")

	resp, err2 := client.Do(req)
	if err2 != nil {
		return fmt.Errorf("Request failed, since: %v", err2)
	}
	defer resp.Body.Close()
	return nil
}
