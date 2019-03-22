#!/bin/bash

VIP_BIND="bind"
VIP_UNBIND="unbind"

function bind_usage(){
    cat << EOF
$VIP_BIND -v VIP -n NAMESPACE [-s SERVICE] [-e API_ENDPOINT]

When API_ENDPOINT is given, API will be executed.
EOF
    exit
}

function unbind_usage(){
    cat << EOF
$VIP_UNBIND -v VIP -n NAMESPACE [-s SERVICE] [-e API_ENDPOINT]

When API_ENDPOINT is given, API will be executed.
EOF
    exit
}

function parse_vip_handle_params(){
    _case=$1
    shift
    vip=""
    ns=""
    svc=""
    ep=""
    while (($#)); do
        case $1 in
            "-v")
                vip=$2
                shift 2
                ;;
            "-n")
                ns=$2
                shift 2
                ;;
            "-s")
                svc=$2
                shift 2
                ;;
            "-e")
                ep=$2
                shift 2
                ;;
            *)
                vip=""
                break
                ;;
        esac
    done
    echo "$vip;$ns;$svc;$ep"
}

function bind_or_unbind() {
    _case=$1
    params=`parse_vip_handle_params $@`
    vip=`echo $params | cut -d ';' -f 1`
    ns=`echo $params | cut -d ';' -f 2`
    svc=`echo $params | cut -d ';' -f 3`
    ep=`echo $params | cut -d ';' -f 4`
    if test -z $vip || test -z $ns; then
        if [[ $_case == $VIP_BIND ]]; then
            bind_usage
        else
            unbind_usage
        fi
    fi
    if [[ $_case == $VIP_BIND ]]; then
        withSvc=""
        if test ! -z $svc; then
            withSvc="\"service\": \"$svc\","
        fi
        if test -z $ep; then
            echo "curl -X POST http://\$API_ENDPOINT/vip -d '{\"vip\": \"$vip\", $withSvc\"namespace\": \"$ns\"}'"
        else
            set -x
            curl -X POST http://$ep/vip -d '{"vip": "'$vip'", '"$withSvc"' "namespace": "'$ns'"}'
        fi
    else
        withSvc=""
        if test ! -z $svc; then
            withSvc="/$svc"
        fi
        if test -z $ep; then
            echo "curl -X DELETE http://\$API_ENDPOINT/vip/$vip/${ns}${withSvc}"
        else
            set -x
            curl -X DELETE http://$ep/vip/$vip/${ns}${withSvc}
        fi
    fi
}
