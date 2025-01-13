

Useful helps:

- [digitalocean API reference](https://docs.digitalocean.com/reference/api/api-reference/).


```bash

source cicd/keys/testing/digitalocean-rw-stackql.sh

stackql shell --approot=./test/tmp/.create-digitalocean-k8s.stackql --http.log.enabled
```

```sql

registry pull digitalocean v24.11.00274;

insert into 
digitalocean.projects.projects (
    data__name,
    data__purpose,
    data__environment
)
select
  'dashboard-k8s-project-uat',
  'Hosting materials for k8s stackql dashboards',
  'Staging'
;

-- this needs to be captured... '18e7b3f9-46d5-4bc9-81cd-08eb19aad862' in this particular case
select id from digitalocean.projects.projects where name = 'dashboard-k8s-project-uat';

-- GOTTA click-ops this to be default project... for now




-- predicated upon correct default project
insert into 
digitalocean.kubernetes.clusters (
  data__name,
  data__region,
  data__version,
  data__node_pools,
  data__tags
)
select
  'stackql-dashboards-k8s-uat-01',
  'syd1',
  '1.31.1-do.5',
  '[
    {
    "size": "s-1vcpu-2gb",
    "count": 2,
    "name": "worker-pool"
    }
  ]',
  '[
    "stackql-dashboards-k8s-uat-01",
    "stackql-dashboards",
    "uat"
  ]'
;


-- here, wanna capture "id" for subsequent... 'b46e7d4f-da8b-4d7f-972c-544f2a37bf7d' ... can actually loop on this until ready -> may take ages
select name, id from digitalocean.kubernetes.clusters where name = 'stackql-dashboards-k8s-uat-01';


select * from digitalocean.kubernetes.clusters_credentials where cluster_id = 'b46e7d4f-da8b-4d7f-972c-544f2a37bf7d';

```

```bash

_clusterId=$(stackql exec --approot=./test/tmp/.create-digitalocean-k8s.stackql -o json "select name, id from digitalocean.kubernetes.clusters where name = 'stackql-dashboards-k8s-uat-01';" | jq -r '.[0].id')

_clusterAuthObj=$(stackql exec --approot=./test/tmp/.create-digitalocean-k8s.stackql -o json "select * from digitalocean.kubernetes.clusters_credentials where cluster_id = '${_clusterId}';" | jq -r '.[0]')

echo "cluster auth data = ${_clusterAuthObj}"

curl -H "Authorization: Bearer ${DIGITALOCEAN_TOKEN}" -X GET https://api.digitalocean.com/v2/kubernetes/clusters/${_clusterId}/kubeconfig > $(pwd)/test/tmp/kubeconfig/do-uat-kubeconfig.yaml


kubectl --kubeconfig=$(pwd)/test/tmp/kubeconfig/do-uat-kubeconfig.yaml get ns
```

## Notes

- `describe digitalocean.projects.projects;` returns only `column_anon`; but decent selct queries actually work.
    - Same also true for `describe digitalocean.droplets.droplets;`; probably endemic to the provider.
- Pagination does not currently work, and appears to revolve around response body `"links":{"pages":{"last":"https://api.digitalocean.com/v2/images?page=14","next":"https://api.digitalocean.com/v2/images?page=2"}}`.
- `stackql  >>update digitalocean.projects.projects set is_default = true where project_id = 'c9027ad4-8e8a-47e8-9cb0-d9f0ae4d4b74';` -> `update statement RHS of type 'sqlparser.BoolVal' not yet supported`.
- Debugging droplet init `cat /var/log/cloud-init-output.log`.
- Viewing droplet startup script `cat /var/lib/cloud/instance/scripts/part-001`.
- Apache2 service status `sudo systemctl status apache2`.