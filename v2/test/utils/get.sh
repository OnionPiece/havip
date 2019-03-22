#!/bin/bash

GET="get"

function get_usage(){
    cat << EOF
$GET {-v VIP|-r VRID|-n NODENAME|vrd|vips|onevip|nodes|vrs} [-e API_ENDPOINT]

vrd		get virtual router defaults.
vips		get all VIPs.
onevip	        get one unused VIP.
nodes		get all nodes.
vrs		get all VRIDs with VIPs supported by them.

When API_ENDPOINT is given, API will be executed.
EOF
    exit
}

function get(){
    path=""
    ep=""
    shift
    while (($#)); do
        case $1 in
            "-v")
                path="vip/$2"
                shift 2
                ;;
            "-r")
                path="virtual_router/$2"
                shift 2
                ;;
            "-n")
                path="node/$2"
                shift 2
                ;;
            "vrd")
                path="virtual_router_default"
                shift 1
                ;;
            "vips")
                path="vips"
                shift 1
                ;;
            "onevip")
                path="get_one_vip"
                shift 1
                ;;
            "nodes")
                path="nodes"
                shift 1
                ;;
            "vrs")
                path="virtual_routers"
                shift 1
                ;;
            "-e")
                ep=$2
                shift 2
                ;;
            *)
                get_usage
                ;;
        esac
    done
    if [[ "${path}" == "" ]]; then
        get_usage
    fi
    if test -z $ep; then
        echo "curl http://\$API_ENDPOINT/$path"
    else
        curl http://$ep/$path
    fi
}
