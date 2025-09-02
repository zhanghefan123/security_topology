# zgc version
#/home/zhf/go/go1.23.0/bin/go mod tidy
#/home/zhf/go/go1.23.0/bin/go build

# buaa version / go1.23 configured in path variable
#go mod tidy
#go build

# buaa latest version / go1.23 configured in path variable
#/home/zhf/go/go1.23.1/bin/go mod tidy
#/home/zhf/go/go1.23.1/bin/go build

# 在新的版本之中, 都配置到了环境变量之中
go mod tidy
go build -ldflags="-r /usr/local/lib"
# 加上  -ldflags="-r /usr/local/lib" 的目的是这个库在 fisco-bcos-go-api 调用的时候需要