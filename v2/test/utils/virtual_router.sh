#!/bin/bash

SET_VR="setvr"
DEL_VR="delvr"

function set_vr_usage(){
    cat > /proc/$$/fd/1 << EOF
$SET_VR -i VRID {-v VIPS|-vs START_VIP -ve END_VIP} [-a ADVERT_INTERVAL] [-f INTERFACE] [-e API_ENDPOINT]

1) VRID works as unique ID to identify each virtual routers.
2) IPs in VIPS should be seperated by comma, like 1.1.1.1,2.2.2.2,3.3.3.3
3) As a alternative way, can pass START_VIP and END_VIP to define an IP range,
   instead of directly pass VIPS.
4) When API_ENDPOINT is given, API will be executed.
EOF
    exit
}

function del_vr_usage(){
    cat > /proc/$$/fd/1 << EOF
$DEL_VR -i VRID [-e API_ENDPOINT]

When API_ENDPOINT is given, API will be executed.
EOF
    exit
}

function parse_virtual_router_handle_params(){
    _case=$1
    shift
    vrid=""
    vips=""
    startVip=""
    endVip=""
    ad_int=""
    interface=""
    ep=""
    while (($#)); do
        case $1 in
            "-i")
                vrid=$2
                shift 2
                ;;
            "-v")
                vips=$2
                shift 2
                ;;
            "-vs")
                startVip=$2
                shift 2
                ;;
            "-ve")
                endVip=$2
                shift 2
                ;;
            "-a")
                ad_int=$2
                shift 2
                ;;
            "-f")
                interface=$2
                shift 2
                ;;
            "-e")
                ep=$2
                shift 2
                ;;
            *)
                vrid=""
                break
                ;;
        esac
    done
    echo "$vrid;$vips;$startVip;$endVip;$ad_int;$interface;$ep"
}

function set_or_del_vr() {
    _case=$1
    params=`parse_virtual_router_handle_params $@`
    vrid=`echo $params | cut -d ';' -f 1`
    vips=`echo $params | cut -d ';' -f 2`
    startVip=`echo $params | cut -d ';' -f 3`
    endVip=`echo $params | cut -d ';' -f 4`
    ad_int=`echo $params | cut -d ';' -f 5`
    interface=`echo $params | cut -d ';' -f 6`
    ep=`echo $params | cut -d ';' -f 7`
    if [[ $vrid == "" ]]; then
        if [[ $_case == $SET_VR ]]; then
            set_vr_usage
        else
            del_vr_usage
        fi
    fi
    if [[ $_case == $SET_VR ]]; then
        if test -z "$vips" && (test -z "$startVip" || test -z "$endVip"); then
            set_vr_usage
        fi
        if test ! -z $vips; then
            vips=`echo $vips | sed 's/,/", "/g'`
            vips='"'$vips'"'
        fi
        if test -z "$ep"; then
            echo "curl -X POST http://\$API_ENDPOINT/virtual_router -d '{\"vrid\": \"$vrid\", \"vips\": [$vips], \"advert_interval\": \"$ad_int\", \"interface\": \"$interface\", \"start_vip\": \"$startVip\", \"end_vip\": \"$endVip\"}'"
        else
            set -x
            curl -X POST http://$ep/virtual_router -d '{"vrid": "'$vrid'", "vips": ['$vips'], "advert_interval": "'$ad_int'", "interface": "'$interface'", "start_vip": "'$startVip'", "end_vip": "'$endVip'"}'
        fi
    else
        if test -z "$ep"; then
            echo "curl -X DELETE http://\$API_ENDPOINT/virtual_router/$vrid"
        else
            set -x
            curl -X DELETE http://$ep/virtual_router/$vrid
        fi
    fi
}
