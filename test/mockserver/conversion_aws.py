from collections import OrderedDict
import json
import os

aws_input = json.load(open('/Users/admin/stackql/stackql-devel/test/mockserver/expectations/static-aws-expectations.json'))

new_aws_cfg = json.load(open('/Users/admin/stackql/stackql-devel/test/python/flask/aws/root_path_cfg.json'), object_pairs_hook=OrderedDict)

TPL_DIR = '/Users/admin/stackql/stackql-devel/test/python/flask/aws/templates'

import base64

i = 0
for k, v in new_aws_cfg.items():
    print('')
    print(f'i = {i}, k = {k}')
    template = v.get('template')
    rhs = aws_input[i]
    i += 1
    raw_response_string = ''
    rhs_response_body = rhs.get('httpResponse', {}).get('body', {})
    if not rhs_response_body:
        print(f'No response body found for i = {i}, k = {k}')
        continue
    if rhs_response_body.get('base64Bytes'):
        raw_response_string = base64.b64decode(rhs_response_body.get('base64Bytes')).decode()
    else:
        raw_response_string = json.dumps(rhs_response_body, indent=4)
    print(f'i = {i}, k = {k}, raw_response_string: {raw_response_string}')
    

    if not template:
        continue
    output_file = f"{TPL_DIR}/{template}"
    file_exists = os.path.exists(output_file)
    if file_exists:
        continue

    with open(output_file, 'w') as f:
        f.write(raw_response_string)
    