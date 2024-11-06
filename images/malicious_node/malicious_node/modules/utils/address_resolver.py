from typing import Dict


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
            container_name, ip_address = line.split("->")
            result_map[container_name] = ip_address
    return result_map
