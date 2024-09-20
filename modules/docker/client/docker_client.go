package client

import (
	dockerClient "github.com/docker/docker/client"
	"github.com/sirupsen/logrus"
	"zhanghefan123/security_topology/modules/logger"
)

var (
	DockerClient       *dockerClient.Client
	moduleDockerLogger *logrus.Entry = logger.GetLogger(logger.ModuleConstellation)
)

// InitDockerClient 初始化 docker 容器
func InitDockerClient() {
	url := "unix:///var/run/docker.sock"
	cli, err := dockerClient.NewClientWithOpts(dockerClient.WithHost(url))
	if err != nil {
		moduleDockerLogger.Errorf("Init Docker Client Error %s", err.Error())
	}
	DockerClient = cli
}
