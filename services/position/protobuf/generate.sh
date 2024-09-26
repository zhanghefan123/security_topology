#!/bin/bash

# 进入包含 .proto 文件的目录
PROTO_DIR="./"

# 为 link.proto 生成 Go 文件
protoc --proto_path="$PROTO_DIR" --go_out="$PROTO_DIR" "$PROTO_DIR/link/link.proto"

# 为 node.proto 生成 Go 文件
protoc --proto_path="$PROTO_DIR" --go_out="$PROTO_DIR" "$PROTO_DIR/node/node.proto"

echo "Go files generated successfully."
