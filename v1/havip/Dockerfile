FROM centos

RUN yum install -y kmod keepalived iproute psmisc nmap-ncat net-tools python-netifaces python-requests python-gunicorn python-netaddr

COPY . /var/lib/ipfailover/keepalived/

EXPOSE 1985
WORKDIR /var/lib/ipfailover
CMD ["/var/lib/ipfailover/keepalived/monitor.sh"]
