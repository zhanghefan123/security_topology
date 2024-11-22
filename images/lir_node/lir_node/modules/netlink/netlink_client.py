from pyroute2.netlink import NLM_F_REQUEST, genlmsg
from pyroute2.netlink.generic import GenericNetlinkSocket

if __name__ == "__main__":
    import sys
    sys.path.append("../../")


VERSION_NR = 1


class NetlinkMessageType:
    CMD_UNSPEC = 0  # 未指定的命令
    CMD_ECHO = 1  # 测试回显命令

    @classmethod
    def str_to_netlink_message_type(cls, netlink_message_type_str: str):
        if "ECHO" == netlink_message_type_str:
            return cls.CMD_ECHO
        else:
            raise Exception("unknow netlink message type")


class NetlinkMessageFormat(genlmsg):
    nla_map = (
        ('RLINK_ATTR_UNSPEC', 'none'),
        ('RLINK_ATTR_DATA', 'asciiz'),
        ('RLINK_ATTR_LEN', 'uint32'),
    )


class NetlinkClient(GenericNetlinkSocket):
    def __init__(self):
        super().__init__()
        self.bind("EXMPL_GENL", NetlinkMessageFormat)

    def send_netlink_data(self, data: str, message_type: int):
        """
        进行 netlink 数据的发送
        :param data: 数据
        :param message_type: 消息类型
        :return:
        """
        print("---------SEND NETLINK MSG TO KERNEL----------", flush=True)
        message = NetlinkMessageFormat()
        message["cmd"] = message_type
        message["version"] = VERSION_NR
        message["attrs"] = [("RLINK_ATTR_DATA", data)]
        kernel_response = self.nlm_request(message, self.prid, msg_flags=NLM_F_REQUEST)
        print("---------SEND NETLINK MSG TO KERNEL----------", flush=True)
        print("---------RCEIVE KERNEL RESPONSE----------", flush=True)
        response_data = kernel_response[0]
        print(f"RLINK_ATTR_LEN = {response_data.get_attr('RLINK_ATTR_LEN')}", flush=True)
        print(f"RLINK_ATTR_DATA = {response_data.get_attr('RLINK_ATTR_DATA')}", flush=True)
        print("---------RCEIVE KERNEL RESPONSE----------", flush=True)


if __name__ == "__main__":
    netlink_client = NetlinkClient()
    while True:
        netlink_message_type_str_tmp = input("please input input message type: [1. ECHO] [q or quit to exit]:")
        if "q" == netlink_message_type_str_tmp or "quit" == netlink_message_type_str_tmp:
            break
        else:
            netlink_message_type = NetlinkMessageType.str_to_netlink_message_type(netlink_message_type_str_tmp)
            data_tmp = input("please input message you want to send:")
            netlink_client.send_netlink_data(data_tmp, netlink_message_type)
