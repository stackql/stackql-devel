

```sql

registry pull digitalocean v24.11.00274;

insert into 
digitalocean.projects.projects (
    data__name,
    data__purpose,
    data__environment
)
select
  'dashboard-project-uat',
  'Hosting materials for stackql dashboards',
  'Staging'
;

-- this needs to be captured... 'c9027ad4-8e8a-47e8-9cb0-d9f0ae4d4b74' in this particular case
select id from digitalocean.projects.projects where name = 'dashboard-project-uat';

-- GOTTA click-ops this to be default project... for now





insert into 
digitalocean.droplets.droplets (
  data__name,
  data__region,
  data__image,
  data__size,
  data__user_data
)
select
  'test-04.uat.app.stackql.io',
  'syd',
  'ubuntu-22-04-x64',
  's-1vcpu-1gb',
  '#!/bin/bash\nexport DEBIAN_FRONTEND=noninteractive\nsudo apt-get update\nsudo apt-get -y install apache2\necho ''<!doctype html><html><body><h1>Hello from stackql droplet auto-provisioned.</h1></body></html>'' | sudo tee /var/www/html/index.html'
;



```

## Notes

- `describe digitalocean.projects.projects;` returns only `column_anon`; but decent selct queries actually work.
    - Same also true for `describe digitalocean.droplets.droplets;`; probably endemic to the provider.
- Pagination does not currently work, and appears to revolve around response body `"links":{"pages":{"last":"https://api.digitalocean.com/v2/images?page=14","next":"https://api.digitalocean.com/v2/images?page=2"}}`.
- `stackql  >>update digitalocean.projects.projects set is_default = true where project_id = 'c9027ad4-8e8a-47e8-9cb0-d9f0ae4d4b74';` -> `update statement RHS of type 'sqlparser.BoolVal' not yet supported`.
- Debugging droplet init `cat /var/log/cloud-init-output.log`.
- Apache2 service status `sudo systemctl status apache2`.