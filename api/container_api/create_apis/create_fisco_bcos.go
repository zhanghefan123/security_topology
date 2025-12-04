package create_apis

import (
	"context"
	"fmt"
	"github.com/docker/docker/api/types/container"
	docker "github.com/docker/docker/client"
	"github.com/docker/go-connections/nat"
	"path/filepath"
	"zhanghefan123/security_topology/configs"
	"zhanghefan123/security_topology/modules/entities/real_entities/nodes"
	"zhanghefan123/security_topology/modules/entities/types"
)

// 构建的 bash 脚本
/*
#!/bin/bash
SHELL_FOLDER=$(cd $(dirname $0);pwd)

LOG_ERROR() {
    content=${1}
    echo -e "\033[31m[ERROR] ${content}\033[0m"
}

LOG_INFO() {
    content=${1}
    echo -e "\033[32m[INFO] ${content}\033[0m"
}

fisco_bcos=${SHELL_FOLDER}/../fisco-bcos
export RUST_LOG=bcos_wasm=error
cd ${SHELL_FOLDER}
node=$(basename ${SHELL_FOLDER})
node_pid=$(docker ps |grep ${SHELL_FOLDER//\//} | grep -v grep|awk '{print $1}')
ulimit -n 1024
#start monitor
dirs=($(ls -l ${SHELL_FOLDER} | awk '/^d/ {print $NF}'))
for dir in ${dirs[*]}
do
    if [[ -f "${SHELL_FOLDER}/${dir}/node.mtail" && -f "${SHELL_FOLDER}/${dir}/start_mtail_monitor.sh" ]];then
        echo "try to start ${dir}"
        bash ${SHELL_FOLDER}/${dir}/start_mtail_monitor.sh &
    fi
done


if [ ! -z ${node_pid} ];then
    kill -USR1 ${node_pid}
    sleep 0.2
    kill -USR2 ${node_pid}
    sleep 0.2
    echo " ${node} is running, container id is $node_pid."
    exit 0
else
*/
// 最主要的命令
// docker run -d --rm --name ${SHELL_FOLDER//\//} -v ${SHELL_FOLDER}:/data --network=host -w=/data fiscoorg/fiscobcos:v3.12.1 -c config.ini -g config.genesis
/*
    sleep 1.5
fi
try_times=4
i=0
while [ $i -lt ${try_times} ]
do
    node_pid=$(docker ps |grep ${SHELL_FOLDER//\//} | grep -v grep|awk '{print $1}')
    success_flag=success
    if [[ ! -z ${node_pid} && ! -z "${success_flag}" ]];then
        echo -e "\033[32m ${node} start successfully pid=${node_pid}\033[0m"
        exit 0
    fi
    sleep 0.5
    ((i=i+1))
done
echo -e "\033[31m  Exceed waiting time. Please try again to start ${node} \033[0m"
tail -n20 $(docker inspect --format='{{.LogPath}}' ${SHELL_FOLDER//\//})

*/

