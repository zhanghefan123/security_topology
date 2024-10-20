package test

import (
	"context"
	"fmt"
	"github.com/coreos/etcd/clientv3"
	"testing"
	"time"
	"zhanghefan123/security_topology/api/container_api"
	"zhanghefan123/security_topology/configs"
	"zhanghefan123/security_topology/modules/docker/client"
	"zhanghefan123/security_topology/modules/entities/abstract_entities/node"
	"zhanghefan123/security_topology/modules/entities/real_entities/services/etcd"
	"zhanghefan123/security_topology/modules/entities/types"
)

func TestEtcd(t *testing.T) {
	// 1. 首先初始化配置
	err := configs.InitLocalConfig()
	if err != nil {
		t.Error(err)
	}
	// 2. 创建 dockerClient
	dockerClient, err := client.NewDockerClient()
	if err != nil {
		t.Error(err)
	}
	// 3. 解析配置
	etcdConfig := configs.TopConfiguration.ServicesConfig.EtcdConfig
	clientPort := etcdConfig.ClientPort
	peerPort := etcdConfig.PeerPort
	dataDir := etcdConfig.DataDir
	etcdName := etcdConfig.EtcdName
	imageName := configs.TopConfiguration.ImagesConfig.EtcdServiceImageName
	// 4. 根据配置创建节点
	actualEtcdNode := etcd.NewEtcdNode(types.NetworkNodeStatus_Logic, clientPort,
		peerPort, dataDir, etcdName, imageName)
	// 5. 创建抽象节点
	etcdNode := node.NewAbstractNode(types.NetworkNodeType_EtcdService, actualEtcdNode)
	// 6. 进行容器的创建和启动
	err = container_api.CreateContainer(dockerClient, etcdNode)
	if err != nil {
		t.Error(err)
	}
	err = container_api.StartContainer(dockerClient, etcdNode)
	if err != nil {
		t.Error(err)
	}
	// 7. 进行 etcd 的服务的调用
	err = CoreFunction()
	if err != nil {
		t.Error(err)
	}

	// 8. 进行容器的停止和关闭
	err = container_api.StopContainer(dockerClient, etcdNode)
	if err != nil {
		t.Error(err)
	}
	err = container_api.RemoveContainer(dockerClient, etcdNode)
	if err != nil {
		t.Error(err)
	}
}

// CoreFunction 进行 etcd 的核心测试
func CoreFunction() error {
	localNetworkAddress := configs.TopConfiguration.NetworkConfig.LocalNetworkAddress
	listenPort := configs.TopConfiguration.ServicesConfig.EtcdConfig.ClientPort
	endPoint := fmt.Sprintf("%s:%d", localNetworkAddress, listenPort)
	cli, err := clientv3.New(clientv3.Config{
		Endpoints:   []string{endPoint},
		DialTimeout: 10 * time.Second,
	})
	kv := clientv3.NewKV(cli)
	_, err = kv.Put(context.Background(), "/hello/first", "hello world 1")
	if err != nil {
		return fmt.Errorf("kv.put error %w", err)
	}
	_, err = kv.Put(context.Background(), "/hello/second", "hello world 2")
	if err != nil {
		return fmt.Errorf("kv.put error %w", err)
	}
	// 按照前缀进行搜索
	getResp, err := kv.Get(context.Background(), "/hello/", clientv3.WithPrefix())
	if err != nil {
		return fmt.Errorf("kv.get error %w", err)
	}
	for _, keyValue := range getResp.Kvs {
		fmt.Println(string(keyValue.Value))
	}
	return nil
}
