import os


class WorkDirChanger:
    def __init__(self, changed_work_dir):
        """
        进行工作目录的更改
        :param changed_work_dir:
        """
        self.changed_work_dir = changed_work_dir
        self.old_work_dir = os.getcwd()

    def __enter__(self):
        """
        进入上下文之后切换到目标工作目录
        :return:
        """
        os.chdir(self.changed_work_dir)

    def __exit__(self, exc_type, exc_val, exc_tb):
        """
        退出上下文之后切换回原来的工作目录
        :param exc_type:
        :param exc_val:
        :param exc_tb:
        :return:
        """
        os.chdir(self.old_work_dir)
