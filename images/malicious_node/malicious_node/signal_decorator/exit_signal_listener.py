import signal


def signal_decorator(func):
    """
    装饰器：作用是代表
    :param func:
    :return:
    """

    def signal_decorated(*args):
        """
        监听到 SIGTERM 和 SIGINT 退出
        :param args:
        :return:
        """
        signal.signal(signal.SIGTERM, lambda signum, frame: exit())
        signal.signal(signal.SIGINT, lambda signum, frame: exit())
        func(*args)

    return signal_decorated
