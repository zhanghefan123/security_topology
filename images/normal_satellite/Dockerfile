# the first stage
FROM go_env:latest

COPY ./normal_satellite /normal_satellite

RUN cd /normal_satellite && go mod tidy

RUN cd /normal_satellite/main && go build

WORKDIR /normal_satellite/main

# 进行拷贝
COPY resources/daemons /etc/frr/daemons

CMD ["./main"]