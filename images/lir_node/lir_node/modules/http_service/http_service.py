from flask import Flask
from gevent.pywsgi import WSGIServer
from flask_cors import *
from modules.config import env_loader as elm

flask_instance = Flask(__name__)  # 创建 flask 实例


def start_flask_http_service():
    """
    进行 http_server 的启动
    """
    CORS(flask_instance, supports_credentials=True)  # CORS 应对措施
    listen_ip_address = "0.0.0.0"
    listen_port = int(elm.env_loader.listen_port)
    http_server = WSGIServer((listen_ip_address, listen_port), flask_instance)
    print("http server start successfully", flush=True)
    http_server.serve_forever()
