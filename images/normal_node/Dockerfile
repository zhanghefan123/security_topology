FROM go_env:latest

COPY ./normal_node /normal_node

RUN cd /normal_node && go mod tidy

RUN cd /normal_node/main && go build

WORKDIR /normal_node/main

# 进行拷贝
COPY resources/daemons /etc/frr/daemons

# 执行函数
CMD ["./main"]