# Copy and customized from:
#   https://github.com/openshift/origin/tree/release-3.9/images/ipfailover/keepalived

# havip pod should config lifecycle.preStop to do service unregister, like:
#   command: ['su', '-c', './keepalived/app/svc_register.py down']
