import os


class EnvLoader:
    def __init__(self):
        self.listen_port = os.getenv("LISTEN_PORT")
        self.enable_frr = os.getenv("ENABLE_FRR")
        self.container_name = os.getenv("CONTAINER_NAME")
        self.interface_name = os.getenv("INTERFACE_NAME")


env_loader = EnvLoader()
