from flask import Flask, request, render_template, make_response, jsonify
import os
import logging
import re
import json

app = Flask(__name__)
app.template_folder = os.path.join(os.path.dirname(__file__), "templates")

# Configure logging
logging.basicConfig(level=logging.INFO, format="%(asctime)s - %(name)s - %(levelname)s - %(message)s")
logger = logging.getLogger(__name__)

@app.before_request
def log_request_info():
    logger.info(f"Request: {request.method} {request.path}\n  - Query: {request.args}\n  - Headers: {request.headers}\n  - Body: {request.get_data()}\n")

class GetMatcherConfig:

    _ROOT_PATH_CFG: dict = {}

    @staticmethod
    def load_config_from_file(file_path):
        try:
            with open(file_path, 'r') as f:
                GetMatcherConfig._ROOT_PATH_CFG = json.load(f)
                logger.info("Configuration loaded successfully.")
        except Exception as e:
            logger.error(f"Failed to load configuration: {e}")

    @staticmethod
    def get_config(path_name):
        return GetMatcherConfig._ROOT_PATH_CFG.get(path_name, None)

# Load the configuration at startup
config_path = os.path.join(os.path.dirname(__file__), "root_path_cfg.json")
GetMatcherConfig.load_config_from_file(config_path)

# Routes generated from mockserver configuration
@app.route('/', methods=['POST'])
def handle_post_requests():
    """Route POST requests to the correct template based on mockserver rules."""
    # Iterate over the mockserver configuration to match the correct response
    for route_name, cfg in GetMatcherConfig._ROOT_PATH_CFG.items():
        # Match headers
        if not re.match(cfg["auth_header_regex"], request.headers.get("Authorization", "")):
            continue
        if not re.match(cfg["amz_target_header_regex"], request.headers.get("X-Amz-Target", "")):
            continue

        # Match body conditions if specified
        body_conditions = cfg.get("body_conditions", {})
        if body_conditions:
            request_body = request.get_json(silent=True) or {}
            for key, regex in body_conditions.items():
                if not re.match(regex, request_body.get(key, "")):
                    break
            else:
                # All conditions matched
                response = make_response(render_template(cfg["template"]))
                response.headers.update(cfg["headers"])
                response.status_code = cfg["status"]
                return response

    return jsonify({'error': 'No matching template found'}), 404

if __name__ == "__main__":
    app.run(debug=True, host="0.0.0.0", port=5000)
