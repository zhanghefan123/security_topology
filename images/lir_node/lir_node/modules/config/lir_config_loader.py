from modules.netlink import netlink_client as ncm


class LirConfigLoader:
    def __init__(self):
        pass

    def start(self):
        netlink = ncm.NetlinkClient()
