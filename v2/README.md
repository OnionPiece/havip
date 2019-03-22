## Note

1. Before running on OpenShift via DaemonSet, need create secret with etcd CA, client cert and key firstly, like:

        oc create secret generic havipv2-secret --from-file=etcd-ca=/path/to/master.etcd-ca.crt --from-file=etcd-cert=/path/to/master.etcd-client.crt --from-file=etcd-key=/path/to/master.etcd-client.key

2. Both havip/keepalived and controller needs the following environment variables:

        ETCD_CERT=/path/to/master.etcd-client.crt
        ETCD_KEY=/path/to/master.etcd-client.key
        ETCD_CA=/path/to/master.etcd-ca.crt

    Besides, controller also needs:

        ETCD_ENDPOINTS=https://IP1:2379,https://IP2:2379,https://IP3:2379