// CreateFiscoBcosNode 创建 FiscoBcosNode 容器的代码
func CreateFiscoBcosNode(client *docker.Client, fiscoBcosNode *nodes.FiscoBcosNode, graphNodeId int) error {
	if fiscoBcosNode.Status != types.NetworkNodeStatus_Logic {
		return fmt.Errorf("fisco bcos node not in logic status cannot create")
	}

	// 1 获取第一个接口
	firstInterfaceName := fiscoBcosNode.Interfaces[0].IfName
	firstInterfaceAddress := fiscoBcosNode.Interfaces[0].SourceIpv4Addr[:len(fiscoBcosNode.Interfaces[0].SourceIpv4Addr)-3]

	// 2 创建 sysctls
	sysctls := map[string]string{
		// ipv4 的相关网络配置
		"net.ipv4.ip_forward":          "1",
		"net.ipv4.conf.all.forwarding": "1",

		// ipv6 的相关网络配置
		"net.ipv6.conf.default.disable_ipv6":     "0",
		"net.ipv6.conf.all.disable_ipv6":         "0",
		"net.ipv6.conf.all.forwarding":           "1",
		"net.ipv6.conf.default.seg6_enabled":     "1",
		"net.ipv6.conf.eth0.seg6_enabled":        "1",
		"net.ipv6.conf.lo.seg6_enabled":          "1",
		"net.ipv6.conf.all.seg6_enabled":         "1",
		"net.ipv6.conf.all.keep_addr_on_down":    "1",
		"net.ipv6.route.skip_notify_on_dev_down": "1",
		"net.ipv4.conf.all.rp_filter":            "0",
		"net.ipv6.seg6_flowlabel":                "1",
	}

	// 3. 创建容器卷映射
	examplePath := configs.TopConfiguration.FiscoBcosConfig.ExamplePath
	simulationDir := configs.TopConfiguration.PathConfig.ConfigGeneratePath
	selfDefinedNodeDir := filepath.Join(simulationDir, fiscoBcosNode.ContainerName)
	originalNodePath := filepath.Join(examplePath, fmt.Sprintf("nodes/127.0.0.1/node%d/", fiscoBcosNode.Id-1))

	volumes := []string{
		fmt.Sprintf("%s:%s", selfDefinedNodeDir, fmt.Sprintf("/configuration/%s", fiscoBcosNode.ContainerName)),
		fmt.Sprintf("%s:%s", originalNodePath, fmt.Sprintf("/data")),
	}

	// 4. 配置环境变量
	enableFrr := configs.TopConfiguration.NetworkConfig.EnableFrr
	webPort := configs.TopConfiguration.ServicesConfig.WebConfig.StartPort + graphNodeId
	leaderPeriod := configs.TopConfiguration.FiscoBcosConfig.LeaderPeriod
	consensusTimeoutMs := configs.TopConfiguration.FiscoBcosConfig.ConsensusTimeout
	sealerWaitMs := configs.TopConfiguration.FiscoBcosConfig.SealWatiMs
	allowedLaggingBehindBlocks := configs.TopConfiguration.FiscoBcosConfig.AllowedLaggingBehindBlocks
	waterMarkLimit := configs.TopConfiguration.FiscoBcosConfig.WaterMarkLimit
	minSealTime := configs.TopConfiguration.FiscoBcosConfig.MinSealTime
	recursiveTrigger := configs.TopConfiguration.FiscoBcosConfig.RecursiveTrigger
	fastTrigger := configs.TopConfiguration.FiscoBcosConfig.FastTrigger
	largeInterval := configs.TopConfiguration.FiscoBcosConfig.LargeInterval
	useModifiedRequestBlocks := configs.TopConfiguration.FiscoBcosConfig.UseModifiedRequestBlocks
	enablePending := configs.TopConfiguration.FiscoBcosConfig.EnablePending
	envs := []string{
		// zhf 添加的环境变量
		fmt.Sprintf("%s=%s", "FIRST_INTERFACE_NAME", firstInterfaceName),
		fmt.Sprintf("%s=%s", "FIRST_INTERFACE_ADDR", firstInterfaceAddress),
		fmt.Sprintf("%s=%d", "NODE_ID", fiscoBcosNode.Id),
		fmt.Sprintf("%s=%s", "CONTAINER_NAME", fiscoBcosNode.ContainerName),
		fmt.Sprintf("%s=%t", "ENABLE_FRR", enableFrr),
		fmt.Sprintf("%s=%d", "WEB_SERVER_LISTEN_PORT", webPort),
		fmt.Sprintf("%s=%d", "LEADER_PERIOD", leaderPeriod),
		fmt.Sprintf("%s=%d", "CONSENSUS_TIMEOUT", consensusTimeoutMs),
		fmt.Sprintf("%s=%d", "SEALER_WAIT_MS", sealerWaitMs),
		fmt.Sprintf("%s=%d", "MIN_SEAL_TIME", minSealTime),
		fmt.Sprintf("%s=%d", "ALLOWED_LAGGING_BEHIND_BLOCKS", allowedLaggingBehindBlocks),
		fmt.Sprintf("%s=%d", "WATER_MARK_LIMIT", waterMarkLimit),
		fmt.Sprintf("%s=%t", "RECURSIVE_TRIGGER", recursiveTrigger),
		fmt.Sprintf("%s=%t", "FAST_TRIGGER", fastTrigger),
		fmt.Sprintf("%s=%t", "LARGE_INTERVAL", largeInterval),
		fmt.Sprintf("%s=%t", "USE_MODIFIED_REQUEST_BLOCKS", useModifiedRequestBlocks),
		fmt.Sprintf("%s=%t", "ENABLE_PENDING", enablePending),
		fmt.Sprintf("%s=%s", "ETCD_LISTEN_ADDR", configs.TopConfiguration.NetworkConfig.LocalNetworkAddress),
		fmt.Sprintf("%s=%d", "ETCD_LISTEN_PORT", configs.TopConfiguration.ServicesConfig.EtcdConfig.ClientPort),
		fmt.Sprintf("%s=%d", "SYNC_IDLE_WAIT", configs.TopConfiguration.FiscoBcosConfig.SyncIdleWait),
		fmt.Sprintf("%s=%d", "DOWNLOAD_BLOCK_PROCESSOR_THREAD_COUNT", configs.TopConfiguration.FiscoBcosConfig.DownloadBlockProcessorThreadCount),
		fmt.Sprintf("%s=%d", "SEND_BLOCK_PROCESSOR_THREAD_COUNT", configs.TopConfiguration.FiscoBcosConfig.SendBlockProcessorThreadCount),
		fmt.Sprintf("%s=%d", "MAX_SHARD_PER_PEER", configs.TopConfiguration.FiscoBcosConfig.MaxShardPerPeer),
		fmt.Sprintf("%s=%d", "MAX_BLOCKS_PER_REQUEST", configs.TopConfiguration.FiscoBcosConfig.MaxBlocksPerRequest),
		fmt.Sprintf("%s=%d", "REQUEST_PEER_LIMIT", configs.TopConfiguration.FiscoBcosConfig.RequestPeerLimit),
		fmt.Sprintf("%s=%d", "EXPECTED_TTL", configs.TopConfiguration.FiscoBcosConfig.ExpectedTTL),
		fmt.Sprintf("%s=%d", "SYNC_SLEEP_MS", configs.TopConfiguration.FiscoBcosConfig.SyncSleepMs),
		fmt.Sprintf("%s=%t", "ENABLE_BLACK_LIST", configs.TopConfiguration.FiscoBcosConfig.EnableBlackList),
		fmt.Sprintf("%s=%d", "BLOCK_INTERVAL_MS", configs.TopConfiguration.FiscoBcosConfig.BlockIntervalMs),
		fmt.Sprintf("%s=%d", "NEW_VIEW_WAIT_MS", configs.TopConfiguration.FiscoBcosConfig.NewViewWaitMs),
	}

	// 5. 资源限制
	//cpuLimit := 5
	//memoryLimit := 2 * 1024
	//resourcesLimit := container.Resources{
	//	NanoCPUs: int64(cpuLimit * 1e9),
	//	Memory:   int64(memoryLimit * 1024 * 1024),
	//}

	// 6. 端口映射 (现在暂时没有端口映射)
	rpcStartPort := configs.TopConfiguration.FiscoBcosConfig.RpcStartPort
	fmt.Printf(fmt.Sprintf("%d/tcp\n", rpcStartPort+fiscoBcosNode.Id-1))
	rpcPort := nat.Port(fmt.Sprintf("%d/tcp", rpcStartPort+fiscoBcosNode.Id-1))

	exposedPorts := nat.PortSet{
		rpcPort: {},
	}

	portBindings := nat.PortMap{
		rpcPort: []nat.PortBinding{
			{
				HostIP:   "0.0.0.0",
				HostPort: string(rpcPort),
			},
		},
	}

	// 7. 创建容器配置
	//错误的 path -> 认为是宿主机路径下的配置文件
	/*
		    pathToConfigIni := filepath.Join(examplePath, fmt.Sprintf("nodes/127.0.0.1/node%d/config.ini", fiscoBcosNode.Id-1))
			pathToGenesis := filepath.Join(examplePath, fmt.Sprintf("nodes/127.0.0.1/node%d/config.genesis", fiscoBcosNode.Id-1))
	*/
	containerConfig := &container.Config{
		Image:        configs.TopConfiguration.ImagesConfig.FiscoBcosImageName,
		Tty:          true,
		Env:          envs,
		ExposedPorts: exposedPorts,
		WorkingDir:   "/data",                                                                           // 对应于 -w=/data
		Cmd:          []string{"/usr/local/bin/fisco-bcos", "-c", "config.ini", "-g", "config.genesis"}, // 这两个配置文件的路径都是容器内的
	}

	// 8. hostConfig
	var hostConfig *container.HostConfig
	hostConfig = &container.HostConfig{
		// 容器数据卷映射
		Binds:        volumes,
		CapAdd:       []string{"NET_ADMIN"},
		Privileged:   true,
		Sysctls:      sysctls,
		PortBindings: portBindings,
		//NetworkMode:  "host", // 对应于 --network=host
	}
	fmt.Println("no resource limit")

	// 9. 进行容器的创建
	response, err := client.ContainerCreate(
		context.Background(),
		containerConfig,
		hostConfig,
		nil,
		nil,
		fiscoBcosNode.ContainerName,
	)
	if err != nil {
		return fmt.Errorf("create fisco bcos failed %v", err)
	}

	// 10. 设置创建后分配的 id
	fiscoBcosNode.ContainerId = response.ID

	// 11. 状态转换
	fiscoBcosNode.Status = types.NetworkNodeStatus_Created

	return nil
}
