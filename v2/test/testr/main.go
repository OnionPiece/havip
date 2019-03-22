package main

// Refer: https://gist.github.com/17twenty/2fb30a22141d84e52446
//        https://stackoverflow.com/questions/24455147/how-do-i-send-a-json-string-in-a-post-request-in-go

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"gotest.tools/assert"
)

func SetVirtualRouterDefault(adv_intv, chk_timeout, chk_intv, chk_rise, chk_fall string) (respMsg string) {
	client := &http.Client{}
	data := map[string]string{
		"interface":       "eth0",
		"advert_interval": adv_intv,
		"check_timeout":   chk_timeout,
		"check_interval":  chk_intv,
		"check_rise":      chk_rise,
		"check_fall":      chk_fall,
	}
	jsonBytes, err := json.Marshal(data)
	if err != nil {
		panic(err)
	}
	req, err := http.NewRequest(http.MethodPost, "http://172.28.2.1:8080/virtual_router_default", bytes.NewBuffer(jsonBytes))
	if err != nil {
		log.Println(err)
	}
	resp, err := client.Do(req)
	if err != nil {
		log.Println(err)
	}

	f, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Fatal(err)
	}
	resp.Body.Close()
	respMsg = string(f)
}

func SetVirtualRouter(vrid, adv_intv, intf, start_vip, end_vip string, vips []string) {
}

func main() {
	log.SetOutput(os.Stdout)
	SetVirtualRouterDefault("3", "3", "3", "3", "3")
}
