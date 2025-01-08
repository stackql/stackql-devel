
## Background

In this walkthrough, we go through the setup of a webserver using `stackql`.  This is useful in itself for development purposes, and we will build on it in more complex examples.

This walkthrough is not at all original; it is an amalgam of materials freely (and redundantly) available elsewehere.  It is heavily inspired by:

- [apache2 config file documentation](https://httpd.apache.org/docs/2.4/configuring.html).  On `ubuntu`, the root config file is `/etc/apache2/apache2.conf`.
- [apache2 documentation on TLS](https://httpd.apache.org/docs/2.4/ssl/ssl_howto.html).

## Setup

The project to be used requires the following google APIs to be activated:

- `compute`.
- `sqladmin`.
- `servicenetworking`.

First, create a google service account key using the GCP Console, per [the GCP documentation](https://cloud.google.com/iam/docs/keys-create-delete).  Grant the service account at least requisite compute and firewall mutation privileges, per [the GCP documentation](https://cloud.google.com/iam/docs/create-service-agents#grant-roles);  corresponding to [this flask deployment example](https://cloud.google.com/docs/terraform/deploy-flask-web-server#permissions):


>  - `compute.instances.*`
>  - `compute.firewalls.*`

Then, do this in bash:

```bash setup stackql-shell credentials_path=cicd/keys/testing/google-rw-credentials.json app_root_path=./test/tmp/.create-google-vm-webserver.stackql

export GOOGLE_CREDENTIALS="$(cat <<credentials_path>>)";

stackql shell --approot=<<app_root_path>>
```

## Method

Do this in the `stackql` shell, replacing `<<project>>` with your GCP project name, '<<region>>', and `<<zone>>` as desired, eg: `australia-southeast1-a`:

```sql stackql-shell input required my_ephemeral_network_name=my-ephemeral-network-tls-01 my_vm_name=my-ephemeral-vm-tls-01 project=stackql-demo region=australia-southeast2 zone=australia-southeast2-a fw_name=ephemeral-https-01 google_provider_version=v24.11.00274

registry pull google <<google_provider_version>>;

insert /*+ AWAIT */ into 
google.compute.networks (
  project,
  data__name,
  data__autoCreateSubnetworks
)
select
  '<<project>>',
  '<<my_ephemeral_network_name>>',
  true
;

insert /*+ AWAIT */ into 
google.compute.addresses (
  project,
  data__name,
  data__autoCreateSubnetworks
)
select
  '<<project>>',
  '<<my_ephemeral_network_name>>',
  true
;

insert /*+ AWAIT */ into 
google.sqladmin.instances (
  project,
  data__name,
  data__autoCreateSubnetworks
)
select
  '<<project>>',
  '<<my_sql_instance_name>>',
  true
;

insert /*+ AWAIT */ into 
google.compute.instances (
  project,
  zone,
  data__name,
  data__machineType,
  data__metadata,
  data__networkInterfaces,
  data__disks
)
select 
  '<<project>>',
  '<<zone>>',
  '<<my_vm_name>>',
  'zones/<<zone>>/machineTypes/n1-standard-1',
  '{
    "items": [
      {
        "key": "startup-script",
        "value": "#! /bin/bash\\nsudo apt-get update\\nsudo apt-get -y install apache2\\necho ''<!doctype html><html><body><h1>Hello from stackql auto-provisioned.</h1></body></html>'' | sudo tee /var/www/html/index.html"
      }
    ]
  }',
  '[
      {
        "stackType": "IPV4_ONLY",
        "accessConfigs": [
          {
            "name": "External NAT",
            "type": "ONE_TO_ONE_NAT",
            "networkTier": "PREMIUM"
          }
        ],
        "subnetwork": "projects/<<project>>/regions/<<region>>/subnetworks/<<my_ephemeral_network_name>>"
      }
    ]',
    '[
      {
        "autoDelete": true,
        "boot": true,
        "initializeParams": {
          "diskSizeGb": "10",
          "sourceImage": "https://compute.googleapis.com/compute/v1/projects/ubuntu-os-cloud/global/images/family/ubuntu-2004-lts"
        },
        "mode": "READ_WRITE",
        "type": "PERSISTENT"
      }
    ]'
;

insert /*+ AWAIT */ into 
google.compute.firewalls (
   project,
   data__name,
   data__network,
   data__allowed,
   data__direction,
   data__sourceRanges
)
select
  '<<project>>',
  '<<fw_name>>',
  'global/networks/<<my_ephemeral_network_name>>',
  '[
    {
      "IPProtocol": "tcp",
      "ports": [
        "80",
        "443",
        "22"
      ]
    }
  ]',
  'INGRESS',
  '[
    "0.0.0.0/0"
  ]'
;

```

## Proving ground

```sql
insert /*+ AWAIT */ into google.compute.networks ( project, data__name, data__autoCreateSubnetworks ) select 'stackql-dashboards', 'my-dashboard-nw-01', true ;

insert /*+ AWAIT */ into 
google.compute.global_addresses (
  project,
  data__name,
  data__prefixLength,
  data__purpose,
  data__addressType,
  data__description,
  data__network
)
select
  'stackql-dashboards',
  'private-services-db-01',
  16,
  'VPC_PEERING',
  'INTERNAL',
  'support for private services access',
  'projects/stackql-dashboards/global/networks/my-dashboard-nw-01'
;


-- await logic does not currently work for this interface
insert into 
google.servicenetworking.connections (
  servicesId,
  data__network,
  data__reservedPeeringRanges
)
select
  'servicenetworking.googleapis.com',
  'projects/stackql-dashboards/global/networks/my-dashboard-nw-01',
  '[
    "private-services-db-01"
  ]'
;


sleep 30000;  -- cover for lacking await

insert /*+ AWAIT */ into 
google.sqladmin.instances (
  project,
  data__name,
  data__region,
  data__databaseVersion,
  data__edition,
  data__settings
)
select
  'stackql-dashboards',
  'dashoboards-inst-01',
  'australia-southeast1',
  'POSTGRES_14',
  'ENTERPRISE',
  '{
    "ipConfiguration": {
      "privateNetwork": "projects/stackql-dashboards/global/networks/my-dashboard-nw-01",
      "authorizedNetworks": [
        {
          "value": "0.0.0.0/0"
        }
      ],
      "ipv4Enabled": false  
    },
    "tier": "db-custom-2-7680"
  }'
;




```


```json

{
  "name": "dashoboards-inst-01",
  "databaseVersion": "POSTGRES_14",
  "region": "australia-southeast1",
  "settings": {
    "ipConfiguration": {
      "privateNetwork": "projects/stackql-dashboards/global/networks/my-dashboard-nw-01",
      "authorizedNetworks": [
        {
          "value": "0.0.0.0/0"
          
        }
        
      ],
      "ipv4Enabled": false
      
    },
    "tier": "db-custom-2-7680"
  }
  
}
```

```bash setup credentials_path=cicd/keys/testing/google-rw-credentials.json app_root_path=./test/tmp/.create-google-vm-webserver.stackql my_vm_name=my-ephemeral-vm-tls-01 project=stackql-demo zone=australia-southeast2 zone=australia-southeast2-a

export GOOGLE_CREDENTIALS="$(cat <<credentials_path>>)";

publicIpAddress=$(stackql --approot=<<app_root_path>> exec "select json_extract(\"networkInterfaces\", '\$[0].accessConfigs[0].natIP') as public_ipv4_address from   google.compute.instances where project = '<<project>>' and zone = '<<zone>>' and instance = '<<my_vm_name>>';" -o json | jq -r '.[0].public_ipv4_address')

echo "publicIpAddress=${publicIpAddress}"
result=""
for i in $(seq 1 20); do
  sleep 5;
  result="$(curl http://${publicIpAddress} | grep 'auto-provisioned')";
  if [ "${result}" != "" ]; then
    break
  fi
done

echo "${result}";

```

## Result


You will see exactly this in the output:

```html expectation stdout-contains-all
<!doctype html><html><body><h1>Hello from stackql auto-provisioned.</h1></body></html>
```

## Cleanup

```bash teardown best-effort app_root_path=./test/tmp/.create-google-vm-webserver.stackql credentials_path=cicd/keys/testing/google-rw-credentials.json my_ephemeral_network_name=my-ephemeral-network-tls-01 my_vm_name=my-ephemeral-vm-tls-01 project=stackql-demo region=australia-southeast2 zone=australia-southeast2-a fw_name=ephemeral-https-01

echo "begin teardown";

export GOOGLE_CREDENTIALS="$(cat <<credentials_path>>)";

stackql --approot=<<app_root_path>> exec "delete /*+ AWAIT */ from google.compute.instances where project = '<<project>>' and zone = '<<zone>>' and instance = '<<my_vm_name>>';"

stackql --approot=<<app_root_path>> exec "delete /*+ AWAIT */ from google.compute.firewalls where project = '<<project>>' and firewall= '<<fw_name>>';"

stackql --approot=<<app_root_path>> exec "delete /*+ AWAIT */ from google.compute.networks where project = '<<project>>' and network = '<<my_ephemeral_network_name>>';"

rm -rf <<app_root_path>> ;

echo "conclude teardown";

```
