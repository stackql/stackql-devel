from flask import Flask, request, render_template, make_response, jsonify
import os
import logging
import re

app = Flask(__name__)
app.template_folder = os.path.join(os.path.dirname(__file__), "templates")

# Configure logging
logging.basicConfig(level=logging.INFO, format="%(asctime)s - %(name)s - %(levelname)s - %(message)s")
logger = logging.getLogger(__name__)

@app.before_request
def log_request_info():
    logger.info(f"Request: {request.method} {request.path}\n  - Query: {request.args}\n  - Headers: {request.headers}\n  - Body: {request.get_data()}\n")




class GetMatcherConfig:

    _ROOT_PATH_CFG: dict = {
        "cloud_resources_list": {
            "template": "cloud_resources_list.jinja.json",
            "status": 200,
            "headers": {
                "Content-Type": "application/json"
            },
            "auth_header_regex": r'^.*ap-southeast-2.*SignedHeaders=accept;content-type;host;x-amz-date;x-amz-target.*$',
            "amz_target_header_regex": r'^CloudApiService\.ListResources$'
        },
        "delete_user": {
            "template": "delete_user.jinja.json",
            "status": 200,
            "headers": {
                "Content-Type": "application/json"
            },
            "auth_header_regex": r'^.*SignedHeaders=content-type;host;x-amz-date;x-amz-target.*$',
            "amz_target_header_regex": r'^TransferService\.DeleteUser$'
        },
        # Additional entries below
        "list_instances": {
            "template": "list_instances.jinja.json",
            "status": 200,
            "headers": {
                "Content-Type": "application/json"
            },
            "auth_header_regex": r'^.*SignedHeaders=content-type;host;x-amz-date;x-amz-target.*$',
            "amz_target_header_regex": r'^ComputeService\.ListInstances$'
        },
        "describe_volumes": {
            "template": "describe_volumes.jinja.json",
            "status": 200,
            "headers": {
                "Content-Type": "application/json"
            },
            "auth_header_regex": r'^.*SignedHeaders=content-type;host;x-amz-date;x-amz-target.*$',
            "amz_target_header_regex": r'^EBSService\.DescribeVolumes$'
        },
        "create_snapshot": {
            "template": "create_snapshot.jinja.json",
            "status": 200,
            "headers": {
                "Content-Type": "application/json"
            },
            "auth_header_regex": r'^.*SignedHeaders=content-type;host;x-amz-date;x-amz-target.*$',
            "amz_target_header_regex": r'^SnapshotService\.CreateSnapshot$'
        }
    }

    @staticmethod
    def get_config(path_name):
        return GetMatcherConfig._ROOT_PATH_CFG.get(path_name, None)

# Routes generated from mockserver configuration
@app.route('/', methods=['POST'])
def handle_post_requests():
    """Route POST requests to the correct template based on mockserver rules."""
    # Iterate over the mockserver configuration to match the correct response

    for route_name, cfg in GetMatcherConfig._ROOT_PATH_CFG.items():
        if re.match(cfg["auth_header_regex"], request.headers.get("Authorization", "")) and re.match(cfg["amz_target_header_regex"], request.headers.get("X-Amz-Target", "")):
            response = make_response(render_template(cfg["template"]))
            response.headers.update(cfg["headers"])
            response.status_code = cfg["status"]
            return response
    return jsonify({'error': 'No matching template found'}), 404
    ## END BLOCK

if __name__ == "__main__":
    app.run(debug=True, host="0.0.0.0", port=5000)
