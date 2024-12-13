from flask import Flask, request, render_template, make_response, jsonify
import os
import logging
import re
import json
import base64

app = Flask(__name__)
app.template_folder = os.path.join(os.path.dirname(__file__), "templates")

# Configure logging
logging.basicConfig(level=logging.INFO, format="%(asctime)s - %(name)s - %(levelname)s - %(message)s")
logger = logging.getLogger(__name__)

@app.before_request
def log_request_info():
    logger.info(f"Request: {request.method} {request.path}\n  - Query: {request.args}\n  - Headers: {request.headers}\n  - Body: {request.get_data()}")

class GetMatcherConfig:

    _ROOT_PATH_CFG: dict = {}

    @staticmethod
    def load_config_from_file(file_path):
        try:
            with open(file_path, 'r') as f:
                GetMatcherConfig._ROOT_PATH_CFG = json.load(f)

                # Decode base64 responses in templates
                for route_name, cfg in GetMatcherConfig._ROOT_PATH_CFG.items():
                    if "base64_template" in cfg:
                        try:
                            decoded_content = base64.b64decode(cfg["base64_template"]).decode("utf-8")
                            template_path = os.path.join(app.template_folder, cfg["template"])
                            with open(template_path, "w") as tpl_file:
                                tpl_file.write(decoded_content)
                            logger.info(f"Decoded base64 template for route: {route_name}")
                        except Exception as e:
                            logger.error(f"Failed to decode base64 template for route: {route_name}: {e}")

                logger.info("Configuration loaded and templates processed successfully.")
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
    for route_name, cfg in GetMatcherConfig._ROOT_PATH_CFG.items():
        logger.debug(f"Evaluating route: {route_name}")

        # Match headers
        auth_regex = cfg.get("auth_header_regex", "") or ""
        if not isinstance(auth_regex, str):
            logger.error(f"Invalid auth_header_regex for route {route_name}: {auth_regex}")
            continue

        if not re.match(auth_regex, request.headers.get("Authorization", "")):
            logger.debug(f"Header mismatch: auth_regex='{auth_regex}' did not match '{request.headers.get('Authorization', '')}'")
            continue

        amz_target_regex = cfg.get("amz_target_header_regex", "") or ""
        if not isinstance(amz_target_regex, str):
            logger.error(f"Invalid amz_target_header_regex for route {route_name}: {amz_target_regex}")
            continue

        if not re.match(amz_target_regex, request.headers.get("X-Amz-Target", "")):
            logger.debug(f"Header mismatch: amz_target_regex='{amz_target_regex}' did not match '{request.headers.get('X-Amz-Target', '')}'")
            continue

        # Match body conditions if specified
        body_conditions = cfg.get("body_conditions", {})
        if body_conditions:
            request_body = request.get_json(silent=True) or {}
            for key, regex in body_conditions.items():
                if not isinstance(regex, str) or not regex:
                    logger.warning(f"Skipping invalid regex for body condition key='{key}': {regex}")
                    continue

                if not re.match(regex, request_body.get(key, "")):
                    logger.debug(f"Body mismatch: key='{key}', value='{request_body.get(key, '')}', regex='{regex}'")
                    break
            else:
                # All conditions matched
                logger.info(f"Match found for route: {route_name}")
                if "template" not in cfg:
                    logger.error(f"Missing template for route: {route_name}")
                    return jsonify({'error': f'Missing template for route: {route_name}'}), 500
                response = make_response(render_template(cfg["template"]))
                response.headers.update(cfg.get("headers", {}))
                response.status_code = cfg.get("status", 200)
                return response

    logger.error("No matching configuration found for the request.")
    return jsonify({'error': 'No matching template found'}), 404

@app.route('/', methods=['GET'])
def handle_get_requests():
    """Handle GET requests if required by mockserver rules."""
    logger.info("GET request received but no rules defined.")
    return jsonify({'error': 'GET requests not supported in this configuration'}), 405

if __name__ == "__main__":
    app.run(debug=True, host="0.0.0.0", port=5000)
