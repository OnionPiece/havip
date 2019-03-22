#!/bin/bash

ABS_PATH="$(cd `dirname $0`; pwd)"

source $ABS_PATH/vip_binding.sh
source $ABS_PATH/virtual_router.sh
source $ABS_PATH/virtual_router_default.sh
source $ABS_PATH/node.sh
source $ABS_PATH/get.sh

function usage(){
	cat > /proc/$$/fd/1 << EOF
$VIP_BIND          bind VIP to namespace and service.
$VIP_UNBIND        unbind VIP from service or namespace.
$SET_VR         set virtual router, which provide VIPs.
$DEL_VR         del virtual router.
$SET_NODE          set node, which hold virtual routers.
$DEL_NODE          del node.
$SET_VR_DEF        set virtual router default attributes, such as rise and fall for check script.
$GET           get vip binding, virtual router, virtual router default, or node.
help CMD      to see more details about sub commands.
EOF
}

case $1 in
    $VIP_BIND|$VIP_UNBIND)
        bind_or_unbind $@
        ;;
    $SET_VR|$DEL_VR)
        set_or_del_vr $@
        ;;
    $SET_NODE|$DEL_NODE)
        set_or_del_node $@
        ;;
    $SET_VR_DEF)
        set_vr_def $@
        ;;
    $GET)
        get $@
        ;;
    "help")
        case $2 in
            $VIP_BIND)
                bind_usage
                ;;
            $VIP_UNBIND)
                unbind_usage
                ;;
            $SET_VR)
                set_vr_usage
                ;;
            $DEL_VR)
                del_vr_usage
                ;;
            $SET_NODE)
                set_node_usage
                ;;
            $DEL_NODE)
                del_node_usage
                ;;
            $SET_VR_DEF)
                def_usage
                ;;
            $GET)
                get_usage
                ;;
            *)
                usage
                ;;
        esac
        ;;
    *)
        usage
        ;;
esac
