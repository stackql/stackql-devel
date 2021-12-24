
## Required

- [ ] Identity Providers
- [x] Applications
- [ ] Groups
- [ ] Users
- [ ] Policies

## Golden Path

```sh
OKTA_KEYFILE_PATH="${HOME}/moonlighting/infraql-original/keys/okta-token.txt"

./infraql shell --keyfilepath=${OKTA_KEYFILE_PATH} --keyfiletype=api_key
```

```sql
select id from okta.application.apps where subdomain = 'dev-79923018-admin';

insert into okta.application.apps() where subdomain = 'dev-79923018-admin';


INSERT INTO okta.application.apps(
  data__accessibility,
  data__credentials,
  data__features,
  data__label,
  data__licensing,
  data__profile,
  data__settings,
  data__signOnMode,
  data__visibility
)
SELECT
  '{ "errorRedirectUrl": "{{ .values.data__accessibility.errorRedirectUrl }}", "loginRedirectUrl": "{{ .values.data__accessibility.loginRedirectUrl }}", "selfService": {{ .values.data__accessibility.selfService }} }',
  '{ "signing": { "kid": "{{ .values.data__credentials.signing.kid }}", "lastRotated": "{{ .values.data__credentials.signing.lastRotated }}", "nextRotation": "{{ .values.data__credentials.signing.nextRotation }}", "rotationMode": "{{ .values.data__credentials.signing.rotationMode }}", "use": "{{ .values.data__credentials.signing.use }}" }, "userNameTemplate": { "suffix": "{{ .values.data__credentials.userNameTemplate.suffix }}", "template": "{{ .values.data__credentials.userNameTemplate.template }}", "type": "{{ .values.data__credentials.userNameTemplate.type }}" } }',
  '[ "{{ .values.data__features[0] }}" ]',
  '{{ .values.data__label }}',
  '{ "seatCount": {{ .values.data__licensing.seatCount }} }',
  '{ "{{ .values.data__profile[0].key }}": {{ .values.data__profile[0].val }} }',
  '{ "implicitAssignment": {{ .values.data__settings.implicitAssignment }}, "inlineHookId": "{{ .values.data__settings.inlineHookId }}", "notes": { "admin": "{{ .values.data__settings.notes.admin }}", "enduser": "{{ .values.data__settings.notes.enduser }}" }, "notifications": { "vpn": { "helpUrl": "{{ .values.data__settings.notifications.vpn.helpUrl }}", "message": "{{ .values.data__settings.notifications.vpn.message }}", "network": { "connection": "{{ .values.data__settings.notifications.vpn.network.connection }}" } } } }',
  '{{ .values.data__signOnMode }}',
  '{ "appLinks": { "{{ .values.data__visibility.appLinks[0].key }}": {{ .values.data__visibility.appLinks[0].val }} }, "autoLaunch": {{ .values.data__visibility.autoLaunch }}, "autoSubmitToolbar": {{ .values.data__visibility.autoSubmitToolbar }}, "hide": { "iOS": {{ .values.data__visibility.hide.iOS }}, "web": {{ .values.data__visibility.hide.web }} } }'
;

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