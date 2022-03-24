
Setup:

```bash
export REG_STR='{ "url": "file:///Users/admin/stackql/stackql-devel/docs/../test/registry", "localDocRoot": "/Users/admin/stackql/stackql-devel/docs/../test/registry", "verifyConfig": {"nopVerify": true } }'

export OKTA_SECRET_KEY=$(cat /Users/admin/stackql/stackql-devel/keys/okta-token.txt)

export AUTH_STR='{ "google": { "credentialsfilepath": "/Users/admin/stackql/stackql-devel/keys/sa-key.json", "type": "service_account" }, "okta": { "credentialsenvvar": "OKTA_SECRET_KEY", "type": "api_key" } }'


$(pwd)/../build/stackql --auth="${AUTH_STR}" --registry="${REG_STR}" shell
```


All of these ought to work:


```sql
select d1.name, d1.id from google.compute.disks d1 inner join google.compute.disks d2 on d1.id = d2.id where d1.project = 'lab-kr-network-01' and d1.zone = 'australia-southeast1-a' and d2.project = 'lab-kr-network-01' and d2.zone = 'australia-southeast1-a' ;

-- Error: Recovered in HandlePanic(): interface conversion: interface is nil, not sqlparser.SQLNode



```