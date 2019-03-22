FROM centos

ENV PATH=$PATH:/usr/local/go/bin

RUN yum install git -y && \
	curl -o go.tar.gz -L https://dl.google.com/go/go1.9.3.linux-amd64.tar.gz && \
	tar -C /usr/local -xzf go.tar.gz && \
	git clone https://github.com/kabukky/httpscerts.git /root/go/src/github.com/kabukky/httpscerts && \
	go install github.com/kabukky/httpscerts

ADD . ./

RUN go build server.go

CMD ["/server"]
