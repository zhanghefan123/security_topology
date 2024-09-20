linux: all gen

windows: gen # 需要进行 protobuf 下载 https://github.com/protocolbuffers/protobuf/releases

gen:
	protoc --go_out=../types entities.proto

dev_linux:
	go get github.com/golang/protobuf/protoc-gen-go
	go install github.com/golang/protobuf/protoc-gen-go