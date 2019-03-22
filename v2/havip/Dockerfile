FROM centos

RUN curl -o confd -L --connect-timeout 300 --retry 3 https://github.com/kelseyhightower/confd/releases/download/v0.16.0/confd-0.16.0-linux-amd64 && \
    mv confd /usr/bin/ && \
    chmod +x /usr/bin/confd && \
    yum install -y aclocal autoheader automake autoreconf gcc openssl-devel make python-requests python-netaddr python-netifaces && \
    curl -o keepalived.tar.gz https://codeload.github.com/acassen/keepalived/tar.gz/v2.0.10 && \
    tar xf keepalived.tar.gz && \
    cd keepalived-2.0.10 && \
    ./build_setup && ./configure && make && make install && \
    mkdir -p /etc/keepalived

ENV HOME=/home/app

WORKDIR $HOME

COPY . ./

CMD ["/bin/bash", "run.sh"]
