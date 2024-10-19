package topology

import (
	"fmt"
	"github.com/c-robinson/iplib/v2"
	"zhanghefan123/security_topology/configs"
	"zhanghefan123/security_topology/modules/utils/network"
)

type InitFunction func() error

const (
	GenerateNodes                 = "GenerateNodes"                 // 生成卫星
	GenerateSubnets               = "GenerateSubnets"               // 创建子网
	GenerateLinks                 = "GenerateLinks"                 // 生成链路
	GenerateFrrConfigurationFiles = "GenerateFrrConfigurationFiles" // 生成 frr 配置
	GeneratePeerIdAndPrivateKey   = "GeneratePeerIdAndPrivateKey"   // 生成 peerId 以及私钥
	CalculateSegmentRoutes        = "CalculateSegmentRoutes"        // 进行段路由的计算
)

// Init 进行初始化
func (t *Topology) Init() {
	initSteps := []map[string]InitFunction{
		{GenerateNodes: t.GenerateNodes},
		{GenerateSubnets: t.GenerateSubnets},
		{GenerateLinks: t.GenerateLinks},
		{GenerateFrrConfigurationFiles: t.GenerateFrrConfigurationFiles},
		{GeneratePeerIdAndPrivateKey: t.GeneratePeerIdAndPrivateKey},
		{CalculateSegmentRoutes: t.CalculateSegmentRoutes},
	}
	err := t.initializeSteps(initSteps)
	if err != nil {
		// 所有的错误都添加了完整的上下文信息并在这里进行打印
		topologyLogger.Errorf("constellation init failed: %v", err)
	}
}

// InitializeSteps 按步骤进行初始化
func (t *Topology) initializeSteps(initSteps []map[string]InitFunction) (err error) {
	fmt.Println()
	moduleNum := len(initSteps)
	for idx, initStep := range initSteps {
		for name, initFunc := range initStep {
			if err = initFunc(); err != nil {
				return fmt.Errorf("init step [%s] failed, %s", name, err)
			}
			topologyLogger.Infof("BASE INIT STEP (%d/%d) => init step [%s] success)", idx+1, moduleNum, name)
		}
	}
	fmt.Println()
	return
}

// GenerateNodes 进行节点的生成
func (t *Topology) GenerateNodes() error {
	if _, ok := t.topologyInitSteps[GenerateNodes]; ok {
		topologyLogger.Infof("already generate satellites")
		return nil // 重复生成没有必要，但是实际上只要返回就行，不是错误
	}

	t.topologyInitSteps[GenerateNodes] = struct{}{}
	topologyLogger.Infof("generate satellites")

	return nil
}

func (t *Topology) GenerateSubnets() error {
	if _, ok := t.topologyInitSteps[GenerateSubnets]; ok {
		topologyLogger.Infof("already generate subnets")
		return nil
	}
	var err error
	var ipv4Subnets []iplib.Net4
	var ipv6Subnets []iplib.Net6

	// 进行 ipv4 的子网的生成
	ipv4Subnets, err = network.GenerateIPv4Subnets(configs.TopConfiguration.NetworkConfig.BaseV4NetworkAddress)
	if err != nil {
		return fmt.Errorf("generate subnets: %w", err)
	}
	t.Ipv4SubNets = ipv4Subnets

	// 进行 ipv6 的子网的生成
	ipv6Subnets, err = network.GenerateIpv6Subnets(configs.TopConfiguration.NetworkConfig.BaseV6NetworkAddress)
	if err != nil {
		return fmt.Errorf("generate subnets: %w", err)
	}
	t.Ipv6SubNets = ipv6Subnets

	t.topologyInitSteps[GenerateSubnets] = struct{}{}
	topologyLogger.Infof("generate subnets")
	return nil
}
