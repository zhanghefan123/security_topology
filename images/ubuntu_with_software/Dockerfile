FROM ubuntu:20.04
LABEL mantainer="HeFan Zhang"
ENV TZ=Asia/Shanghai
USER root
RUN ln -snf /usr/share/zoneinfo/$TZ /etc/localtime && echo $TZ > /etc/timezone
RUN echo "deb http://mirrors.aliyun.com/ubuntu/ focal main restricted universe multiverse" > /etc/apt/sources.list && \
    echo "deb http://mirrors.aliyun.com/ubuntu/ focal-security main restricted universe multiverse" >> /etc/apt/sources.list && \
    echo "deb http://mirrors.aliyun.com/ubuntu/ focal-updates main restricted universe multiverse" >> /etc/apt/sources.list && \
    echo "deb http://mirrors.aliyun.com/ubuntu/ focal-backports main restricted universe multiverse" >> /etc/apt/sources.list

# 安装依赖应用
RUN apt-get update \
    && apt-get install -y \
    build-essential \
    curl \
    lsb-release \
    zlib1g-dev \
    libncurses5-dev  \
    libgdbm-dev \
    libnss3-dev \
    libssl-dev \
    libreadline-dev \
    libffi-dev \
    wget \
    libnetfilter-queue-dev \
    tcpdump \
    libbz2-dev \
    net-tools \
    iptables \
    vim \
    iputils-ping \
    cmake \
    libsm6 \
    libxext6 \
    libxrender-dev \
    xauth \
    libpcap-dev \
    libssl-dev \
    libsqlite3-dev

# libpcap-dev 和 libssl-dev 是为 libtins 安装的

COPY ./libtins /libtins

# 克隆并安装 libtins
RUN cd /libtins \
    && mkdir build \
    && cd build \
    && cmake ../ -DLIBTINS_ENABLE_CXX11=1 \
    && make \
    && make install


RUN curl -s https://deb.frrouting.org/frr/keys.gpg | tee /usr/share/keyrings/frrouting.gpg > /dev/null
ENV FRRVER="frr-9.1"
RUN echo deb '[signed-by=/usr/share/keyrings/frrouting.gpg]' https://deb.frrouting.org/frr $(lsb_release -s -c) $FRRVER | tee -a /etc/apt/sources.list.d/frr.list
RUN apt update && apt install -y -o Dpkg::Options::="--force-confdef" -o Dpkg::Options::="--force-confold" frr frr-pythontools

