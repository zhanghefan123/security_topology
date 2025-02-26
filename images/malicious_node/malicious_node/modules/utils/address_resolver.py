from typing import Dict


# 下面是 address_mapping.conf 的一个示意内容
# -----------------------------------------
# Router-4->4->192.168.0.6
# Router-7->7->192.168.0.10
# Router-3->3->192.168.0.2
# ChainMakerNode-2->9->192.168.0.21
# MaliciousNode-1->14->192.168.0.50
# Router-6->6->192.168.0.14
# ChainMakerNode-1->8->192.168.0.1
# ChainMakerNode-3->10->192.168.0.5
# ChainMakerNode-4->11->192.168.0.9
# Router-1->1->192.168.0.22
# Router-2->2->192.168.0.38
# Router-5->5->192.168.0.18
# ChainMakerNode-5->12->192.168.0.13
# ChainMakerNode-6->13->192.168.0.17
# -----------------------------------------

def resolve_address_mapping(address_mapping_file: str) -> Dict[str, str]:
    """
    解析 address_mapping
    :param address_mapping_file: 地址映射文件的路径
    :return: 地址映射
    """
    result_map = {}
    with open(address_mapping_file, "r") as f:
        lines = f.readlines()
        for line in lines:
            line = line.rstrip("\n")
            container_name, graph_id, ip_address = line.split("->")
            result_map[container_name] = ip_address
    return result_map
