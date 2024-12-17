

# HTTP(S) servers for simulated integration / regression testing

## Flask

We have now migrated entirely to [flask](https://flask.palletsprojects.com/en/stable/), from the prior java [mockserver](https://www.mock-server.com/).  There is no disparaging of mockserver whatsoever; rather this was motivated in large part by different behaviour against versions of `java` / dependency libraries, also by the community support and knowledge base for `flask` and `jinja`.  That said, the mock defninitions to some degree are a holdover from `mockserver`; this should diminish over time.

One pertinent fact in life with `flask` is that processes die hard; so it generally pays this before testing mocks:

```bash
pgrep -f flask | xargs kill -9
```


### To Run

GCP mocks:

```bash
java  -Dfile.encoding=UTF-8 -Dmockserver.initializationJsonPath=${HOME}/stackql/stackql-devel/test/mockserver/expectations/static-gcp-expectations.json -jar ${HOME}/stackql/stackql-devel/test/downloads/mockserver-netty-5.12.0-shaded.jar  -serverPort 1080 -logLevel INFO
```

Azure mocks:

```bash
java  -Dfile.encoding=UTF-8 -Dmockserver.initializationJsonPath=${HOME}/stackql/stackql-devel/test/mockserver/expectations/static-azure-expectations.json -jar ${HOME}/stackql/stackql-devel/test/downloads/mockserver-netty-5.12.0-shaded.jar  -serverPort 1095 -logLevel INFO
```

Okta mocks:

```bash
java  -Dfile.encoding=UTF-8 -Dmockserver.initializationJsonPath=${HOME}/stackql/stackql-devel/test/mockserver/expectations/static-okta-expectations.json -jar ${HOME}/stackql/stackql-devel/test/downloads/mockserver-netty-5.12.0-shaded.jar  -serverPort 1090 -logLevel INFO
```

AWS mocks:

```bash
java  -Dfile.encoding=UTF-8 -Dmockserver.initializationJsonPath=${HOME}/stackql/stackql-devel/test/mockserver/expectations/static-aws-expectations.json -jar ${HOME}/stackql/stackql-devel/test/downloads/mockserver-netty-5.12.0-shaded.jar  -serverPort 1091 -logLevel INFO
```

Github mocks:

```bash
java  -Dfile.encoding=UTF-8 -Dmockserver.initializationJsonPath=${HOME}/stackql/stackql-devel/test/mockserver/expectations/static-github-expectations.json -jar ${HOME}/stackql/stackql-devel/test/downloads/mockserver-netty-5.12.0-shaded.jar  -serverPort 1093 -logLevel INFO
```

Sumologic mocks:

```bash
java  -Dfile.encoding=UTF-8 -Dmockserver.initializationJsonPath=${HOME}/stackql/stackql-devel/test/mockserver/expectations/static-sumologic-expectations.json -jar ${HOME}/stackql/stackql-devel/test/downloads/mockserver-netty-5.12.0-shaded.jar  -serverPort 1096 -logLevel INFO
```

Digitalocean mocks:

```bash
java  -Dfile.encoding=UTF-8 -Dmockserver.initializationJsonPath=${HOME}/stackql/stackql-devel/test/mockserver/expectations/static-digitalocean-expectations.json -jar ${HOME}/stackql/stackql-devel/test/downloads/mockserver-netty-5.12.0-shaded.jar  -serverPort 1097 -logLevel INFO
```

`googleadmin` mocks:

```bash
java  -Dfile.encoding=UTF-8 -Dmockserver.initializationJsonPath=${HOME}/stackql/stackql-devel/test/mockserver/expectations/static-google-admin-expectations.json -jar ${HOME}/stackql/stackql-devel/test/downloads/mockserver-netty-5.12.0-shaded.jar  -serverPort 1098 -logLevel INFO
```

stackql auth testing mocks:

```bash
java  -Dfile.encoding=UTF-8 -Dmockserver.initializationJsonPath=${HOME}/stackql/stackql-devel/test/mockserver/expectations/static-auth-testing-expectations.json -jar ${HOME}/stackql/stackql-devel/test/downloads/mockserver-netty-5.12.0-shaded.jar  -serverPort 1170 -logLevel INFO
```

### Expectations from local file

As per [expectations/static-gcp-expectations.json](/test/server/expectations/static-gcp-expectations.json)


Basic idea is to rewrite openapi docs and also dummy credentials file such that 
all requests go to localhost.  We will pass in the dummy server CA to StackQL at init time.
This will obviously only occur in testing.

```
"select ipCidrRange, sum(5) cc  from  google.container.`projects.aggregated.usableSubnetworks` where projectsId = 'testing-project' group by \"ipCidrRange\" having sum(5) >= 5 order by ipCidrRange desc;"
```


### Manually testing mocks

With embedded `sqlite` (default):

```bash
export workspaceFolder='/path/to/repository/root'  # change this

stackql --registry="{ \"url\": \"file://${workspaceFolder}/test/registry-mocked\", \"localDocRoot\": \"${workspaceFolder}/test/registry-mocked\", \"verifyConfig\": { \"nopVerify\": true } }" --tls.allowInsecure shell
```

With `postgres`:

```bash
docker compose -f docker-compose-externals.yml up postgres_stackql -d

export workspaceFolder='/path/to/repository/root'  # change this

stackql --registry="{ \"url\": \"file://${workspaceFolder}/test/registry-mocked\", \"localDocRoot\": \"${workspaceFolder}/test/registry-mocked\", \"verifyConfig\": { \"nopVerify\": true } }" --tls.allowInsecure --sqlBackend="{ \"dbEngine\": \"postgres_tcp\", \"sqlDialect\": \"postgres\", \"dsn\": \"postgres://stackql:stackql@127.0.0.1:7432/stackql\" }" shell
```
