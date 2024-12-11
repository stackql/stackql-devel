
import logging
from flask import Flask, render_template, request, jsonify

app = Flask(__name__)

# Configure logging
logging.basicConfig(level=logging.INFO, format="%(asctime)s - %(name)s - %(levelname)s - %(message)s")
logger = logging.getLogger(__name__)

@app.before_request
def log_request_info():
    logger.info(f"Request: {request.method} {request.path} - Query: {request.args}")



@app.route('/v1/projects/testing-project-three/locations/global/keyRings/testing-three/cryptoKeys', methods=['GET'])
def v1_projects_testing_project_three_locations_global_keyRings_testing_three_cryptoKeys():
    response = {
        "cryptoKeys": [
            {
                "name": "projects/testing-project-three/locations/global/keyRings/testing-three/cryptoKeys/demo-key",
                "primary": {
                    "state": "ENABLED",
                    "createTime": "2024-05-22T14:00:00.000000000Z"
                },
                "purpose": "ENCRYPT_DECRYPT"
            }
        ]
    }
    return jsonify(response)

@app.route('/v1/projects/testing-project-three/locations/australia-southeast2/keyRings/big-m-testing-three/cryptoKeys', methods=['GET'])
def v1_projects_testing_project_three_locations_australia_southeast2_keyRings_big_m_testing_three_cryptoKeys():
    return render_template('route_2_template.json'), 200, {'Content-Type': 'application/json'}

@app.route('/v1/projects/testing-project-three/locations/australia-southeast1/keyRings', methods=['GET'])
def v1_projects_testing_project_three_locations_australia_southeast1_keyRings():
    return render_template('route_3_template.json'), 200, {'Content-Type': 'application/json'}

@app.route('/v1/projects/testing-project-three/locations/global/keyRings', methods=['GET'])
def v1_projects_testing_project_three_locations_global_keyRings():
    return render_template('route_4_template.json'), 200, {'Content-Type': 'application/json'}

@app.route('/v1/projects/testing-project-three/locations/australia-southeast2/keyRings', methods=['GET'])
def v1_projects_testing_project_three_locations_australia_southeast2_keyRings():
    return render_template('route_5_template.json'), 200, {'Content-Type': 'application/json'}

@app.route('/v1/projects/testing-project-two/locations/global/keyRings/testing-two/cryptoKeys', methods=['GET'])
def v1_projects_testing_project_two_locations_global_keyRings_testing_two_cryptoKeys():
    return render_template('route_6_template.json'), 200, {'Content-Type': 'application/json'}

@app.route('/v1/projects/testing-project-two/locations/australia-southeast2/keyRings/big-m-testing-two/cryptoKeys', methods=['GET'])
def v1_projects_testing_project_two_locations_australia_southeast2_keyRings_big_m_testing_two_cryptoKeys():
    return render_template('route_7_template.json'), 200, {'Content-Type': 'application/json'}

@app.route('/v1/projects/testing-project-two/locations/australia-southeast1/keyRings', methods=['GET'])
def v1_projects_testing_project_two_locations_australia_southeast1_keyRings():
    return render_template('route_8_template.json'), 200, {'Content-Type': 'application/json'}

@app.route('/v1/projects/testing-project-two/locations/global/keyRings', methods=['GET'])
def v1_projects_testing_project_two_locations_global_keyRings():
    return render_template('route_9_template.json'), 200, {'Content-Type': 'application/json'}

@app.route('/v1/projects/testing-project-two/locations/australia-southeast2/keyRings', methods=['GET'])
def v1_projects_testing_project_two_locations_australia_southeast2_keyRings():
    return render_template('route_10_template.json'), 200, {'Content-Type': 'application/json'}

@app.route('/v1/projects/testing-project/locations/global/keyRings/testing/cryptoKeys', methods=['GET'])
def v1_projects_testing_project_locations_global_keyRings_testing_cryptoKeys():
    return render_template('route_11_template.json'), 200, {'Content-Type': 'application/json'}

@app.route('/v1/projects/testing-project/locations/australia-southeast1/keyRings', methods=['GET'])
def v1_projects_testing_project_locations_australia_southeast1_keyRings():
    return render_template('route_12_template.json'), 200, {'Content-Type': 'application/json'}

@app.route('/v1/projects/testing-project/locations/global/keyRings', methods=['GET'])
def v1_projects_testing_project_locations_global_keyRings():
    return render_template('route_13_template.json'), 200, {'Content-Type': 'application/json'}

@app.route('/projects/testing-project/zones/australia-southeast1-a/instances/000000001/getIamPolicy', methods=['GET'])
def projects_testing_project_zones_australia_southeast1_a_instances_000000001_getIamPolicy():
    return render_template('route_14_template.json'), 200, {'Content-Type': 'application/json'}

@app.route('/projects/testing-project/zones/australia-southeast1-a/machineTypes', methods=['GET'])
def projects_testing_project_zones_australia_southeast1_a_machineTypes():
    return render_template('route_15_template.json'), 200, {'Content-Type': 'application/json'}


## TODO: geet rid once all else stable
@app.route('/token', methods=['POST'])
def token():
    return render_template('route_16_template.json'), 200, {'Content-Type': 'application/json'}

