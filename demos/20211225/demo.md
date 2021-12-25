
## Required

- [x] Identity Providers
- [x] Applications
- [x] Groups
- [x] Users
- [x] Policies

## Golden Path

```sh
OKTA_KEYFILE_PATH="${HOME}/moonlighting/infraql-original/keys/okta-token.txt"

./infraql shell --keyfilepath=${OKTA_KEYFILE_PATH} --keyfiletype=api_key
```

```sql
select id from okta.application.apps where subdomain = 'dev-79923018-admin';

insert into okta.application.apps() where subdomain = 'dev-79923018-admin';


select * from okta.application.users where appId = '0oa3cy8k41YM15j4H5d7' and subdomain = 'dev-79923018-admin';


exec /*+ SHOWRESULTS */ okta.application.users.get @appId = '0oa3cy8k41YM15j4H5d7', @userId = '00u3cy8k9ovI1nGq95d7', @subdomain = 'dev-79923018-admin';

select id, JSON_EXTRACT(profile, '$.email') as em  from okta.user.users where subdomain = 'dev-79923018-admin';

exec /*+ SHOWRESULTS */ okta.user.users.get @userId = '00u3cy8k9ovI1nGq95d7', @subdomain = 'dev-79923018-admin';

SELECT id from (exec /*+ SHOWRESULTS */ okta.user.users.list @subdomain = 'dev-79923018-admin');

```

## Theoretical

```sql


INSERT INTO okta.application.apps(
  data__name,
  data__label,
  data__settings,
  data__signOnMode
)
SELECT

  '{{ .values.data__name }}',
  '{{ .values.data__label }}',
  '{ "seatCount": {{ .values.data__licensing.seatCount }} }',
  '{ "{{ .values.data__profile[0].key }}": {{ .values.data__profile[0].val }} }',
  '{ "app": { "requestIntegration": "{{ .values.data__settings.app.requestIntegration }}", "url": "{{ .values.data__settings.app.url }}" } }',
  '{{ .values.data__signOnMode }}'
;
```