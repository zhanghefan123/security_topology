FROM go_env:latest

COPY ./router /router

RUN cd /router && go mod tidy

RUN cd /router/main && go build

WORKDIR /router/main

# 进行拷贝
COPY resources/daemons /etc/frr/daemons

# 执行入口程序
CMD ["./main"]