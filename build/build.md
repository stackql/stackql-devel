
## Build Directory

This directory is the location for building and testing artifacts.  Various walkthroughs and tests depend upon the executable being pre-built to this directory.



## Active defects

## Multi Dependent nonsense

This query:

```sql
select
keyz.name as key_name, 
keyz.tags as key_tags, 
json_extract(detail.properties, '$.kty') as key_class, 
json_extract(detail.properties, '$.keySize') as key_size, 
json_extract(detail.properties, '$.keyOps') as key_ops, 
keyz.type as key_type 
from azure.key_vault.vaults vaultz 
inner join azure.key_vault.keys keyz 
on keyz.vaultName = split_part(vaultz.id, '/', -1) 
and keyz.subscriptionId = vaultz.subscriptionId 
and keyz.resourceGroupName = split_part(vaultz.id, '/', 5) 
inner join azure.key_vault.keys detail 
on detail.vaultName = split_part(vaultz.id, '/', -1) 
and detail.subscriptionId = '000000-0000-0000-0000-000000000011' 
and detail.resourceGroupName = split_part(vaultz.id, '/', 5) 
and detail.keyName = split_part(keyz.id, '/', -1) 
where vaultz.subscriptionId = '000000-0000-0000-0000-000000000011' 
;
```

Returns this (simulated) result:

```
|--------------|----------|-----------|----------|-------------------------------------------------------------|--------------------------------|
|   key_name   | key_tags | key_class | key_size |                           key_ops                           |            key_type            |
|--------------|----------|-----------|----------|-------------------------------------------------------------|--------------------------------|
| dummy-key-01 | {}       | RSA       |     2048 | ["sign","verify","wrapKey","unwrapKey","encrypt","decrypt"] | Microsoft.KeyVault/vaults/keys |
|--------------|----------|-----------|----------|-------------------------------------------------------------|--------------------------------|
| dummy-key-01 | {}       | RSA       |     2048 | ["sign","verify","wrapKey","unwrapKey","encrypt","decrypt"] | Microsoft.KeyVault/vaults/keys |
|--------------|----------|-----------|----------|-------------------------------------------------------------|--------------------------------|
| dummy-key-01 | {}       | RSA       |     2048 | ["sign","verify","wrapKey","unwrapKey","encrypt","decrypt"] | Microsoft.KeyVault/vaults/keys |
|--------------|----------|-----------|----------|-------------------------------------------------------------|--------------------------------|
| dummy-key-01 | {}       | RSA       |     2048 | ["sign","verify","wrapKey","unwrapKey","encrypt","decrypt"] | Microsoft.KeyVault/vaults/keys |
|--------------|----------|-----------|----------|-------------------------------------------------------------|--------------------------------|
| dummy-key-01 | {}       | RSA       |     2048 | ["sign","verify","wrapKey","unwrapKey","encrypt","decrypt"] | Microsoft.KeyVault/vaults/keys |
|--------------|----------|-----------|----------|-------------------------------------------------------------|--------------------------------|
| dummy-key-01 | {}       | RSA       |     2048 | ["sign","verify","wrapKey","unwrapKey","encrypt","decrypt"] | Microsoft.KeyVault/vaults/keys |
|--------------|----------|-----------|----------|-------------------------------------------------------------|--------------------------------|
| dummy-key-01 | {}       | RSA       |     2048 | ["sign","verify","wrapKey","unwrapKey","encrypt","decrypt"] | Microsoft.KeyVault/vaults/keys |
|--------------|----------|-----------|----------|-------------------------------------------------------------|--------------------------------|
| dummy-key-01 | {}       | RSA       |     2048 | ["sign","verify","wrapKey","unwrapKey","encrypt","decrypt"] | Microsoft.KeyVault/vaults/keys |
|--------------|----------|-----------|----------|-------------------------------------------------------------|--------------------------------|
```

## Already fixed bug scenario


```
stackql  >>select * from aws.ec2.vpcs_list_only where region in ('ap-southeast-1', 'ap-southeast-2');
|----------------|-----------------------|
|     region     |        vpc_id         |
|----------------|-----------------------|
| ap-southeast-2 | vpc-aaaaaaaaa00000001 |                                                                                                
|----------------|-----------------------|                                                                                                
| ap-southeast-1 | vpc-aaaaaaaaa00000002 |                                                                                                
|----------------|-----------------------|                                                                                                
| ap-southeast-1 | vpc-aaaaaaaaa00000003 |                                                                                                
|----------------|-----------------------|                                                                                                
stackql  >>select * from aws.ec2.vpcs_list_only where region = 'ap-southeast-1';
|----------------|-----------------------|                                                                                                
|     region     |        vpc_id         |                                                                                                
|----------------|-----------------------|                                                                                                
| ap-southeast-1 | vpc-aaaaaaaaa00000002 |                                                                                                
|----------------|-----------------------|                                                                                                
| ap-southeast-1 | vpc-aaaaaaaaa00000003 |                                                                                                
|----------------|-----------------------|                                                                                                
stackql  >>select tag_key, tag_value from aws.ec2.vpc_tags where region = 'ap-southeast-1';
|-------------|------------|                                                                                                              
|   tag_key   | tag_value  |                                                                                                              
|-------------|------------|                                                                                                              
| Provisioner | stackql    |                                                                                                              
|-------------|------------|                                                                                                              
| StackEnv    | dev        |                                                                                                              
|-------------|------------|                                                                                                              
| Name        | test       |                                                                                                              
|-------------|------------|                                                                                                              
| StackName   | test-stack |                                                                                                              
|-------------|------------|                                                                                                              
| Name        | legacy     |                                                                                                              
|-------------|------------|                                                                                                              
stackql  >>select * from aws.ec2.vpcs where region = 'ap-southeast-1';
Error: Recovered in HandlePanic(): runtime error: invalid memory address or nil pointer dereference                                       
stackql  >>exit

```

## No control counters present in streams

The selection contexts formed for sql streaming queries lack control counters.

Do not have a convenient test to hand.


## Control counter invariants

- What are they??!!
- Enforce strictly.
- Document.

