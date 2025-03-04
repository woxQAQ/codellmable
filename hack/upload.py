from typing import dataclass_transform
import requests
import json
import os

# 设置参数
dataset_id = "b9379f84-f7b5-49fb-9bf6-715473911204"
api_key = "dataset-CU6K3rTavQrb75206Yb8a8sB"
folder_path = "../output/"  # 替换为实际文件路径
api_base = "localhost"

# 构造请求 URL
url = f"http://{api_base}/v1/datasets/{dataset_id}/document/create-by-file"

# 请求头
headers = {
    "Authorization": f"Bearer {api_key}"
}
# 构造 data 参数的 JSON 结构（所有文件共享相同配置）
data_params = {
    "indexing_technique": "high_quality",
    "process_rule": {
        "rules": {
            "pre_processing_rules": [
                {"id": "remove_extra_spaces", "enabled": True},
                {"id": "remove_urls_emails", "enabled": True}
            ],
            "segmentation": {
                "separator": "###",
                "max_tokens": 500
            }
        },
        "mode": "custom"
    }
}

def upload_file(file_path):
    """上传单个文件的函数"""
    try:
        with open(file_path, 'rb') as file_obj:
            files = {
                # 将 data 参数转换为 JSON 字符串并指定 content-type
                'data': (None, json.dumps(data_params), 'application/json'),
                # 使用原文件名作为参数
                'file': (os.path.basename(file_path), file_obj)
            }
            response = requests.post(url, headers=headers, files=files)
            print(response.json())
            response.raise_for_status()
            return True, response.json()
    except Exception as e:
        return False, str(e)

# 遍历文件夹并上传所有文件
# for filename in os.listdir(folder_path):
#     full_path = os.path.join(folder_path, filename)
#
#     # 跳过子目录（按需可以改为递归遍历）
#     if not os.path.isfile(full_path):
#         continue
#
#     success, result = upload_file(full_path)
for root, dirs, files in os.walk(folder_path):
    for filename in files:
        full_path = os.path.join(root, filename)
        success, result = upload_file(full_path)
        if success:
            print(f"✅ 文件 {filename} 上传成功！响应：{result}")
        else:
            print(f"❌ 文件 {filename} 上传失败：{result}")
            exit(0)

# 可选：如果需要递归上传子目录，可以将上面的遍历改为：

        # ...（同上）
