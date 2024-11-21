import os
import json
from flask import Flask, request
from gevent.pywsgi import WSGIServer
from flask_cors import *
from modules.attacks.manager import AttackManager
from modules.utils import json_response as jrm
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


@flask_instance.route("/startAttack", methods=["POST"])
def start_attack():
    """
    开启攻击
    """
    data = json.loads(request.data)
    print(data, flush=True)
    # {'attack_thread_count': 10,
    # 'attack_type': 'udp flood attack',
    # 'attack_node': 'MaliciousNode-1',
    # 'attacked_node': 'ChainMakerNode-1',
    # 'attack_duration': 1}
    try:
        attack_manager = AttackManager(attack_thread_count=data["attack_thread_count"],
                                       attack_type=data["attack_type"],
                                       attacked_node=data["attacked_node"],
                                       attack_duration=data["attack_duration"])
        attack_manager.start_attack()
        response_data = {
            "status": "success"
        }
        return jrm.get_json_response_from_map(response_data, 200)
    except Exception as e:
        response_data = {
            "status": "error",
            "message": e
        }
        return jrm.get_json_response_from_map(response_data, 500)