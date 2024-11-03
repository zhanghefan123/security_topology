import multiprocessing
import os
from modules.attacks import icmp_flood_attack as ifam
from modules.attacks import udp_flood_attack as ufam
from modules.config import env_loader as elm
from modules.config.env_loader import env_loader
from modules.utils import address_resolver as arm


class AttackManager:
    def __init__(self, attack_thread_count: int, attack_type: str, attacked_node: str, attack_duration: int):
        """
        初始化攻击管理器
        :param attack_thread_count: 攻击线程数量
        :param attack_type: 攻击类型
        :param attacked_node: 被攻击节点
        :param attack_duration: 攻击时长
        """
        self.attack_thread_count = attack_thread_count
        self.attack_type = attack_type
        self.attacked_node = attacked_node
        self.attack_duration = attack_duration
        self.attack_node_ip: str = ""
        self.attacked_node_ip: str = ""
        self.resolve_attacked_node_address()

    def resolve_attacked_node_address(self):
        address_mapping_file = f"/configuration/{elm.env_loader.container_name}/address_mapping.conf"
        address_mapping = arm.resolve_address_mapping(address_mapping_file=address_mapping_file)
        self.attack_node_ip = address_mapping[env_loader.container_name]
        self.attacked_node_ip = address_mapping[self.attacked_node]

    def start_attack(self):
        """
        开启攻击
        :return:
        """
        if self.attack_type.upper() == "ICMP FLOOD ATTACK":
            process = multiprocessing.Process(target=ifam.icmp_flood_attack, args=(self.attack_thread_count,
                                                                                   self.attack_node_ip,
                                                                                   self.attacked_node_ip,
                                                                                   self.attack_duration))
            process.start()
        elif self.attack_type.upper() == "UDP FLOOD ATTACK":
            process = multiprocessing.Process(target=ufam.udp_flood_attack, args=(self.attack_thread_count, self.attacked_node_ip, self.attack_duration))
            process.start()
        else:
            raise ValueError("unsupported attacked type")
