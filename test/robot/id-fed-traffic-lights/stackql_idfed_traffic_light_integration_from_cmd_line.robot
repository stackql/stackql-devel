*** Settings ***
Resource          ${CURDIR}/stackql.resource

*** Test Cases ***

IDFed AWS S3 Buckets List
    Sleep    2s
    ${awsRoleArn} =    OperatingSystem.Get Environment Variable    STACKQL_IDFED_ROLE_ARN
    Should Not Be Empty    ${awsRoleArn}
    ${awsAuthCfg} =    Catenate
    ...    { "aws": { "type":"aws_web_identity", "aws_role_arn": "${awsRoleArn}", "aws_sts_region": "us-east-1", "oidc_subject_token_file_env_var": "OIDC_SUBJECT_TOKEN_FILE" } }
    ${bucketsListQuery} =    Catenate
    ...    select * from aws.s3.buckets where region = 'ap-southeast-2';
    ${result} =    Run Process
    ...    ${STACKQL_EXE}
    ...    \-\-auth
    ...    ${awsAuthCfg}
    ...    \-\-registry
    ...    { "url": "file://${REPOSITORY_ROOT}/test/registry", "localDocRoot": "${REPOSITORY_ROOT}/test/registry", "verifyConfig": { "nopVerify": true } }
    ...    exec
    ...    ${bucketsListQuery}
    ...    cwd=${REPOSITORY_ROOT}
    ...    stdout=${CURDIR}/tmp/IDFed-AWS-S3-Buckets-List.tmp
    ...    stderr=${CURDIR}/tmp/IDFed-AWS-S3-Buckets-List-stderr.tmp
    Should Be Equal As Integers    ${result.rc}           0
    Should Be Empty                ${result.stderr}
    Should Contain                 ${result.stdout}       stackql-trial-bucket-02

IDFed Azure VNETs List
    Sleep    2s
    ${azureTargetSubscription} =    OperatingSystem.Get Environment Variable    AZURE_TARGET_SUBSCRIPTION_ID
    Should Not Be Empty    ${azureTargetSubscription}
    ${azureAuthCfg} =    Catenate
    ...    { "azure": { "type": "azure_federated", "azure_tenant_id": "${AZURE_TENANT_ID}", "client_id": "${AZURE_CLIENT_ID}", "scopes": ["https://management.azure.com/.default"], "oidc_subject_token_file_env_var": "OIDC_SUBJECT_TOKEN_FILE" } }
    ${bucketsListQuery} =    Catenate
    ...    select location, name from azure.network.virtual_networks where subscriptionId = '${azureTargetSubscription}';
    ${result} =    Run Process
    ...    ${STACKQL_EXE}
    ...    \-\-auth
    ...    ${azureAuthCfg}
    ...    \-\-registry
    ...    { "url": "file://${REPOSITORY_ROOT}/test/registry", "localDocRoot": "${REPOSITORY_ROOT}/test/registry", "verifyConfig": { "nopVerify": true } }
    ...    exec
    ...    ${bucketsListQuery}
    ...    cwd=${REPOSITORY_ROOT}
    ...    stdout=${CURDIR}/tmp/IDFed-Azure-VNETs-List.tmp
    ...    stderr=${CURDIR}/tmp/IDFed-Azure-VNETs-List-stderr.tmp
    Should Be Equal As Integers    ${result.rc}           0
    Should Be Empty                ${result.stderr}
    Should Contain                 ${result.stdout}       inspector-network

IDFed Google Buckets List
    Sleep    2s
    ${gcpAudience} =    OperatingSystem.Get Environment Variable    GCP_OIDC_AUDIENCE
    ${gcpServiceAccount} =    OperatingSystem.Get Environment Variable    GCP_SERVICE_ACCOUNT_EMAIL
    Should Not Be Empty    ${gcpAudience}
    Should Not Be Empty    ${gcpServiceAccount}
    ${gcpAuthCfg} =    Catenate
    ...    { "google": { "type": "gcp_workload_identity", "gcp_workload_identity_audience": "${gcpAudience}", "gcp_service_account_impersonation_url": "https://iamcredentials.googleapis.com/v1/projects/-/serviceAccounts/${gcpServiceAccount}:generateAccessToken", "scopes": ["https://www.googleapis.com/auth/cloud-platform"], "oidc_subject_token_file_env_var": "OIDC_SUBJECT_TOKEN_FILE" } }
    ${bucketsListQuery} =    Catenate
    ...    select location, name from google.storage.buckets where project = 'stackql-demo';
    ${result} =    Run Process
    ...    ${STACKQL_EXE}
    ...    \-\-auth
    ...    ${gcpAuthCfg}
    ...    \-\-registry
    ...    { "url": "file://${REPOSITORY_ROOT}/test/registry", "localDocRoot": "${REPOSITORY_ROOT}/test/registry", "verifyConfig": { "nopVerify": true } }
    ...    exec
    ...    ${bucketsListQuery}
    ...    cwd=${REPOSITORY_ROOT}
    ...    stdout=${CURDIR}/tmp/IDFed-Google-Buckets-List.tmp
    ...    stderr=${CURDIR}/tmp/IDFed-Google-Buckets-List-stderr.tmp
    Should Be Equal As Integers    ${result.rc}           0
    Should Be Empty                ${result.stderr}
    Should Contain                 ${result.stdout}       stackql-demo-bucket
