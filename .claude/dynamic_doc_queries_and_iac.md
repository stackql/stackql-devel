
# Dynamic Queries and IAC in omnisdk

### Calling it from Go

The CLI is a thin consumer of the facade; a client such as stackql imports `pkg/omnisdk` and reaches
the same functions. Both paths return the same `Plan`/`Rows` a single-method query returns, so a
consumer iterates one cursor shape whether it is reading or provisioning.

**Auth and document location behave exactly as elsewhere.** The provider document declares the scheme
— `aws_signing_v4`, `service_account`, `oauth2` — and the credential comes from `Args.Auth`, falling
back to the canonical `AWS_*`, `AZURE_*` and `GOOGLE_*` variables. Only the credential a document
actually declares is required, so a run touching AWS alone does not fail because a Google key
elsewhere is stale; a credential that IS present but unusable says so rather than reporting as
absent. The registry root is a parameter rather than configuration, and is required — which document
set a run resolves against is scope, and scope is never inferred.

**1. `stackql_dynamic`  A dynamic query across provider documents** — the `doc-graph` equivalent:

```go
const vpcs, subnets = "stackql_unstable_aws.ec2.vpcs", "stackql_unstable_aws.ec2.subnets"

g, err := omnisdk.NewGraph(
    []string{vpcs, subnets},
    []omnisdk.Wiring{omnisdk.NewWiring(
        subnets,                                                        // the consumer
        []omnisdk.Inbound{omnisdk.NewInbound(vpcs, "VpcId", "vpc_id")}, // β: src → inbox label
        gotemplate.TypeJSON1,                                           // T_in, or "" for identity
        `{"Filter.1.Name":"vpc-id","Filter.1.Value.1":"{{ .vpc_id }}"}`,
        "Filter.1.Name", "Filter.1.Value.1",                            // what T_in provides
    )},
    // Corrections to what a document says about its response, where it is wrong for this engine:
    // omnisdk.NewOverride(addr, "$.items", "", "", ""),
)
if err != nil {
    return err
}

pl, err := omnisdk.NewGraphQuery(registryRoot, g, omnisdk.Args{
    Auth:   auth,                                        // nil falls back to the env
    Params: map[string]string{"region": "us-east-1"},    // scope
})
rows, err := pl.Open(ctx)
defer rows.Close()
for rows.Next() {
    row := rows.Row() // {"VpcId":…, "SubnetId":…, "CidrBlock":…}
}
return rows.Err()
```

**2. `stackql_iac` An idempotent IaC run** — the `iac-apply` equivalent:

```go
res := []omnisdk.ManagedResource{
    omnisdk.NewResource(
        "aws/ec2/vpc",                   // key within the collection
        "aws", "ec2.vpcs",               // registry provider, document address
        []byte(`{"CidrBlock":"10.42.0.0/16"}`),
        map[string]string{
            "TagSpecification.1.ResourceType": "vpc",
            "TagSpecification.1.Tag.1.Key":    "omnisdk:key",
        },
        nil, "", "",                     // inbound, T_in type, T_in program
        "line_items.VpcId",              // where the minted id sits in the projected response
        "VpcId",                         // the parameter addressing an existing object
        "TagSpecification.1.Tag.1.Value", // the parameter taking the correlation stamp
    ),
    omnisdk.NewResource(
        "aws/ec2/subnet", "aws", "ec2.subnets",
        []byte(`{"CidrBlock":"10.42.1.0/24"}`), nil,
        []omnisdk.Arrival{{From: "aws/ec2/vpc", As: "VpcId"}}, "", "",
        "line_items.SubnetId", "SubnetId", "",
    ),
}

pl, err := omnisdk.Converge(registryRoot, "scratch", stateDir, "" /* runID: timestamp */, res,
    omnisdk.Args{Auth: auth, Params: map[string]string{"region": "us-east-1"}})
if err != nil {
    return err
}
rows, err := pl.Open(ctx)   // the run happens here
```

