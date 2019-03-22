echo "[havip/v2] Using etcd endpoints: $ETCD_ENDPOINTS"
nodes=`echo $ETCD_ENDPOINTS | sed 's/,/ -node /g'`

mv app /etc/keepalived/

with_ca=`test ! -z $ETCD_CA && test -f $ETCD_CA && echo "-client-ca-keys $ETCD_CA" || echo ""`
with_cert=`test ! -z $ETCD_CERT && test -f $ETCD_CERT && echo "-client-cert $ETCD_CERT" || echo ""`
with_key=`test ! -z $ETCD_KEY && test -f $ETCD_KEY && echo "-client-key $ETCD_KEY" || echo ""`

echo "[havip/v2] Running confd as: confd -backend etcd --confdir ${HOME}/etc -node $nodes $with_ca $with_cert $with_key -watch"
confd -backend etcdv3 --confdir ${HOME}/etc -node $nodes $with_ca $with_cert $with_key -watch
