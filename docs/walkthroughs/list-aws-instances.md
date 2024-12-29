
## Setup

First, create a set of AWS CLI credentials per [the AWS documentation](https://docs.aws.amazon.com/cli/latest/userguide/cli-authentication-user.html#cli-authentication-user-get), and store them in the appropriate environment variables `AWS_ACCESS_KEY_ID` and `AWS_SECRET_ACCESS_KEY`.

Then, do this in bash:

```bash setup stackql-shell credentials_path=cicd/keys/testing/google-ro-credentials.json app_root_path=./test/tmp/.list-aws-instances.stackql


stackql shell --approot=<<app_root_path>>
```

## Method

Do this in the `stackql` shell, replacing `<<project>>` with your GCP project name, and `<<zone>>` as desired, eg: `australia-southeast1-a`:

```sql stackql-shell input required project=stackql-demo zone=australia-southeast1-a

registry pull aws;

SELECT instance_id, region
FROM aws.ec2.instances
WHERE region IN ('us-east-1', 'eu-west-1');

```

## Result


You will see exactly this included in the output:

```sql expectation stdout-contains-all
|---------------------|-------------------------|
|        name         |          kind           |
|---------------------|-------------------------|
| nvidia-tesla-t4-vws | compute#acceleratorType |
|---------------------|-------------------------|
| nvidia-tesla-t4     | compute#acceleratorType |
|---------------------|-------------------------|
| nvidia-tesla-p4-vws | compute#acceleratorType |
|---------------------|-------------------------|
| nvidia-tesla-p4     | compute#acceleratorType |
|---------------------|-------------------------|
```

## Cleanup

```bash teardown best-effort app_root_path=./test/tmp/.list-aws-instances.stackql

rm -rf <<app_root_path>>

```