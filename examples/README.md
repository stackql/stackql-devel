

## BYO Registry


```bash
PROVIDER_REGISTRY_ROOT_DIR="$(pwd)/../examples/registry"

REG_STR='{ "url": "file://'${PROVIDER_REGISTRY_ROOT_DIR}'", "localDocRoot": "'${PROVIDER_REGISTRY_ROOT_DIR}'",  "useEmbedded": false, "verifyConfig": {"nopVerify": true } }'

AUTH_STR='{ "publicapis": { "type": "null_auth" }  }'

./stackql --auth="${AUTH_STR}" --registry="${REG_STR}" exec "select API from publicapis.api.apis where API like 'Dog%' limit 10;"

./stackql --auth="${AUTH_STR}" --registry="${REG_STR}" exec "select API from publicapis.api.random where title =  'Dog';"

./stackql --auth="${AUTH_STR}" --registry="${REG_STR}" exec "select * from publicapis.api.categories limit 5;"
```