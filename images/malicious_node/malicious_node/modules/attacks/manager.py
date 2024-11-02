from modules.attacks import icmp_flood_attack as ifam
from modules.attacks import udp_flood_attack as ufam


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

    def start_attack(self):
        """
        开启攻击
        :return:
        """
        if self.attack_type.upper() == "ICMP_FLOOD_ATTACK":
            icmp_flood_attacker = ifam.icmp_flood_attack(attack_thread_count=self.attack_thread_count,
                                                         attacked_node=self.attacked_node,
                                                         attack_duration=self.attack_duration)
            icmp_flood_attacker.start()
        elif self.attack_type.upper() == "UDP_FLOOD_ATTACK":
            udp_flood_attacker = ufam.udp_flood_attack(attack_thread_count=self.attack_thread_count,
                                                       attacked_node=self.attacked_node,
                                                       attack_duration=self.attack_duration)
            udp_flood_attacker.start()
        else:
            raise ValueError("unsupported attacked type")
