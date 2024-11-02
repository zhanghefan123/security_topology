import socket
from typing import List, Callable, Any
from string import ascii_letters
from socket import inet_ntop, inet_ntoa, AF_INET6
from sys import maxsize
from struct import pack
from os import urandom
from contextlib import suppress


class Tools:
    @classmethod
    def raw_socket_creator(cls, upper_layer_protocol: int) -> socket.socket:
        created_raw_socket = socket.socket(socket.AF_INET, socket.SOCK_RAW, upper_layer_protocol)
        created_raw_socket.setsockopt(socket.IPPROTO_IP, socket.IP_HDRINCL, 1)
        return created_raw_socket


class RandomApi:
    letters: List[str] = list(ascii_letters)
    rand_str: Callable[[int], str] = lambda length=16: ''.join(
        RandomApi.rand_choice(*RandomApi.letters) for _ in range(length))
    rand_char: Callable[[int], chr] = lambda length=16: "".join(
        [chr(RandomApi.rand_int(0, 1000)) for _ in range(length)])
    rand_ipv4: Callable[[], str] = lambda: inet_ntoa(
        pack('>I', RandomApi.rand_int(1, 0xffffffff)))
    rand_ipv6: Callable[[], str] = lambda: inet_ntop(
        AF_INET6, pack('>QQ', RandomApi.rand_bits(64), RandomApi.rand_bits(64)))
    rand_int: Callable[[int, int], int] = lambda minimum=0, maximum=maxsize: int(
        RandomApi.rand_float(minimum, maximum))
    rand_choice: Callable[[List[Any]], Any] = lambda *data: data[
        (RandomApi.rand_int(maximum=len(data) - 2) or 0)]
    rand: Callable[[], int] = lambda: (int.from_bytes(urandom(7), 'big') >> 3) * (2 ** -53)

    @staticmethod
    def rand_bits(maximum: int = 255) -> int:
        numbytes = (maximum + 7) // 8
        return int.from_bytes(urandom(numbytes),
                              'big') >> (numbytes * 8 - maximum)

    @staticmethod
    def rand_float(minimum: float = 0.0,
                   maximum: float = (maxsize * 1.0)) -> float:
        with suppress(ZeroDivisionError):
            return abs((RandomApi.rand() * maximum) % (minimum -
                                                       (maximum + 1))) + minimum
        return 0.0
