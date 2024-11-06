from typing import Dict, Tuple


def resolve_port_mapping(port_mapping_file: str) -> Dict[str, Tuple[int, int]]:
    """
    解析 port_mapping
    :param port_mapping_file 从容器到端口的映射
    :return 端口映射
    """
    result_map = {}
    with open(port_mapping_file, "r") as f:
        lines = f.readlines()
        for line in lines:
            line = line.rstrip("\n")
            container_name, p2pPort, rpcPort = line.split("->")
            result_map[container_name] = (int(p2pPort), int(rpcPort))
    return result_map
