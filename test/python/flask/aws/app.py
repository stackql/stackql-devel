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

# Routes generated from mockserver configuration
@app.route('/', methods=['POST'])
def handle_post_requests():
    """Route POST requests to the correct template based on mockserver rules."""
    # Iterate over the mockserver configuration to match the correct response

    if re.match(r'^.*ap-southeast-2.*SignedHeaders=accept;content-type;host;x-amz-date;x-amz-target.*$', request.headers.get('Authorization', '')) and request.headers.get('X-Amz-Target', '').startswith('CloudApiService.ListResources'):
        return render_template('template_1.json')
    if re.match(r'^.*ap-southeast-2.*SignedHeaders=.*content-type;host;x-amz-date;x-amz-target.*$', request.headers.get('Authorization', '')) and request.headers.get('X-Amz-Target', '').startswith('CloudApiService.GetResource'):
        return render_template('template_2.json')
    if re.match(r'^.*ap-southeast-2.*SignedHeaders=.*content-type;host;x-amz-date;x-amz-target.*$', request.headers.get('Authorization', '')) and request.headers.get('X-Amz-Target', '').startswith('CloudApiService.GetResource'):
        return render_template('template_3.json')
    if re.match(r'^.*ap-southeast-2.*SignedHeaders=.*content-type;host;x-amz-date;x-amz-target.*$', request.headers.get('Authorization', '')) and request.headers.get('X-Amz-Target', '').startswith('CloudApiService.GetResource'):
        return render_template('template_4.json')
    if re.match(r'^.*ap-southeast-2.*SignedHeaders=.*content-type;host;x-amz-date;x-amz-target.*$', request.headers.get('Authorization', '')) and request.headers.get('X-Amz-Target', '').startswith('CloudApiService.GetResource'):
        return render_template('template_5.json')
    if re.match(r'^.*ap-southeast-2.*SignedHeaders=.*content-type;host;x-amz-date;x-amz-target.*$', request.headers.get('Authorization', '')) and request.headers.get('X-Amz-Target', '').startswith('CloudApiService.GetResource'):
        return render_template('template_6.json')
    if re.match(r'^.*us-east-1.*SignedHeaders=.*host;x-amz-date.*$', request.headers.get('Authorization', '')) and request.form.get('Action') == '^ListUserPolicies$' and request.form.get('Version') == '^2010\-05\-08$':
        return render_template('template_10.json')
    if re.match(r'^.*SignedHeaders=.*content-type;host;x-amz-date.*$', request.headers.get('Authorization', '')) and request.form.get('Action') == '^DescribeVolumes$' and request.form.get('Version') == '^2016\-11\-15$':
        return render_template('template_15.json')
    if re.match(r'^.*SignedHeaders=.*content-type;host;x-amz-date.*$', request.headers.get('Authorization', '')) and request.form.get('Action') == '^ListUsers$' and request.form.get('Version') == '^2010\-05\-08$':
        return render_template('template_16.json')
    if re.match(r'^.*SignedHeaders=.*content-type;host;x-amz-date.*$', request.headers.get('Authorization', '')) and request.form.get('Action') == '^DescribeVpnGateways$' and request.form.get('Version') == '^2016\-11\-15$':
        return render_template('template_17.json')
    if re.match(r'^.*SignedHeaders=.*content-type;host;x-amz-date.*$', request.headers.get('Authorization', '')) and request.form.get('Action') == '^DescribeInstances$' and request.form.get('Version') == '^2016\-11\-15$':
        return render_template('template_18.json')
    if re.match(r'^.*ap-southeast-1.*SignedHeaders=accept;content-type;host;x-amz-date;x-amz-target.*$', request.headers.get('Authorization', '')) and request.headers.get('X-Amz-Target', '').startswith('CloudApiService.ListResources'):
        return render_template('template_19.json')
    if re.match(r'^.*ap-southeast-1.*SignedHeaders=.*content-type;host;x-amz-date;x-amz-target.*$', request.headers.get('Authorization', '')) and request.headers.get('X-Amz-Target', '').startswith('CloudApiService.GetResource'):
        return render_template('template_20.json')
    if re.match(r'^.*ap-southeast-1.*SignedHeaders=.*content-type;host;x-amz-date;x-amz-target.*$', request.headers.get('Authorization', '')) and request.headers.get('X-Amz-Target', '').startswith('CloudApiService.GetResource'):
        return render_template('template_21.json')
    if re.match(r'^.*us-east-1.*SignedHeaders=accept;content-type;host;x-amz-date;x-amz-target.*$', request.headers.get('Authorization', '')) and request.headers.get('X-Amz-Target', '').startswith('CloudApiService.ListResources'):
        return render_template('template_22.json')
    if re.match(r'^.*us-east-1.*SignedHeaders=.*content-type;host;x-amz-date;x-amz-target.*$', request.headers.get('Authorization', '')) and request.headers.get('X-Amz-Target', '').startswith('CloudApiService.GetResource'):
        return render_template('template_23.json')
    if re.match(r'^.*us-east-1.*SignedHeaders=.*content-type;host;x-amz-date;x-amz-target.*$', request.headers.get('Authorization', '')) and request.headers.get('X-Amz-Target', '').startswith('CloudApiService.GetResource'):
        return render_template('template_24.json')
    if re.match(r'^.*us-west-1.*SignedHeaders=accept;content-type;host;x-amz-date;x-amz-target.*$', request.headers.get('Authorization', '')) and request.headers.get('X-Amz-Target', '').startswith('CloudApiService.ListResources'):
        return render_template('template_25.json')
    if re.match(r'^.*us-west-1.*SignedHeaders=.*content-type;host;x-amz-date;x-amz-target.*$', request.headers.get('Authorization', '')) and request.headers.get('X-Amz-Target', '').startswith('CloudApiService.GetResource'):
        return render_template('template_26.json')
    if re.match(r'^.*us-west-1.*SignedHeaders=.*content-type;host;x-amz-date;x-amz-target.*$', request.headers.get('Authorization', '')) and request.headers.get('X-Amz-Target', '').startswith('CloudApiService.GetResource'):
        return render_template('template_27.json')
    if re.match(r'^.*eu-west-1.*SignedHeaders=accept;content-type;host;x-amz-date;x-amz-target.*$', request.headers.get('Authorization', '')) and request.headers.get('X-Amz-Target', '').startswith('CloudApiService.ListResources'):
        return render_template('template_28.json')
    if re.match(r'^.*eu-west-1.*SignedHeaders=.*content-type;host;x-amz-date;x-amz-target.*$', request.headers.get('Authorization', '')) and request.headers.get('X-Amz-Target', '').startswith('CloudApiService.GetResource'):
        return render_template('template_29.json')
    if re.match(r'^.*eu-west-1.*SignedHeaders=.*content-type;host;x-amz-date;x-amz-target.*$', request.headers.get('Authorization', '')) and request.headers.get('X-Amz-Target', '').startswith('CloudApiService.GetResource'):
        return render_template('template_30.json')
    if re.match(r'^.*us-west-2.*SignedHeaders=accept;content-type;host;x-amz-date;x-amz-target.*$', request.headers.get('Authorization', '')) and request.headers.get('X-Amz-Target', '').startswith('CloudApiService.ListResources'):
        return render_template('template_31.json')
    if re.match(r'^.*us-west-2.*SignedHeaders=.*content-type;host;x-amz-date;x-amz-target.*$', request.headers.get('Authorization', '')) and request.headers.get('X-Amz-Target', '').startswith('CloudApiService.GetResource'):
        return render_template('template_32.json')
    if re.match(r'^.*us-west-2.*SignedHeaders=.*content-type;host;x-amz-date;x-amz-target.*$', request.headers.get('Authorization', '')) and request.headers.get('X-Amz-Target', '').startswith('CloudApiService.GetResource'):
        return render_template('template_33.json')
    if re.match(r'^.*eu-west-2.*SignedHeaders=accept;content-type;host;x-amz-date;x-amz-target.*$', request.headers.get('Authorization', '')) and request.headers.get('X-Amz-Target', '').startswith('CloudApiService.ListResources'):
        return render_template('template_34.json')
    if re.match(r'^.*eu-west-2.*SignedHeaders=.*content-type;host;x-amz-date;x-amz-target.*$', request.headers.get('Authorization', '')) and request.headers.get('X-Amz-Target', '').startswith('CloudApiService.GetResource'):
        return render_template('template_35.json')
    if request.headers.get('Authorization', '').startswith('^.*eu-west-2.*SignedHeaders=.*content-type;host;x-amz-date;x-amz-target.*$') and request.headers.get('X-Amz-Target', '').startswith('CloudApiService.GetResource'):
        return render_template('template_36.json')
    if request.headers.get('Authorization', '').startswith('^.*ca-central-1.*SignedHeaders=accept;content-type;host;x-amz-date;x-amz-target.*$') and request.headers.get('X-Amz-Target', '').startswith('CloudApiService.ListResources'):
        return render_template('template_37.json')
    if request.headers.get('Authorization', '').startswith('^.*ca-central-1.*SignedHeaders=.*content-type;host;x-amz-date;x-amz-target.*$') and request.headers.get('X-Amz-Target', '').startswith('CloudApiService.GetResource'):
        return render_template('template_38.json')
    if request.headers.get('Authorization', '').startswith('^.*ca-central-1.*SignedHeaders=.*content-type;host;x-amz-date;x-amz-target.*$') and request.headers.get('X-Amz-Target', '').startswith('CloudApiService.GetResource'):
        return render_template('template_39.json')
    if request.headers.get('Authorization', '').startswith('^.*ap-southeast-1.*SignedHeaders=accept;content-type;host;x-amz-date;x-amz-target.*$') and request.headers.get('X-Amz-Target', '').startswith('CloudApiService.ListResources'):
        return render_template('template_40.json')
    if request.headers.get('Authorization', '').startswith('^.*ap-southeast-2.*SignedHeaders=accept;content-type;host;x-amz-date;x-amz-target.*$') and request.headers.get('X-Amz-Target', '').startswith('CloudApiService.ListResources'):
        return render_template('template_41.json')
    if request.headers.get('Authorization', '').startswith('^.*ap-southeast-2.*SignedHeaders=.*content-type;host;x-amz-date;x-amz-target.*$') and request.headers.get('X-Amz-Target', '').startswith('CloudApiService.GetResource'):
        return render_template('template_42.json')
    if request.headers.get('Authorization', '').startswith('^.*ap-southeast-1.*SignedHeaders=.*content-type;host;x-amz-date;x-amz-target.*$') and request.headers.get('X-Amz-Target', '').startswith('CloudApiService.GetResource'):
        return render_template('template_43.json')
    if request.headers.get('Authorization', '').startswith('^.*ap-southeast-1.*SignedHeaders=.*content-type;host;x-amz-date;x-amz-target.*$') and request.headers.get('X-Amz-Target', '').startswith('CloudApiService.GetResource'):
        return render_template('template_44.json')
    if request.headers.get('Authorization', '').startswith('^.*ap-southeast-1.*SignedHeaders=.*content-type;host;x-amz-date;x-amz-target.*$') and request.headers.get('X-Amz-Target', '').startswith('CloudApiService.GetResource'):
        return render_template('template_45.json')
    if request.headers.get('Authorization', '').startswith('^.*ap-southeast-1.*SignedHeaders=.*content-type;host;x-amz-date;x-amz-target.*$') and request.headers.get('X-Amz-Target', '').startswith('CloudApiService.GetResource'):
        return render_template('template_46.json')
    if request.headers.get('Authorization', '').startswith('^.*ap-southeast-1.*SignedHeaders=.*content-type;host;x-amz-date;x-amz-target.*$') and request.headers.get('X-Amz-Target', '').startswith('CloudApiService.GetResource'):
        return render_template('template_47.json')
    if request.headers.get('Authorization', '').startswith('^.*SignedHeaders=content-type;host;x-amz-date;x-amz-target.*$') and request.headers.get('X-Amz-Target', '').startswith('TransferService.DeleteServer'):
        return render_template('template_48.json')
    if request.headers.get('Authorization', '').startswith('^.*SignedHeaders=content-type;host;x-amz-date;x-amz-target.*$') and request.headers.get('X-Amz-Target', '').startswith('TransferService.DeleteUser'):
        return render_template('template_49.json')
    if request.headers.get('Authorization', '').startswith('^.*SignedHeaders=content-type;host;x-amz-date;x-amz-target.*$') and request.headers.get('X-Amz-Target', '').startswith('TransferService.UpdateServer'):
        return render_template('template_50.json')
    if request.headers.get('Authorization', '').startswith('^.*SignedHeaders=content-type;host;x-amz-date;x-amz-target.*$') and request.headers.get('X-Amz-Target', '').startswith('TransferService.UpdateUser'):
        return render_template('template_51.json')
    if request.headers.get('Authorization', '').startswith('^.*SignedHeaders=content-type;host;x-amz-date;x-amz-target.*$') and request.headers.get('X-Amz-Target', '').startswith('TransferService.CreateServer'):
        return render_template('template_52.json')
    if request.headers.get('Authorization', '').startswith('^.*SignedHeaders=content-type;host;x-amz-date;x-amz-target.*$') and request.headers.get('X-Amz-Target', '').startswith('TransferService.CreateServer'):
        return render_template('template_53.json')
    if request.headers.get('Authorization', '').startswith('^.*SignedHeaders=content-type;host;x-amz-date;x-amz-target.*$') and request.headers.get('X-Amz-Target', '').startswith('TransferService.StopServer'):
        return render_template('template_54.json')
    if request.headers.get('Authorization', '').startswith('^.*SignedHeaders=content-type;host;x-amz-date;x-amz-target.*$') and request.headers.get('X-Amz-Target', '').startswith('BaldrApiService.DescribeClusters'):
        return render_template('template_55.json')
    if request.headers.get('Authorization', '').startswith('^.*/rubbish-region/.*SignedHeaders=content-type;host;x-amz-date;x-amz-target.*$') and request.headers.get('X-Amz-Target', '').startswith('BaldrApiService.DescribeBackups'):
        return render_template('template_56.json')
    if request.headers.get('Authorization', '').startswith('^.*/another-rubbish-region/.*SignedHeaders=content-type;host;x-amz-date;x-amz-target.*$') and request.headers.get('X-Amz-Target', '').startswith('BaldrApiService.DescribeBackups'):
        return render_template('template_57.json')
    if request.headers.get('Authorization', '').startswith('^.*SignedHeaders=content-type;host;x-amz-date;x-amz-target.*$') and request.headers.get('X-Amz-Target', '').startswith('BaldrApiService.DescribeBackups'):
        return render_template('template_58.json')
    if request.headers.get('Authorization', '').startswith('^.*SignedHeaders=content-encoding;content-type;host;x-amz-date;x-amz-target.*$') and request.headers.get('X-Amz-Target', '').startswith('Logs_20140328.GetLogEvents'):
        return render_template('template_59.json')
    if request.headers.get('Authorization', '').startswith('^.*ap-southeast-1.*SignedHeaders.*content-type;host;x-amz-date;x-amz-target.*$') and request.headers.get('X-Amz-Target', '').startswith('CloudApiService.GetResource'):
        return render_template('template_60.json')
    if request.headers.get('Authorization', '').startswith('^.*ap-southeast-1.*SignedHeaders.*content-type;host;x-amz-date;x-amz-target.*$') and request.headers.get('X-Amz-Target', '').startswith('CloudApiService.GetResource'):
        return render_template('template_61.json')
    if request.headers.get('Authorization', '').startswith('^.*ap-southeast-2.*SignedHeaders.*content-type;host;x-amz-date;x-amz-target.*$') and request.headers.get('X-Amz-Target', '').startswith('CloudApiService.GetResource'):
        return render_template('template_62.json')
    if request.headers.get('Authorization', '').startswith('^.*SignedHeaders.*content-type;host;x-amz-date;x-amz-target.*$') and request.headers.get('X-Amz-Target', '').startswith('CloudApiService.GetResource'):
        return render_template('template_63.json')
    if request.headers.get('Authorization', '').startswith('^.*ap-southeast-1.*SignedHeaders.*content-type;host;x-amz-date;x-amz-target.*$') and request.headers.get('X-Amz-Target', '').startswith('CloudApiService.ListResources'):
        return render_template('template_64.json')
    if request.headers.get('Authorization', '').startswith('^.*ap-southeast-2.*SignedHeaders.*content-type;host;x-amz-date;x-amz-target.*$') and request.headers.get('X-Amz-Target', '').startswith('CloudApiService.ListResources'):
        return render_template('template_65.json')
    if request.headers.get('Authorization', '').startswith('^.*us-east-1.*SignedHeaders.*content-type;host;x-amz-date;x-amz-target.*$') and request.headers.get('X-Amz-Target', '').startswith('CloudApiService.ListResources'):
        return render_template('template_66.json')
    if request.headers.get('Authorization', '').startswith('^.*us-east-1.*SignedHeaders.*content-type;host;x-amz-date;x-amz-target.*$') and request.headers.get('X-Amz-Target', '').startswith('CloudApiService.GetResource'):
        return render_template('template_67.json')
    if request.headers.get('Authorization', '').startswith('^.*us-east-1.*SignedHeaders.*content-type;host;x-amz-date;x-amz-target.*$') and request.headers.get('X-Amz-Target', '').startswith('CloudApiService.GetResource'):
        return render_template('template_68.json')
    if request.headers.get('Authorization', '').startswith('^.*us-east-1.*SignedHeaders.*content-type;host;x-amz-date;x-amz-target.*$') and request.headers.get('X-Amz-Target', '').startswith('CloudApiService.GetResource'):
        return render_template('template_69.json')
    if request.headers.get('Authorization', '').startswith('^.*SignedHeaders=accept;content-type;host;x-amz-date;x-amz-target.*$') and request.headers.get('X-Amz-Target', '').startswith('CloudApiService.ListResourceRequests'):
        return render_template('template_70.json')
    if request.headers.get('Authorization', '').startswith('^.*SignedHeaders=.*content-type;host;x-amz-date;x-amz-target.*$') and request.headers.get('X-Amz-Target', '').startswith('CloudApiService.CreateResource'):
        return render_template('template_71.json')
    if request.headers.get('Authorization', '').startswith('^AWS4-HMAC-SHA256 Credential=.*/cloudcontrolapi/aws4_request.*SignedHeaders=.*content-type;host;x-amz-date;x-amz-target.*$') and request.headers.get('X-Amz-Target', '').startswith('CloudApiService.CreateResource'):
        return render_template('template_72.json')
    if request.headers.get('Authorization', '').startswith('^AWS4-HMAC-SHA256 Credential=.*/cloudcontrolapi/aws4_request.*SignedHeaders=.*content-type;host;x-amz-date;x-amz-target.*$') and request.headers.get('X-Amz-Target', '').startswith('CloudApiService.CreateResource'):
        return render_template('template_73.json')
    if request.headers.get('Authorization', '').startswith('^.*SignedHeaders=.*content-type;host;x-amz-date;x-amz-target.*$') and request.headers.get('X-Amz-Target', '').startswith('CloudApiService.UpdateResource'):
        return render_template('template_74.json')
    if request.headers.get('Authorization', '').startswith('^.*SignedHeaders=.*content-type;host;x-amz-date;x-amz-target.*$') and request.headers.get('X-Amz-Target', '').startswith('CloudApiService.CreateResource'):
        return render_template('template_75.json')
    if request.headers.get('Authorization', '').startswith('^.*SignedHeaders=.*content-type;host;x-amz-date;x-amz-target.*$') and request.headers.get('X-Amz-Target', '').startswith('CloudApiService.DeleteResource'):
        return render_template('template_76.json')
    if request.headers.get('Authorization', '').startswith('^.*SignedHeaders=.*content-type;host;x-amz-date;x-amz-target.*$') and request.headers.get('X-Amz-Target', '').startswith('CloudApiService.CancelResourceRequest'):
        return render_template('template_77.json')
    if request.headers.get('Authorization', '').startswith('^.*SignedHeaders=.*content-type;host;x-amz-date;x-amz-target.*$') and request.headers.get('X-Amz-Target', '').startswith('CloudApiService.GetResourceRequestStatus'):
        return render_template('template_78.json')
    if request.headers.get('Authorization', '').startswith('^.*SignedHeaders=.*content-type;host;x-amz-date.*$') and request.form.get('Action') == '^CreateVolume$' and request.form.get('Version') == '^2016\-11\-15$' and request.form.get('Size') == '^10$' and request.form.get('TagSpecification.1.ResourceType') == 'volume' and request.form.get('TagSpecification.1.Tag.1.Key') == 'stack' and request.form.get('TagSpecification.1.Tag.1.Value') == 'production' and request.form.get('TagSpecification.1.Tag.2.Key') == 'name' and request.form.get('TagSpecification.1.Tag.2.Value') == 'multi-tag-volume':
        return render_template('template_79.json')
    if request.headers.get('Authorization', '').startswith('^.*SignedHeaders=.*content-type;host;x-amz-date.*$') and request.form.get('Action') == '^StartInstances$' and request.form.get('Version') == '^2016\-11\-15$' and request.form.get('InstanceId.1') == 'id-001':
        return render_template('template_80.json')
    if request.headers.get('Authorization', '').startswith('^.*SignedHeaders=.*content-type;host;x-amz-date.*$') and request.form.get('Action') == '^ModifyVolume$' and request.form.get('Version') == '^2016\-11\-15$' and request.form.get('Size') == '^12$' and request.form.get('VolumeId') == 'vol-000000000000001':
        return render_template('template_81.json')
    return jsonify({'error': 'No matching template found'}), 404

if __name__ == "__main__":
    app.run(debug=True, host="0.0.0.0", port=5000)
