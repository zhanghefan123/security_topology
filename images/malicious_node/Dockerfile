FROM python_env:latest

COPY ./malicious_node /malicious_node

#COPY malicious_node/connection_exhausted /connection_exhausted

# 进行拷贝
COPY resources/daemons /etc/frr/daemons

COPY resources/requirements.txt /malicious_node/requirements.txt

RUN python -m pip install -r /malicious_node/requirements.txt

RUN cd /malicious_node/modules/attacks/connection_exhausted_attack/build && rm -rf * && cmake .. && make

# 入口点
ENTRYPOINT ["python", "/malicious_node/start.py"]