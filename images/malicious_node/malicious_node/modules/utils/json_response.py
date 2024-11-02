import json


def get_json_response_from_map(response_in_map, code: int):
    """
    将字典类型转换为 json 类型返回
    :param response_in_map: 字典类型变量
    :param code: 响应码
    :return:
    """
    response_json_str = json.dumps(response_in_map)
    headers = {"Content-Type": "application/json"}
    return response_json_str, code, headers