Rows report `{"key":…, "identity":…, "status":…}`, with a final row carrying `error`,
`compensated` and `outstanding` when a run failed. A non-empty `outstanding` means the run is
*partially* compensated — something it created is still there and could not be removed.

Precanned deployments are reachable the same way, and are only a way of building that slice:

```go
bp, ok := omnisdk.BlueprintFor("aws-vpc-subnet")
res, err := bp.Resources(map[string]string{
    "region": "us-east-1", "vpc_cidr": "10.42.0.0/16", "subnet_cidr": "10.42.1.0/24",
})
```

`omnisdk.Blueprints()` lists them with the inputs each declares, which is what `iac-handles` prints.


## stackql invocations

Let us look at some `omnicli` invocations and their `stackql` equivalents.


### Bindings across exchanges


Here is the example `omnicli` functionality:

```bash
omnicli doc-graph /path/to/registry '{
  "addresses": ["stackql_unstable_aws.ec2.vpcs", "stackql_unstable_aws.ec2.subnets"],
  "wirings": [{
    "to": "stackql_unstable_aws.ec2.subnets",
    "inbound": [{"from": "stackql_unstable_aws.ec2.vpcs", "src": "VpcId", "as": "vpc_id"}],
    "via_type": "golang_template_json_v0.1.0",
    "via": "{\"Filter.1.Name\":\"vpc-id\",\"Filter.1.Value.1\":\"{{ .vpc_id }}\"}",
    "provides": ["Filter.1.Name", "Filter.1.Value.1"]
  }]
}' --aws-region us-east-1
```


Here is the equivalent `stackql` functionality:

```bash
stackql exec --preview '{"unstable":true}' "$(cat <<'SQL'
SELECT * FROM stackql_dynamic.graph.query
WHERE region = 'us-east-1'
  AND spec = '{
    "addresses": ["stackql_unstable_aws.ec2.vpcs", "stackql_unstable_aws.ec2.subnets"],
    "wirings": [{
      "to": "stackql_unstable_aws.ec2.subnets",
      "inbound": [{"from": "stackql_unstable_aws.ec2.vpcs", "src": "VpcId", "as": "vpc_id"}],
      "viaType": "golang_template_json_v0.1.0",
      "viaProgram": "{\\"Filter.1.Name\\":\\"vpc-id\\",\\"Filter.1.Value.1\\":\\"{{ .vpc_id }}\\"}",
      "provides": ["Filter.1.Name", "Filter.1.Value.1"]
    }]
  }';
SQL
)"
```

Three things differ from the `omnicli` form:

- The region is a predicate rather than a flag. Every predicate left after the
  control ones are read becomes `Args.Params`, which is where `--aws-region`
  lands.
- The wiring keys are `viaType` and `viaProgram`, where `omnicli` writes
  `via_type` and `via`. Unknown keys are dropped in silence, so a document
  pasted across unchanged loses its transform rather than failing.
- The backslashes in `viaProgram` are **doubled**. The SQL string literal
  consumes one level of escaping before the JSON parser sees the value, so
  `\"` in the `omnicli` payload must be written `\\"` here.

Credentials are unchanged: the provider document declares the scheme, and the
credential comes from stackql's `--auth` for that provider, falling back to the
canonical `AWS_*`, `AZURE_*` and `GOOGLE_*` variables.

Rows arrive exactly as a single-method select's do. A document declares no
egress schema, so the columns are the ones the first row carries, sorted by
name; `SELECT VpcId, SubnetId` projects over them in the order asked for.


### IAC with exchanges


Here is the example `omnicli` functionality:

