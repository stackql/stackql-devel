SELECT 
res.vpc_id, res.region, res.instance_tenancy, res.ipv4_netmask_length, res.cidr_block_associations, res.cidr_block, res.ipv4_ipam_pool_id, res.default_network_acl, res.enable_dns_support, res.ipv6_cidr_blocks, res.default_security_group, res.enable_dns_hostnames, res.tags 
FROM (
    select 
      SPLIT_PART(ResourceARN, '/', -1) as id 
    from awscc.tagging.tagged_resources, 
    json_each(tags) where region = 'us-east-1' and ResourceTypeFilters = '["ec2:vpc"]' and TagFilters = '[{"Key": "Name", "Values": ["databricks-WorkerEnvId(workerenv-7474653953725801-187618cb-7378-439e-9351-01af6d230505)"]}]'
    ) ids 
LEFT JOIN awscc.ec2.vpcs res ON ids.id = res.Identifier 
WHERE region = 'us-east-1';


---

SELECT 
    res.vpc_id, 
    res.region, 
    res.instance_tenancy, 
    res.ipv4_netmask_length, 
    res.cidr_block_associations, 
    res.cidr_block, 
    res.ipv4_ipam_pool_id, 
    res.default_network_acl, 
    res.enable_dns_support, 
    res.ipv6_cidr_blocks, 
    res.default_security_group, 
    res.enable_dns_hostnames, 
    res.tagz 
FROM (
    select 
      SPLIT_PART(tr.ResourceARN, '/', -1) as id,
      tags_split.value as tagz
    from awscc.tagging.tagged_resources tr, 
    json_each(tr.tags) as tags_split
    where 
      tr.region = 'us-east-1' 
      and tr.ResourceTypeFilters = '["ec2:vpc"]' 
      and tr.TagFilters = '[{"Key": "Name", "Values": ["databricks-WorkerEnvId(workerenv-7474653953725801-187618cb-7378-439e-9351-01af6d230505)"]}]'

    ) ids 
LEFT JOIN awscc.ec2.vpcs res ON ids.id = res.Identifier 
WHERE region = 'us-east-1'
;

---

select 
    SPLIT_PART(tr.ResourceARN, '/', -1) as id,
    tags_split.value as tagz,
    res.vpcId
from awscc.tagging.tagged_resources tr
JOIN aws.ec2_native.vpcs res 
ON SPLIT_PART(tr.ResourceARN, '/', -1) = res.vpcId
    and tr.region = 'us-east-1' 
    and res.region = 'us-east-1'
        ,
json_each(tr.tags) as tags_split
where region = 'us-east-1'
    and tr.ResourceTypeFilters = '["ec2:vpc"]' 
    and tr.TagFilters = '[{"Key": "Name", "Values": ["databricks-WorkerEnvId(workerenv-7474653953725801-187618cb-7378-439e-9351-01af6d230505)"]}]'

---


SELECT
id,
create_time,
description,
tag_key,
update_time
FROM databricks_workspace.tags.tag_policies
WHERE deployment_name = 'dbc-74aa95f7-8c7e';