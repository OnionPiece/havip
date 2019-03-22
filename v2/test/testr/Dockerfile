FROM centos

ENV PATH=$PATH:/usr/local/go/bin

RUN curl -o go.tar.gz -L https://dl.google.com/go/go1.9.3.linux-amd64.tar.gz && \
	tar -C /usr/local -xzf go.tar.gz && \
	yum install -y git gcc && \
	go get gotest.tools/assert && \
	go get github.com/coreos/etcd

WORKDIR /root

COPY . ./

RUN go build main.go

CMD ["./main"]