```bash
omnicli iac-apply /path/to/registry '{
  "name": "scratch", "state": "cicd/work/iac-state",
  "resources": [
    {"key": "aws/ec2/vpc", "provider": "aws", "address": "ec2.vpcs",
     "desired": {"CidrBlock": "10.42.0.0/16"},
     "params": {"TagSpecification.1.ResourceType": "vpc",
                "TagSpecification.1.Tag.1.Key": "omnisdk:key"},
     "identity": "line_items.VpcId", "addressed_by": "VpcId",
     "correlation_param": "TagSpecification.1.Tag.1.Value"},
    {"key": "aws/ec2/subnet", "provider": "aws", "address": "ec2.subnets",
     "desired": {"CidrBlock": "10.42.1.0/24"},
     "inbound": [{"from": "aws/ec2/vpc", "as": "VpcId"}],
     "identity": "line_items.SubnetId", "addressed_by": "SubnetId"}
  ],
  "args": {"params": {"region": "us-east-1"}}
}' --aws-region us-east-1
```

Here is the equivalent `stackql` functionality:

```bash
stackql exec --preview '{"unstable":true}' "$(cat <<'SQL'
SELECT * FROM stackql_iac.converge.run
WHERE collection = 'scratch'
  AND state = 'cicd/work/iac-state'
  AND region = 'us-east-1'
  AND resources = '[
    {"key": "aws/ec2/vpc", "provider": "aws", "address": "ec2.vpcs",
     "desired": {"CidrBlock": "10.42.0.0/16"},
     "params": {"TagSpecification.1.ResourceType": "vpc",
                "TagSpecification.1.Tag.1.Key": "omnisdk:key"},
     "identity": "line_items.VpcId", "addressedBy": "VpcId",
     "correlationParam": "TagSpecification.1.Tag.1.Value"},
    {"key": "aws/ec2/subnet", "provider": "aws", "address": "ec2.subnets",
     "desired": {"CidrBlock": "10.42.1.0/24"},
     "inbound": [{"from": "aws/ec2/vpc", "as": "VpcId"}],
     "identity": "line_items.SubnetId", "addressedBy": "SubnetId"}
  ]';
SQL
)"
```

The mapping onto `iac-apply` is one predicate per top-level field: `name`
becomes `collection`, `state` stays `state`, and `args.params` becomes the
predicates left over once the control ones are read. `run_id` is accepted too,
and defaults to a UTC timestamp. `state` may be omitted, in which case the
ledger and run journals sit under `<approot>/iac`.

Two key names differ from the `omnicli` payload, and are dropped in silence if
pasted across unchanged: `addressedBy` for `addressed_by`, and
`correlationParam` for `correlation_param`.

A blueprint is the same relation with the resources named rather than stated:

```bash
stackql exec --preview '{"unstable":true}' \
  "SELECT * FROM stackql_iac.converge.run
   WHERE collection = 'scratch' AND blueprint = 'aws-vpc-subnet'
     AND region = 'us-east-1' AND vpc_cidr = '10.42.0.0/16'
     AND subnet_cidr = '10.42.1.0/24';"
```

`SHOW RESOURCES IN stackql_iac.blueprints;` lists the handles and
`DESCRIBE stackql_iac.blueprints.aws_vpc_subnet;` the inputs each declares,
which is what `iac-handles` prints. A handle is written with hyphens, which no
unquoted SQL identifier can carry, so it is addressed as `aws_vpc_subnet` in
the relation name and accepted either way in the predicate.

Opening the cursor performs the run, so the rows are the report: `key`,
`identity` and `status`, with a failed run carrying `error`, `compensated` and
`outstanding` as well. A non-empty `outstanding` means the run is *partially*
compensated - something it created is still there and could not be removed.
Columns are sorted by name, as everywhere the cursor infers them from the first
row. A run whose endpoint refused the connection reports:

```text
|-------------|--------------------------------|---------------------|-----------------------|-------------------------------|
| compensated |             error              |         key         |      outstanding      |            status             |
|-------------|--------------------------------|---------------------|-----------------------|-------------------------------|
| []          | docrun: ... connection refused | scratch/aws/ec2/vpc | [scratch/aws/ec2/vpc] | failed; partially compensated |
|-------------|--------------------------------|---------------------|-----------------------|-------------------------------|
```
