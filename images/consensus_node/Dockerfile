FROM go_env:latest

COPY ./consensus_node /consensus_node

RUN cd /consensus_node && go mod tidy

RUN cd /consensus_node/main && go build

WORKDIR /normal_node/main

# 进行拷贝
COPY resources/daemons /etc/frr/daemons

# 执行函数
CMD ["./main"]