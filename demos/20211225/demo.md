
```sh
OKTA_KEYFILE_PATH="${HOME}/moonlighting/infraql-original/keys/okta-token.txt"

./infraql shell --keyfilepath=${OKTA_KEYFILE_PATH} --keyfiletype=api_key
```

```sql
select * from okta.application.apps where subdomain = 'dev-79923018-admin';

insert into okta.application.apps() where subdomain = 'dev-79923018-admin';

```