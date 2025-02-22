linux: all gen

gen:
	protoc --go_out=../types entities.proto

dev_linux:
	go get github.com/golang/protobuf/protoc-gen-go
	go install github.com/golang/protobuf/protoc-gen-go