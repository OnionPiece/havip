FROM centos

ENV PATH=$PATH:/usr/local/go/bin

RUN curl -o go.tar.gz -L https://dl.google.com/go/go1.9.3.linux-amd64.tar.gz && \
	tar -C /usr/local -xzf go.tar.gz && \
	yum install -y git gcc && \
	mkdir -p /root/go/src/github.com/coreos/etcd && \
	mkdir -p /root/go/src/go.etcd.io/etcd && \
	git clone https://github.com/coreos/etcd.git /root/go/src/github.com/coreos/etcd && \
	git clone https://github.com/etcd-io/etcd.git /root/go/src/go.etcd.io/etcd && \
	go install github.com/coreos/etcd

COPY . /root/

WORKDIR /root

RUN mkdir -p /root/go/src/utils && \
	mv /root/utils.go /root/go/src/utils && \
	mkdir -p /root/go/src/model && \
	mv /root/model.go /root/go/src/model && \
	mv /root/model_test.go /root/go/src/model && \
	go build apiserver.go && \
	go build watchd.go

CMD ["/bin/bash", "/root/run.sh"]
