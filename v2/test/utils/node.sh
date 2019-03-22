#!/bin/bash

SET_NODE="setn"
DEL_NODE="deln"

function set_node_usage(){
    cat << EOF
$SET_NODE -n NODENAME [-i IP] -v VRIDS [-e API_ENDPOINT]

1) VRID in VRIDS should be seperated by comma, and each VRID should be in [1, 255],
   and no duplicated.
2) When API_ENDPOINT is given, API will be executed.
EOF
    exit
}

function del_node_usage(){
    cat << EOF
$DEL_NODE -n NODENAME [-e API_ENDPOINT]

When API_ENDPOINT is given, API will be executed.
EOF
    exit
}

function parse_node_handle_params(){
    node=""
    nodeip=""
    vrids=""
    ep=""
    while (($#)); do
        case $1 in
            "-n")
                node=$2
                shift 2
                ;;
            "-i")
                nodeip=$2
                shift 2
                ;;
            "-v")
                vrids=$2
                shift 2
                ;;
            "-e")
                ep=$2
                shift 2
                ;;
            *)
                node=""
                break
                ;;
        esac
    done
    echo "$node;$nodeip;$vrids;$ep"
}

function set_or_del_node(){
    _case=$1
    shift
    params=`parse_node_handle_params $@`
    node=`echo $params | cut -d ';' -f 1`
    nodeip=`echo $params | cut -d ';' -f 2`
    vrids=`echo $params | cut -d ';' -f 3`
    if [[ $vrids != "" ]]; then
        vrids="`echo $vrids | sed 's/,/","/g'`"
        vrids='"'$vrids'"'
    fi
    ep=`echo $params | cut -d ';' -f 4`
    if test -z $node; then
        if [[ $_case == $SET_NODE ]]; then
            set_node_usage
        else
            del_node_usage
        fi
    fi
    if [[ $_case == $SET_NODE ]]; then
        if test -z $ep; then
            vrids=`echo $vrids | sed 's/"/\"/g'`
            echo "curl -X POST http://\$API_ENDPOINT/node -d '{\"node\": \"$node\", \"vrids\": [$vrids], \"node_ip\": \"$nodeip\"}'"
        else
            set -x
            curl -X POST http://$ep/node -d '{"node": "'$node'", "vrids": ['$vrids'], "node_ip": "'$nodeip'"}'
        fi
    else
        if test -z $ep; then
            echo "curl -X DELETE http://\$API_ENDPOINT/node/$node"
        else
            set -x
            curl -X DELETE http://$ep/node/$node
        fi
    fi
}
