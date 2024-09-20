FROM ubuntu:20.04
LABEL mantainer="HeFan Zhang"
ENV TZ=Asia/Shanghai
USER root
RUN ln -snf /usr/share/zoneinfo/$TZ /etc/localtime && echo $TZ > /etc/timezone
RUN sed -i "s@http://.*archive.ubuntu.com@http://mirrors.tuna.tsinghua.edu.cn@g" /etc/apt/sources.list && \
    sed -i "s@http://.*security.ubuntu.com@http://mirrors.tuna.tsinghua.edu.cn@g" /etc/apt/sources.list && \
    apt-get update

RUN apt-get install -y vim \
    net-tools \
    tree \
    gcc \
    g++ \
    p7zip-full \
    frr \
    iputils-ping \
    build-essential \
    zlib1g-dev \
    libnetfilter-queue-dev \
    libssl-dev

# 在编译 python 以及安装依赖的时候需要 build-essential / zlib1g-dev / libnetfilter-queue-dev / libssl-dev

# 注意结尾的地方不能加 \

COPY resources/daemons /etc/frr/daemons

RUN echo "net.ipv4.ip_forward=1" >> /etc/sysctl.conf && sysctl -p