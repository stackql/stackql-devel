
import json
import os


REPOSITORY_ROOT = os.path.abspath(os.path.join(__file__, "../../../..")).replace("\\","/")
REGISTRY_ROOT   = os.path.join(REPOSITORY_ROOT, "test/registry").replace("\\","/")
STACKQL_EXE     = os.path.join(REPOSITORY_ROOT, "build/stackql").replace("\\","/")
_RESISTRY_CFG    = { 
  "url": f"file://{REGISTRY_ROOT}",
  "localDocRoot": f"{REGISTRY_ROOT}",
  "useEmbedded": False,
  "verifyConfig": {
    "nopVerify": True 
  } 
}
RESISTRY_CFG_STR = json.dumps(_RESISTRY_CFG)
SHOW_PROVIDERS_STR = "show providers;"
SHOW_OKTA_SERVICES_FILTERED_STR  = "show services from okta like 'app%';"
