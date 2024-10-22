import time
from signal_decorator import exit_signal_listener as eslm
from modules.frr import frr_manager as fmm


class Starter:
    def __init__(self):
        pass

    @eslm.signal_decorator
    def never_stop_until_signal(self):
        while True:
            time.sleep(1)

    def main_logic(self):
        """
        主逻辑
            1. 启动 frr
            2. 启动 无限循环
        :return:
        """
        fmm.start_frr()
        self.never_stop_until_signal()


if __name__ == "__main__":
    starter = Starter()
    starter.main_logic()

