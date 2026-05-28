from flask import Flask, request

from werkzeug.datastructures import ImmutableMultiDict

import json

import base64

import logging

from typing import List
from datetime import datetime, timedelta, timezone
from itertools import count


"""
- Google stuff based upon [the service account key flow doco](https://developers.google.com/identity/protocols/oauth2/service-account#httprest_1).

Example invocation:

```bash
flask --app=test/python/stackql_test_tooling/flask/oauth/token_srv run --port=8070
```


Then, in some other shell:

```bash

curl -d 'grant_type=urn%3Aietf%3Aparams%3Aoauth%3Agrant-type%3Ajwt-bearer&assertion=eyJhIjogImIifQ==.eyJzdWIiOiAianMifQ==.ZXlKaElqb2dJbUlpZlE9PS5leUp6ZFdJaU9pQWlhbk1pZlE9PQ==' http://127.0.0.1:8070/google/simple/token
```
"""

app = Flask(__name__)

app.logger.setLevel(logging.INFO)

class GoogleServiceAccountJWT(object):

    def __init__(self, encoded_jwt: str) -> None:
        self._encoded_jwt = encoded_jwt
        split_jwt = encoded_jwt.split('.')
        if len(split_jwt) != 3:
            raise ValueError(f'invalid JWT; length {len(split_jwt)} != 3')
        self._jwt_header: dict = json.loads(base64.urlsafe_b64decode(self._pad(split_jwt[0])).decode('utf-8'))
        self._jwt_claims: dict = json.loads(base64.urlsafe_b64decode(self._pad(split_jwt[1])).decode('utf-8'))
        self._jwt_signature: bytes = base64.urlsafe_b64decode(self._pad(split_jwt[2]))
    
    def _pad(self, s: str) -> str:
        return s + '=' * (4 - len(s) % 4)

    def generate_google_response_dict(self, default_scopes: List[str]=["https://www.googleapis.com/auth/cloud-platform"]) -> dict:
        return {
            "access_token": "some-dummy-token",
            "scope": ' '.join(default_scopes),
            "token_type": "Bearer",
            "expires_in": 3600
        }

# conforming to [RFC 6749](https://datatracker.ietf.org/doc/html/rfc6749#section-5.1)
_SIMPLE_RESPONSE = {
            "access_token": "some-dummy-token",
            "scope": 'my-scope',
            "token_type": "Bearer",
            "refresh_token": "some-dummy-refresh-token", # optional, per the RFC
            "expires_in": 3600
        }

_AZURE_FEDERATED_TOKEN_COUNTER = count(1)
_GCP_STS_TOKEN_COUNTER = count(1)
_GCP_IMPERSONATED_TOKEN_COUNTER = count(1)


def _utc_now() -> datetime:
    return datetime.now(timezone.utc)


def _isoformat_z(ts: datetime) -> str:
    return ts.astimezone(timezone.utc).replace(microsecond=0).isoformat().replace('+00:00', 'Z')

def _form_to_str(form: ImmutableMultiDict) -> str:
    return json.dumps(form.to_dict(flat=False), sort_keys=True)

@app.route("/google/simple/token", methods=['POST'])
def google_simple_token():
    request_data = request.form
    proffrered_jwt = request_data['assertion']
    app.logger.info(f'proffered_jwt: {proffrered_jwt}')
    google_jwt: GoogleServiceAccountJWT = GoogleServiceAccountJWT(proffrered_jwt)
    return json.dumps(google_jwt.generate_google_response_dict(), sort_keys=True)

@app.route("/contrived/simple/error/token", methods=['POST'])
def google_simple_error_token():
    return json.dumps({"msg": "auth failed"}, sort_keys=True), 401

@app.route("/contrived/simple/token", methods=['POST'])
def contrived_simple_token():
    request_data = request.form
    app.logger.info(f'POST /contrived/simple/token request data: {_form_to_str(request_data)}')
    return json.dumps(_SIMPLE_RESPONSE, sort_keys=True)


@app.route("/aws/sts", methods=['POST'])
def aws_sts_assume_role_with_web_identity():
    request_data = request.form
    action = request_data.get('Action')
    token = request_data.get('WebIdentityToken')
    if action != 'AssumeRoleWithWebIdentity' or not token:
        return json.dumps({"msg": "malformed aws_web_identity request"}, sort_keys=True), 400
    expiration = _isoformat_z(_utc_now() + timedelta(seconds=1))
    response_xml = f"""<?xml version=\"1.0\" encoding=\"UTF-8\"?>
<AssumeRoleWithWebIdentityResponse xmlns=\"https://sts.amazonaws.com/doc/2011-06-15/\">
  <AssumeRoleWithWebIdentityResult>
    <Credentials>
      <AccessKeyId>ASIA_MOCK_IDFED</AccessKeyId>
      <SecretAccessKey>mock-idfed-secret</SecretAccessKey>
      <SessionToken>mock-idfed-session-token</SessionToken>
      <Expiration>{expiration}</Expiration>
    </Credentials>
  </AssumeRoleWithWebIdentityResult>
</AssumeRoleWithWebIdentityResponse>"""
    return response_xml, 200, {"Content-Type": "text/xml"}


@app.route("/gcp/sts/token", methods=['POST'])
def gcp_workload_identity_sts_token():
    request_data = request.form
    subject_token = request_data.get('subject_token')
    if not subject_token:
        return json.dumps({"msg": "missing subject_token"}, sort_keys=True), 400
    token_index = next(_GCP_STS_TOKEN_COUNTER)
    response = {
        "access_token": f"gcp-fed-token-{token_index}",
        "issued_token_type": "urn:ietf:params:oauth:token-type:access_token",
        "token_type": "Bearer",
        "expires_in": 1,
    }
    return json.dumps(response, sort_keys=True)


@app.route("/gcp/iamcredentials/generateAccessToken", methods=['POST'])
def gcp_workload_identity_impersonation():
    token_index = next(_GCP_IMPERSONATED_TOKEN_COUNTER)
    response = {
        "accessToken": f"gcp-impersonated-token-{token_index}",
        "expireTime": _isoformat_z(_utc_now() + timedelta(seconds=1)),
    }
    return json.dumps(response, sort_keys=True)


@app.route("/azure/federated/token", methods=['POST'])
def azure_federated_token():
    request_data = request.form
    subject_token = request_data.get('subject_token')
    if not subject_token:
        return json.dumps({"msg": "missing subject_token"}, sort_keys=True), 400
    token_index = next(_AZURE_FEDERATED_TOKEN_COUNTER)
    response = {
        "access_token": f"azure-fed-token-{token_index}",
        "token_type": "Bearer",
        "expires_in": 1,
    }
    return json.dumps(response, sort_keys=True)
