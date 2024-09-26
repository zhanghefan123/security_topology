package client

import (
	"fmt"
	dockerClient "github.com/docker/docker/client"
)

// NewDockerClient 创建新的 dockerClient
func NewDockerClient() (*dockerClient.Client, error) {
	url := "unix:///var/run/docker.sock"
	cli, err := dockerClient.NewClientWithOpts(dockerClient.WithHost(url))
	if err != nil {
		return nil, fmt.Errorf("new Docker Client Error %w", err)
	}
	return cli, nil
}
