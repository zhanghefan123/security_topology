FROM go_env:latest

COPY ./ground_station /ground_station

RUN cd /ground_station && go mod tidy

RUN cd /ground_station/main && go build

WORKDIR /ground_station/main

COPY resources/daemons /etc/frr/daemons

CMD ["./main"]