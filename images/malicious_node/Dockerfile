FROM python_env:latest

COPY ./malicious_node /malicious_node

# 进行拷贝
COPY resources/daemons /etc/frr/daemons

# 入口点
ENTRYPOINT ["python", "/malicious_node/start.py"]