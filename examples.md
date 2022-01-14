
# Google provider examples

## Assumptions

  - `stackql` is in your `${PATH}`.
  - You have an appropriate key file at the file location `${PATH_TO_KEY_FILE}`.  For example, with the google provider, one might use a service account json key.

If using `service account` auth against the `google` provider, then no ancillary information is required.  If howevere, you are using another key type / provider, then more runtime information is required, eg:

Google:

```sh
AUTH_STR='{ "google": { "keyfilepath": "/Users/admin/moonlighting/stackql-original/keys/sa-key.json" }, "okta": { "keyfilepath": "/Users/admin/moonlighting/stackql-original/keys/okta-token.txt", "keyfiletype": "api_key" } }'

./stackql shell --auth="${AUTH_STR}"


```

### SELECT

```
stackql \
  --auth="${AUTH_STR}" exec  \
  "select * from compute.instances WHERE zone = '${YOUR_GOOGLE_ZONE}' AND project = '${YOUR_GOOGLE_PROJECT}' ;" ; echo

```

Or...

```
stackql \
  --auth="${AUTH_STR}" exec  \
  "select selfLink, projectNumber from storage.buckets WHERE location = '${YOUR_GOOGLE_ZONE}' AND project = '${YOUR_GOOGLE_PROJECT}' ;" ; echo

```

### SHOW SERVICES

```
stackql --providerroot=../test/.stackql \
  --configfile=../test/.iqlrc exec \
  "SHOW SERVICES from google ;" ; echo

```