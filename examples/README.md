

## BYO Registry


```bash
PROVIDER_REGISTRY_ROOT_DIR="$(pwd)/../examples/registry"

REG_STR='{ "url": "file://'${PROVIDER_REGISTRY_ROOT_DIR}'", "localDocRoot": "'${PROVIDER_REGISTRY_ROOT_DIR}'",  "useEmbedded": false, "verifyConfig": {"nopVerify": true } }'

AUTH_STR='{ "publicapis": { "type": "null_auth" }  }'


```