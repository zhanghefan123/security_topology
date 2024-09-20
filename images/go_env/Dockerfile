FROM ubuntu_with_software:latest

LABEL mantainer="HeFan Zhang"

ENV GOLANG_VERSION=1.18.10

COPY ./resources/go${GOLANG_VERSION}.linux-amd64.tar.gz /usr/local/

RUN cd /usr/local && \
    tar -xzf go${GOLANG_VERSION}.linux-amd64.tar.gz && \
    rm go${GOLANG_VERSION}.linux-amd64.tar.gz

ENV PATH=$PATH:/usr/local/go/bin
ENV GOPATH=/go
ENV PATH=$PATH:$GOPATH/bin

# 确认是否已经安装成功
RUN go version