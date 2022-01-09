
# Google provider examples

## Assumptions

  - `infraql` is in your `${PATH}`.
  - You have a google service account json key at the file location `${PATH_TO_KEY_FILE}`.

If using `service account` auth against the `google` provider, then no ancillary information is required.  If howevere, you are using another key type / provider, then more runtime information is required, eg:

Google:

```
./infraql shell --keyfilepath=${HOME}/moonlighting/infraql-original/keys/sa-key.json --provider=google
```

Non-google:

```
./infraql shell --keyfilepath=${HOME}/moonlighting/infraql-original/keys/okta-token.txt --provider=okta --keyfiletype=api_key
```

### SELECT

```
infraql \
  --keyfilepath=${PATH_TO_KEY_FILE} exec  \
  "select * from compute.instances WHERE zone = '${YOUR_GOOGLE_ZONE}' AND project = '${YOUR_GOOGLE_PROJECT}' ;" ; echo

```

Or...

```
infraql \
  --keyfilepath=${PATH_TO_KEY_FILE} exec  \
  "select selfLink, projectNumber from storage.buckets WHERE location = '${YOUR_GOOGLE_ZONE}' AND project = '${YOUR_GOOGLE_PROJECT}' ;" ; echo

```

### SHOW SERVICES

```
infraql --providerroot=../test/.infraql \
  --configfile=../test/.iqlrc exec \
  "SHOW SERVICES from google ;" ; echo

```