#!/bin/bash

function usage(){
    case $1 in
        "")
            echo "$0 bind        bind VIP."
            echo "$0 unbind      unbind VIP."
            echo "$0 setv        set virtual router."
            echo "$0 delv        del virtual router."
            echo "$0 setp        set provider."
            echo "$0 delp        del provider."
            echo "$0 setpd       set provider/default."
            echo "$0 get         get binding or provider(/default) info."
            echo "$0 help CMD    to see more details about sub commands."
            ;;
        "bind")
            echo "$0 bind -v VIP -n NAMESPACE [-s SERVICE] [-e API_ENDPOINT]"
            echo ""
            echo "When API_ENDPOINT is given, instead of printing, API will be executed."
            ;;
        "unbind")
            echo "$0 unbind -v VIP -n NAMESPACE [-s SERVICE] [-e API_ENDPOINT]"
            echo ""
            echo "When API_ENDPOINT is given, instead of printing, API will be executed."
            ;;
        "setp")
            echo "$0 setp -n NODENAME [-i IP] [-r VRID:INTERFACE:VIPS:ADVERT_INTERVAL] [-r ...] [-e API_ENDPOINT]"
            echo ""
            echo "As least IP or virtual_routers should be assigned."
            echo "To only set virtual_router with VIPs for provider, do like -r ::VIPS."
            echo "To unset a property of provider, use \"none\", like -i \"none\", -r ::none."
            echo "When API_ENDPOINT is given, instead of printing, API will be executed."
            ;;
        "delp")
            echo "$0 delp -n NODENAME [-e API_ENDPOINT]"
            echo ""
            echo "When API_ENDPOINT is given, instead of printing, API will be executed."
            ;;
        "setpd")
            echo "$0 setpd [-i INTERFACE] [-a ADVERT_INTERVAL] [-ct CHECK_TIMEOUT] [-ci CHECK_INTERVAL] [-cr CHECK_RISE] [-cf CHECK_FALL] [-e API_ENDPOINT]"
            echo ""
            echo "Beside API_ENDPIONT, as least one parameter should be assigned."
            echo "When API_ENDPOINT is given, instead of printing, API will be executed."
            ;;
        "get")
            echo "$0 get {-v VIP|-n NODENAME|-d|-vips|-onevip} [-e API_ENDPOINT]"
            echo ""
            echo "-d		get provider/default info."
            echo "-vips		get all VIPs."
            echo "-onevip	get one unused VIP."
            echo "When API_ENDPOINT is given, instead of printing, API will be executed."
            ;;
        *)
            echo "Unknown command $1"
    esac
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

function parse_provider_handle_params(){
    _case=$1
    shift
    node=""
    nodeIp=""
    virtualRouters=""
    ep=""
    while (($#)); do
        case $1 in
            "-n")
                node=$2
                shift 2
                ;;
            "-i")
                nodeIp=$2
                shift 2
                ;;
            "-r")
                vrid=`echo $2 | cut -d ':' -f 1`
                intf=`echo $2 | cut -d ':' -f 2`
                vips=`echo $2 | cut -d ':' -f 3`
                intv=`echo $2 | cut -d ':' -f 4`
                newVR="{\"vrid\": \"$vrid\", \"interface\": \"$intf\", \"vips\": \"$vips\", \"advert_interval\": \"$intv\"}"
                if test -z $virtualRouters; then
                    virtualRouters=$newVR
                else
                    virtualRouters="$virtualRouters,$newVR"
                fi
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
    echo "$node;$nodeIp;$virtualRouters;$ep"
}

