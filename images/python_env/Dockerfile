FROM ubuntu_with_software:latest
LABEL mantainer="HeFan Zhang"


COPY ./resources/Python-3.8.0.tgz /

RUN tar xvf Python-3.8.0.tgz \
    && cd Python-3.8.0 && ./configure && make -j64 && make install \
    && cd .. \
    && rm -rf Python-3.8.0.tgz \
    && ln -s /usr/local/bin/python3 /usr/bin/python \
    && ln -s /usr/local/bin/pip3 /usr/bin/pip # 将pip3 替换为 pip


# 从源代码安装Python3.6.9
#COPY resources/Python-3.6.9.tgz /
#
#RUN tar xvf Python-3.6.9.tgz \
#    && cd Python-3.6.9 && ./configure && make && make install \
#    && cd .. \
#    && rm -rf Python-3.6.9.tgz \
#    && ln -s /usr/local/bin/python3 /usr/bin/python \
#    && ln -s /usr/local/bin/pip3 /usr/bin/pip
## 配置pypi 清华源
#RUN pip config set global.index-url https://pypi.tuna.tsinghua.edu.cn/simple



#  https://pypi.tuna.tsinghua.edu.cn/simple
# http://pypi.douban.com/simple/
#

RUN pip config set global.index-url https://mirrors.aliyun.com/pypi/simple/

RUN python -m pip install --upgrade pip