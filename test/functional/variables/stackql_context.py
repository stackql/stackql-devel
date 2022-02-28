
import json
import os


REPOSITORY_ROOT = os.path.abspath(os.path.join(__file__, "../../../..")).replace("\\","/")
REGISTRY_ROOT   = os.path.join(REPOSITORY_ROOT, "test/registry-mocked").replace("\\","/")
STACKQL_EXE     = os.path.join(REPOSITORY_ROOT, "build/stackql").replace("\\","/")
_RESISTRY_CFG    = { 
  "url": f"file://{REGISTRY_ROOT}",
  "localDocRoot": f"{REGISTRY_ROOT}",
  "useEmbedded": False,
  "verifyConfig": {
    "nopVerify": True 
  } 
}
_AUTH_CFG={ 
  "google": { 
    "credentialsfilepath": f"{REPOSITORY_ROOT}/test/assets/credentials/dummy/google/functional-test-dummy-sa-key.json",
    "type": "service_account"
  }, 
  "okta": { 
    "credentialsenvvar": "OKTA_SECRET_KEY",
    "type": "api_key" 
  } 
}

with open(f"{REPOSITORY_ROOT}/test/assets/credentials/dummy/okta/api-key.txt", 'r') as f:
    OKTA_SECRET_STR = f.read()

RESISTRY_CFG_STR = json.dumps(_RESISTRY_CFG)
AUTH_CFG_STR = json.dumps(_AUTH_CFG)
SHOW_PROVIDERS_STR = "show providers;"
SHOW_OKTA_SERVICES_FILTERED_STR  = "show services from okta like 'app%';"
SHOW_OKTA_APPLICATION_RESOURCES_FILTERED_STR  = "show resources from okta.application like 'gr%';"
SELECT_CONTAINER_SUBNET_AGG_DESC = "select ipCidrRange, sum(5) cc  from  google.container.`projects.aggregated.usableSubnetworks` where projectsId = 'testing-project' group by \"ipCidrRange\" having sum(5) >= 5 order by ipCidrRange desc;"
JSON_INIT_FILE_PATH = f'{REPOSITORY_ROOT}/test/server/expectations/static-gcp-expectations.json'
MOCKSERVER_JAR = '/usr/local/lib/mockserver/mockserver-netty-jar-with-dependencies.jar'

with open(f"{REPOSITORY_ROOT}/test/assets/expected/aggregated-select/google/container/agg-subnetworks-allowed/table/simple-count-grouped-variant-desc.txt", 'r') as f:
    SELECT_CONTAINER_SUBNET_AGG_DESC_EXPECTED = f.read().strip()

