#!/bin/bash

SET_VR_DEF="setvrd"

function def_usage(){
    cat << EOF
$SET_VR_DEF [-i INTERFACE] [-a ADVERT_INTERVAL] [-ct CHECK_TIMEOUT] [-ci CHECK_INTERVAL] [-cr CHECK_RISE] [-cf CHECK_FALL] [-e API_ENDPOINT]

1) Beside API_ENDPIONT, as least one parameter should be assigned.
2) INTERFACE and ADVERT_INTERVAL will be used as default interface and
   advert_interval, for virtual router doesn't set them.
3) When API_ENDPOINT is given, API will be executed.
EOF
    exit
}

function set_vr_def(){
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
                def_usage
                break
                ;;
        esac
    done
    if [[ "${intf}${adIntv}${ckIntv}${ckTimeout}${ckRise}${ckFall}" == "" ]]; then
        def_usage
    fi
    if test -z $ep; then
        echo "curl -X POST http://\$API_ENDPOINT/virtual_router_default -d '{\"interface\": \"$intf\", \"advert_interval\": \"$adIntv\", \"check_timeout\": \"$ckTimeout\", \"check_interval\": \"$ckIntv\", \"check_rise\": \"$ckRise\", \"check_fall\": \"$ckFall\"}'"
    else
        set -x
        curl -X POST http://$ep/virtual_router_default -d '{"interface": "'$intf'", "advert_interval": "'$adIntv'", "check_timeout": "'$ckTimeout'", "check_interval": "'$ckIntv'", "check_rise": "'$ckRise'", "check_fall": "'$ckFall'"}'
    fi
}
