import os
from modules.config import env_loader as elm


def start_frr():
    """
    启动 frr
    :return:
    """
    enableFrr = elm.env_loader.enable_frr
    if enableFrr == "true":
        containerName = elm.env_loader.container_name
        sourceFilePath = f"/configuration/{containerName}/route/frr.conf"
        destFilePath = f"/etc/frr/frr.conf"
        copy_command = f"cp {sourceFilePath} {destFilePath}"
        start_frr_command = "service frr start"
        os.system(copy_command)
        os.system(start_frr_command)
        print("start frr", flush=True)
    else:
        print("not start frr", flush=True)
