package raspberrypi_topology

import (
	"fmt"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"zhanghefan123/security_topology/modules/entities/real_entities/raspberrypi_topology/protobuf"
)

// CreateRaspberrypiConnection 创建一个新的 Raspberry Pi 连接
func CreateRaspberrypiConnection(addr string) (*grpc.ClientConn, error) {
	conn, err := grpc.NewClient(addr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		return nil, fmt.Errorf("failed to create client: %v", err)
	}
	return conn, nil
}

// CreateRaspberrypiClient 创建一个新的 Raspberry Pi 客户端
func CreateRaspberrypiClient(conn *grpc.ClientConn) (protobuf.InteractClient, error) {
	client := protobuf.NewInteractClient(conn)
	if client == nil {
		return nil, fmt.Errorf("failed to create client")
	}
	return client, nil
}
