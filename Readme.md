# 1. security_topology

- 功能：生成拓扑供共识协议进行测试

# 2. 注意事项

- 依赖：
- [1] go 1.23.0 ｜ go 1.23.1
- [2] sudo apt-get install build-essential

# 3. 各个文件夹的功能
- [1] api
  - [1] chainmaker_api 长安链相应的 api
  - [2] container_api 容器创建/启动/删除相关的 api
  - [3] etcd_api etcd 设置键值对相关的 api
  - [4] linux_tc_api linux tc 设置带宽, 延迟相关的 api
  - [5] multithreading 多线程执行相关的 api
  - [6] route 计算最短路相关的 api
- [2] cmd
  - [1] constellation cobra 星座命令
  - [2] http_service cobra http 服务命令
  - [3] images cobra 镜像处理命令
  - [4] root cobra 主命令
  - [5] test 测试命令 - cobra 测试代码可以放到里面
  - [6] testdata - cobra 长安链相关的测试数据
  - [7] variables - cobra 可供构建的镜像
- [3] configs 配置读取文件夹
- [4] images 存放各个镜像的 Dockerfile 以及相应的依赖
- [5] modules
  -  [1] chainmaker_prepare 长安链 prepare.sh 脚本的 go 语言实现
  -  [2] docker docker 容器客户端创建 api
  -  [3] entities 拓扑之中的各种实体
    - [1] abstract_entities 抽象实体
      - [1] intf 抽象接口
      - [2] link 抽象链路
      - [3] node 抽象节点
    - [2] real_entities 实际实体
      - [1] constellation 星座
      - [2] nodes 各种节点
      - [3] normal_node 各种节点的基础版
      - [4] position_info 卫星位置信息
      - [5] satellites 卫星
      - [6] services 服务
      - [7] topology 拓扑
  -  [4] interface_rate 监控接口速率
  -  [5] logger 日志
  -  [6] utils 相关工具
  -  [7] webshell 前端 shell 创建 api
- [6] resources 存放资源 / 包含配置文件, 长安链依赖
- [7] scripts 脚本文件存放目录，现在仅有删除所有容器和链路的 delete.sh
- [8] services
  - [1] http 代表 http 服务
  - [2] position 代表卫星位置服务
- [9] test 测试目录

# 4. 构建镜像的详细步骤

- [1] 调整 cmd/build.sh 之中的内容, 利用本机的 go 路径进行 build
- [2] 调整 resources/configuration.yml 之中的 chainmaker_go_project_path 以及 crypto_gen_path
- [3] bash build.sh 进行构建
- [4 (较慢)] ./cmd images -i ubuntu_with_software -o build
- [5 (较慢)] ./cmd images -i python_env -o build
- [6] ./cmd images -i go_env -o build
- [7] ./cmd images -i etcd_service -o build
- [8] 将 resources/configuration.yml 之中的 real_time_position_dir 设置为实际的卫星网络项目文件夹的路径
- [9] ./cmd images -i position_service -o build
- [10] ./cmd images -i normal_satellite -o build
- [11] ./cmd images -i router -o build
- [12] ./cmd images -i normal_node -o build
- [13] ./cmd images -i malicious_node -o build
- [14] ./cmd images -i consensus_node -o build

# 5. 启动的详细步骤

- [1] 进入 security_topology/cmd
- [2] 执行 bash build.sh 进行 build 操作
- [3] sudo ./cmd http_service 就可以完成 http 服务的启动, 注意一定要执行 sudo

# 6. submodule 相关内容

- [1] 如果要提交 submodule 之中的内容, git add [到子模块的路径]
- [2] git commit

# 7. 想要添加自己节点的完整步骤

- [1] 首先在 images_config.go 以及 configuration.yml 之中添加镜像名称.
- [2] 在 entities.proto 之中添加自己的类型, 并在其目录执行 protoc --go_out=../types entities.proto
- [3] 然后在 Topology.go 之中添加相应的集合 (以 LirNode 为例), 需要添加 LirNodes, LirAbstractNodes
- [4] 在 create_container.go 之中添加对应类型节点的容器创建过程
- [5] 在 abstract_node.go 之中添加从抽象节点之中提取普通节点的过程
- [6] 在 modules/entities/real_entities/nodes 下添加自己的节点类型
- [7] 在 topology_init.go 之中的 GenerateNodes 和 getSourceNodeAndTargetNode 函数之中添加新类型的相应处理逻辑
- [8] 在 modules/entities/types/utils.go 之中添加自己的节点的前缀