@app.route('/v1/projects/testing-project/aggregated/usableSubnetworks', methods=['GET'])
def v1_projects_testing_project_aggregated_usableSubnetworks():
    return render_template('route_17_template.json'), 200, {'Content-Type': 'application/json'}

@app.route('/v1/projects/another-project/aggregated/usableSubnetworks', methods=['GET'])
def v1_projects_another_project_aggregated_usableSubnetworks():
    return render_template('route_18_template.json'), 200, {'Content-Type': 'application/json'}

@app.route('/v1/projects/yet-another-project/aggregated/usableSubnetworks', methods=['GET'])
def v1_projects_yet_another_project_aggregated_usableSubnetworks():
    return render_template('route_19_template.json'), 200, {'Content-Type': 'application/json'}

@app.route('/v1/projects/empty-project/aggregated/usableSubnetworks', methods=['GET'])
def v1_projects_empty_project_aggregated_usableSubnetworks():
    return render_template('route_20_template.json'), 200, {'Content-Type': 'application/json'}

@app.route('/projects/testing-project/zones/australia-southeast1-a/acceleratorTypes', methods=['GET'])
def projects_testing_project_zones_australia_southeast1_a_acceleratorTypes():
    return render_template('route_21_template.json'), 200, {'Content-Type': 'application/json'}

@app.route('/projects/another-project/zones/australia-southeast1-a/acceleratorTypes', methods=['GET'])
def projects_another_project_zones_australia_southeast1_a_acceleratorTypes():
    return render_template('route_22_template.json'), 200, {'Content-Type': 'application/json'}

@app.route('/v3/projects/testproject:getIamPolicy', methods=['POST'])
def v3_projects_testproject_getIamPolicy():
    return render_template('route_23_template.json'), 200, {'Content-Type': 'application/json'}

@app.route('/v3/organizations/123456789012:getIamPolicy', methods=['GET'])
def v3_organizations_123456789012_getIamPolicy():
    return render_template('route_24_template.json'), 200, {'Content-Type': 'application/json'}

@app.route('/projects/testing-project/zones/australia-southeast1-a/disks', methods=['GET'])
def projects_testing_project_zones_australia_southeast1_a_disks():
    return render_template('route_25_template.json'), 200, {'Content-Type': 'application/json'}

@app.route('/projects/testing-project/zones/australia-southeast1-b/disks', methods=['GET'])
def projects_testing_project_zones_australia_southeast1_b_disks():
    return render_template('route_26_template.json'), 200, {'Content-Type': 'application/json'}

@app.route('/projects/testing-project/global/networks', methods=['GET'])
def projects_testing_project_global_networks():
    return render_template('route_27_template.json'), 200, {'Content-Type': 'application/json'}

@app.route('/projects/testing-project/regions/australia-southeast1/subnetworks', methods=['GET'])
def projects_testing_project_regions_australia_southeast1_subnetworks():
    return render_template('route_28_template.json'), 200, {'Content-Type': 'application/json'}

@app.route('/projects/testing-project/zones/australia-southeast1-a/instances', methods=['GET'])
def projects_testing_project_zones_australia_southeast1_a_instances():
    return render_template('route_29_template.json'), 200, {'Content-Type': 'application/json'}

@app.route('/projects/testing-project/aggregated/instances', methods=['GET'])
def projects_testing_project_aggregated_instances():
    return render_template('route_30_template.json'), 200, {'Content-Type': 'application/json'}

@app.route('/v1/projects/testing-project/assets', methods=['GET'])
def v1_projects_testing_project_assets():
    return render_template('route_31_template.json'), 200, {'Content-Type': 'application/json'}

@app.route('/v1/projects/testing-project/assets', methods=['GET'])
def v1_projects_testing_project_assets_02():   # had to manually rename
    return render_template('route_32_template.json'), 200, {'Content-Type': 'application/json'}

@app.route('/projects/testing-project/global/firewalls/allow-spark-ui', methods=['PUT'])
def projects_testing_project_global_firewalls_allow_spark_ui():
    return render_template('route_33_template.json'), 200, {'Content-Type': 'application/json'}

@app.route('/projects/testing-project/global/firewalls/some-other-firewall', methods=['PATCH'])
def projects_testing_project_global_firewalls_some_other_firewall():
    return render_template('route_34_template.json'), 200, {'Content-Type': 'application/json'}

@app.route('/projects/testing-project/global/firewalls', methods=['GET'])
def projects_testing_project_global_firewalls():
    return render_template('route_35_template.json'), 200, {'Content-Type': 'application/json'}

@app.route('/projects/changing-project/global/firewalls', methods=['GET'])
def projects_changing_project_global_firewalls():
    return render_template('route_36_template.json'), 200, {'Content-Type': 'application/json'}

@app.route('/projects/changing-project/global/firewalls', methods=['GET'])
def projects_changing_project_global_firewalls_02():  # had to manually rename
    return render_template('route_37_template.json'), 200, {'Content-Type': 'application/json'}

if __name__ == '__main__':
    app.run(debug=True)
