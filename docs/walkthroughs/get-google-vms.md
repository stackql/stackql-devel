
## Setup

First, create a google service account key using the GCP Console, per [the GCP documentation](https://cloud.google.com/iam/docs/keys-create-delete).  Grant the service account at least `Viewer` role equivalent privileges, per [the GCP dumentation](https://cloud.google.com/iam/docs/create-service-agents#grant-roles).

Then, do this in bash:

```bash setup stackql-shell

export GOOGLE_CREDENTIALS="$(cat cicd/keys/testing/google-credentials.json)";

stackql shell --approot=./test/tmp/.get-google-vms.stackql
```

## Method

Do this in the `stackql` shell, replacing `<project>` with your GCP project name:

```sql stackql-shell input required project=ryuki-it-sandbox-01

registry pull google;

select 
  name, 
  id 
FROM google.compute.instances 
WHERE 
  project = 'ryuki-it-sandbox-01' 
  AND zone = 'australia-southeast1-a'
;

```

## Result


You will see something very much like this included in the output, presuming you have one VM (if you have zero, only the headers should appper, more VMs means more rows):

```sql stackql stdout table-contains-data
|--------------------------------------------------|---------------------|
|                       name                       |         id          |
|--------------------------------------------------|---------------------|
| any-compute-cluster-1-default-abcd-00000001-0001 | 1000000000000000001 |
|--------------------------------------------------|---------------------|
```

<!---  STDERR_REGEX_EXACT
google\ provider,\ version\ 'v24.11.00274'\ successfully\ installed
goodbye
-->

## Cleanup

```bash teardown best-effort

rm -rf ./test/tmp/.get-google-vms.stackql

```