case $1 in
    "bind"|"unbind")
        _case=$1
        params=`parse_vip_handle_params $@`
        vip=`echo $params | cut -d ';' -f 1`
        ns=`echo $params | cut -d ';' -f 2`
        svc=`echo $params | cut -d ';' -f 3`
        ep=`echo $params | cut -d ';' -f 4`
        if test -z $vip || test -z $ns; then
            usage $_case
        fi
        if [[ $_case == "bind" ]]; then
            withSvc=""
            if test ! -z $svc; then
                withSvc="\"service\": \"$svc\","
            fi
            if test -z $ep; then
                echo "curl -X POST http://\$API_ENDPOINT/vip -d '{\"vip\": \"$vip\", $withSvc\"namespace\": \"$ns\"}'"
            else
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
                curl -X DELETE http://$ep/vip/$vip/${ns}${withSvc}
            fi
        fi
        ;;
    "setv"|"delv")
        ;;
    "setp"|"delp")
        _case=$1
        params=`parse_provider_handle_params $@`
        node=`echo $params | cut -d ';' -f 1`
        nodeIp=`echo $params | cut -d ';' -f 2`
        virtualRouters=`echo $params | cut -d ';' -f 3`
        ep=`echo $params | cut -d ';' -f 4`
        if [[ $_case == "setp" ]]; then
            if test -z "$node" || ( test -z $nodeIp && test -z $virtualRouters ); then
                usage "setp"
            fi
            withVRs=""
            if test ! -z "$virtualRouters"; then
                withVRs="\"virtual_routers\": [$virtualRouters],"
            fi
            if test -z "$ep"; then
                echo "curl -X POST http://\$API_ENDPOINT/node -d '{\"node\": \"$node\", $withVRs \"node_ip\": \"$nodeIp\"}'"
            else
                curl -X POST http://$ep/node -d '{"node": "'$node'", '"$withVRs"'"node_ip": "'$nodeIp'"}'
            fi
        else
            if test -z "$node"; then
                usage "delp"
            fi
            if test -z "$ep"; then
                echo "curl -X DELETE http://\$API_ENDPOINT/node/$node"
            else
                curl -X DELETE http://$ep/node/$node
            fi
        fi
        ;;
    "setpd")
        intf=""
        adIntv=""
        ckIntv=""
        ckTimeout=""
        ckRise=""
        ckFall=""
        while (($#-1)); do
            case $2 in
                "-i")
                    intf=$3
                    shift 2
                    ;;
                "-a")
                    adIntv=$3
                    shift 2
                    ;;
                "-ct")
                    ckTimeout=$3
                    shift 2
                    ;;
                "-ci")
                    ckIntv=$3
                    shift 2
                    ;;
                "-cr")
                    ckRise=$3
                    shift 2
                    ;;
                "-cf")
                    ckFall=$3
                    shift 2
                    ;;
                "-e")
                    ep=$3
                    shift 2
                    ;;
                *)
                    usage "setpd"
                    ;;
            esac
        done
        if [[ "${intf}${adIntv}${ckIntv}${ckTimeout}${ckRise}${ckFall}" == "" ]]; then
            usage "setpd"
        fi
        if test -z $ep; then
            echo "curl -X POST http://\$API_ENDPOINT/provider_default -d '{\"interface\": \"$intf\", \"advert_interval\": \"$adIntv\", \"check_timeout\": \"$ckTimeout\", \"check_interval\": \"$ckIntv\", \"check_rise\": \"$ckRise\", \"check_fall\": \"$ckFall\"}'"
        else
            curl -X POST http://$ep/provider_default -d '{"interface": "'$intf'", "advert_interval": "'$adIntv'", "check_timeout": "'$ckTimeout'", "check_interval": "'$ckIntv'", "check_rise": "'$ckRise'", "check_fall": "'$ckFall'"}'
        fi
        ;;
    "get")
        path=""
        ep=""
        shift
        while (($#)); do
            case $1 in
                "-v")
                    path="vip/$2"
                    shift 2
                    ;;
                "-n")
                    path="node/$2"
                    shift 2
                    ;;
                "-d")
                    path="provider_default"
                    shift 1
                    ;;
                "-vips")
                    path="vips"
                    shift 1
                    ;;
                "-onevip")
                    path="get_one_vip"
                    shift 1
                    ;;
                "-e")
                    ep=$2
                    shift 2
                    ;;
                *)
                    usage "get"
                    ;;
            esac
        done
        if [[ "${path}" == "" ]]; then
            usage "get"
        fi
        if test -z $ep; then
            echo "curl http://\$API_ENDPOINT/$path"
        else
            curl http://$ep/$path
        fi
        ;;
    "help")
        usage $2
        ;;
    *)
        echo "Unknown command $1"
        usage
        ;;
esac

exit 0
