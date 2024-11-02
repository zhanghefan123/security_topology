import os
import json
from flask import Flask, request
from gevent.pywsgi import WSGIServer
from flask_cors import *
from modules.utils import json_response as jrm

flask_instance = Flask(__name__)  # 创建 flask 实例


def start_flask_http_service():
    """
    进行 http_server 的启动
    """
    CORS(flask_instance, supports_credentials=True)  # CORS 应对措施
    listen_ip_address = "0.0.0.0"
    listen_port = int(os.getenv("LISTEN_PORT"))
    http_server = WSGIServer((listen_ip_address, listen_port), flask_instance)
    print("http server start successfully", flush=True)
    http_server.serve_forever()


@flask_instance.route("/startAttack", methods=["POST"])
def start_attack():
    """
    开启攻击
    """
    data = json.loads(request.data)
    print(data, flush=True)
    response_data = {
        "status": "success"
    }
    return jrm.get_json_response_from_map(response_data, 200)
