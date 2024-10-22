import os


def start_frr():
    """
    启动 frr
    :return:
    """
    enableFrr = os.getenv("ENABLE_FRR")
    if enableFrr == "true":
        containerName = os.getenv("CONTAINER_NAME")
        sourceFilePath = f"/configuration/{containerName}/route/frr.conf"
        destFilePath = f"/etc/frr/frr.conf"
        copy_command = f"cp {sourceFilePath} {destFilePath}"
        start_frr_command = "service frr start"
        os.system(copy_command)
        os.system(start_frr_command)
        print("start frr", flush=True)
    else:
        print("not start frr", flush=True)
