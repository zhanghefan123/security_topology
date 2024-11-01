# 1. security_topology

- 功能：生成拓扑供共识协议进行测试

# 2. 注意事项

- 依赖：
- [1] go 1.23.0 ｜ go 1.23.1
- [2] sudo apt-get install build-essential

# 3. 构建镜像的详细步骤

- [1] 调整 cmd/build.sh 之中的内容
- [2] bash build.sh 进行构建
- [3 (较慢)] ./cmd images -i ubuntu_with_software -o build
- [4 (较慢)] ./cmd images -i python_env -o build
- [5] ./cmd images -i go_env -o build
- [6] ./cmd images -i etcd_service -o build
- [7] 将 resources/configuration.yml 之中的 real_time_position_dir 设置为实际的卫星网络项目文件夹的路径
- [8] ./cmd images -i position_service -o build
- [9] ./cmd images -i normal_satellite -o build
- [10] ./cmd images -i router -o build
- [11] ./cmd images -i normal_node -o build
- [12] ./cmd images -i malicious_node -o build
- [13] ./cmd images -i consensus_node -o build
