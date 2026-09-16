# 🚧⚠️ Preview Queries ⚠️🚧

> [!CAUTION]
> ## 🚨🧪 PREVIEW ONLY: FUNDAMENTALLY UNSTABLE 🧪🚨
>
> 🔥 **Every capability on this page is a preview.** 🔥
>
> - 💥 Relation names, predicates, wire shapes and output columns **can change or disappear without notice**.
> - 🚫 **No compatibility guarantee** between releases. Do **not** build on these.
> - ☢️ Some examples **create real cloud resources**, and a failed run can leave them behind.
> - 🧨 Treat all of it as **unstable**, whether or not `--preview '{"unstable":true}'` is on the command line.

Each example below is a `stackql` call followed by the `omnisdk` facade methods it
reaches. The routing lives in [`internal/stackql/intrinsic`](/internal/stackql/intrinsic).


## Setup

Run from the repo root.

```bash
python cicd/python/build.py --build

# ✅ recommended for every smoke test: exports all of the credentials below
source cicd/vol/vendor-secrets/secrets.sh

# or set individually; omnisdk reads these directly
export AWS_ACCESS_KEY_ID=...  AWS_SECRET_ACCESS_KEY=...          # AWS_SESSION_TOKEN if needed
export AZURE_TENANT_ID=...  AZURE_CLIENT_ID=...  AZURE_CLIENT_SECRET=...
export GOOGLE_CREDENTIALS="$(cat /path/to/sa-key.json)"          # PKCS8 key; or GOOGLE_APPLICATION_CREDENTIALS=/path/to/sa-key.json

# documents for the unstable providers, graph and IaC sections
./build/stackql exec "REGISTRY PULL github v26.08.00448; REGISTRY PULL aws v26.08.00444; REGISTRY PULL google v26.08.00446;"
```


## Precanned streaming audit

Relations under `stackql_preview.audit`; no `--preview` flag needed.

Rules:

- Rows never reach the SQL backend, so `ORDER BY`, `GROUP BY`, `HAVING`, `DISTINCT`, `LIMIT` and aggregates or functions in the select list are refused.
- Only equality predicates are applied; they become `Args.Params`.
- `-o jsonl` writes each row as it arrives. Row order is not deterministic.
- Where more than one method's required params are satisfied, name one with `method = '<name>'`.

Facade methods, in call order, for every example ([`omnisdk.go`](/internal/stackql/intrinsic/omnisdk.go)):

- `omnisdk.Default().Resources(".*")`, to resolve the relation name to a resource path.
- `omnisdk.Default().Methods(resource.Path)`, to pick the method.
- `omnisdk.Default().New(method.Path, omnisdk.Args{...})`.
- `Plan.Open(ctx)`, whose `Rows` stream back as the result set.

Catalogue:

```bash
./build/stackql exec "SHOW RESOURCES IN stackql_preview.audit;"                  # Default().Resources(".*")
./build/stackql exec "DESCRIBE stackql_preview.audit.aws_s3_buckets;"            # Default().Resources(".*")
./build/stackql exec "SHOW METHODS IN stackql_preview.audit.aws_s3_buckets;"     # Default().Resources(".*"), Default().Methods(path)
```

### Scope: whole org, or bounded

Grant read-only roles at the highest scope you want swept. The roles per cloud are in [Foreign auth patterns](/docs/auth.md#foreign-auth-patterns) (`SecurityAudit` for AWS; `Reader` at subscription or management group for Azure; `Viewer`, `Security Reviewer` and `Folder Viewer` at org for GCP).

| Cloud | Whole org | Bounded |
| --- | --- | --- |
| 🟧 AWS | ❌ Not built. One account per run: the credentials' account. For several accounts, run once per account with that account's (e.g. STS assumed-role) credentials. | Credentials of the one account. `region` is the signing region; S3 and IAM listings are account-wide. |
| 🟦 Azure | Every subscription the service principal can read. Assign `Reader` at the tenant root management group. | Assign `Reader` only on the management group or subscriptions to sweep. There is no scope predicate. |
| 🟦 Entra | The tenant the credentials belong to. | `AZURE_TENANT_ID`. |
| 🟩 GCP | `google_org = '<digits>'`: recursive folder → project descent. | `google_project = '<id>'` for a single project; exactly one of the two is required. `SHOW METHODS` lists `project` / `org` for `google_storage_buckets`, but the runtime wants the `google_`-prefixed names. |
| 🌐 omni | AWS, Azure and GCP legs each as wide as the row above allows. | Bound each leg as above. |

One example per relation. The comment is the method path handed to `Default().New`.

```bash
# aws.s3.buckets.list
./build/stackql exec -o jsonl "SELECT * FROM stackql_preview.audit.aws_s3_buckets WHERE region = 'ap-southeast-2' AND method = 'list';"

# aws.s3.buckets.enumerate
./build/stackql exec -o jsonl "SELECT * FROM stackql_preview.audit.aws_s3_buckets WHERE region = 'ap-southeast-2' AND method = 'enumerate';"

# aws.s3.buckets.encryption
./build/stackql exec -o jsonl "SELECT * FROM stackql_preview.audit.aws_s3_buckets WHERE region = 'ap-southeast-2' AND method = 'encryption';"

# aws.iam.principals.list  (IAM is global; region only signs the request)
./build/stackql exec -o jsonl "SELECT * FROM stackql_preview.audit.aws_iam_principals WHERE region = 'us-east-1' AND method = 'list';"

# aws.iam.principals.access
./build/stackql exec -o jsonl "SELECT * FROM stackql_preview.audit.aws_iam_principals WHERE region = 'us-east-1' AND method = 'access';"

# azure.storage.containers.list
./build/stackql exec -o jsonl "SELECT * FROM stackql_preview.audit.azure_storage_containers;"

# azure.network.subnets.list
./build/stackql exec -o jsonl "SELECT * FROM stackql_preview.audit.azure_network_subnets;"

# entra.identities.list
./build/stackql exec -o jsonl "SELECT * FROM stackql_preview.audit.entra_identities WHERE method = 'list';"

# entra.identities.access
./build/stackql exec -o jsonl "SELECT * FROM stackql_preview.audit.entra_identities WHERE method = 'access';"

# google.storage.buckets.list
./build/stackql exec -o jsonl "SELECT * FROM stackql_preview.audit.google_storage_buckets WHERE google_project = 'stackql-demo';"

# gcp.iam.principals.list
./build/stackql exec -o jsonl "SELECT * FROM stackql_preview.audit.gcp_iam_principals WHERE google_project = 'stackql-demo';"

# omni.storage.buckets.list
./build/stackql exec -o jsonl "SELECT * FROM stackql_preview.audit.omni_storage_buckets WHERE region = 'ap-southeast-2' AND google_project = 'stackql-demo';"

# omni.iam.principals.list  (IAM is global; region only signs the AWS leg)
./build/stackql exec -o jsonl "SELECT * FROM stackql_preview.audit.omni_iam_principals WHERE region = 'us-east-1' AND google_project = 'stackql-demo' AND method = 'list';"

# omni.iam.principals.access
./build/stackql exec -o jsonl "SELECT * FROM stackql_preview.audit.omni_iam_principals WHERE region = 'us-east-1' AND google_project = 'stackql-demo' AND method = 'access';"
```

⚠️ These two are `SELECT` in syntax only. Opening the cursor runs a provisioning flow, so they **create real cloud resources** (a VPC and a subnet) and the rows are the report:

```bash
# aws.ec2.networks.provision
./build/stackql exec -o jsonl "SELECT * FROM stackql_preview.audit.aws_ec2_networks WHERE region = 'ap-southeast-2' AND vpc_cidr = '10.42.0.0/16' AND subnet_cidr = '10.42.1.0/24';"

# gcp.compute.networks.provision  (region defaults to us-central1)
./build/stackql exec -o jsonl "SELECT * FROM stackql_preview.audit.gcp_compute_networks WHERE project = 'stackql-demo';"
```

Refused, to check the rules:

```bash
./build/stackql exec "SELECT * FROM stackql_preview.audit.azure_storage_containers ORDER BY name;"
./build/stackql exec "SELECT count(*) FROM stackql_preview.audit.azure_storage_containers;"
```


## Single document exchange

Relations named `stackql_unstable_<provider>.<service>.<resource>`, read straight from the pulled provider document.

```bash
./build/stackql exec --preview '{"unstable":true}' -o jsonl \
  "SELECT login, id, type FROM stackql_unstable_github.orgs.members WHERE org = 'stackql';"
```

Facade methods, in call order ([`doc.go`](/internal/stackql/intrinsic/doc.go)):

- `omnisdk.NewFromCatalog(docDir, "stackql_unstable_github.orgs.members", omnisdk.Args{...})`, where `docDir` is the bundle's latest versioned directory under the local doc root.
- `Plan.Open(ctx)`, whose `Rows` stream back as the result set.

Catalogue:

```bash
./build/stackql exec --preview '{"unstable":true}' "SHOW SERVICES IN stackql_unstable_github;"               # DocCatalog(docDir, bundle)
./build/stackql exec --preview '{"unstable":true}' "SHOW RESOURCES IN stackql_unstable_github.orgs;"         # DocResources(docDir, bundle, service)
./build/stackql exec --preview '{"unstable":true}' "SHOW METHODS IN stackql_unstable_github.orgs.members;"   # DocMethods(docDir, bundle, service, resource)
```


## Bindings across exchanges

```bash
./build/stackql exec --preview '{"unstable":true}' -o jsonl "$(cat <<'SQL'
SELECT * FROM stackql_preview.dynamic_graph.query
WHERE region = 'ap-southeast-2'
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

Facade methods, in call order ([`dynamic.go`](/internal/stackql/intrinsic/dynamic.go)):

- `omnisdk.NewInbound(from, src, as)`, once per `inbound` entry.
- `omnisdk.NewWiring(to, inbound, viaType, viaProgram, provides...)`, once per `wirings` entry.
- `omnisdk.NewOverride(address, objectKey, mediaType, programType, programBody)`, once per `overrides` entry.
- `omnisdk.NewGraph(addresses, wirings, overrides...)`.
- `omnisdk.NewGraphQuery(registryRoot, graph, omnisdk.Args{...})`, where `Args.Params` holds the predicates other than `spec`.
- `Plan.Open(ctx)`, whose `Rows` stream back as the result set.


## IAC with exchanges

Both examples **create a real VPC and subnet**.

```bash
./build/stackql exec --preview '{"unstable":true}' "$(cat <<'SQL'
SELECT * FROM stackql_preview.iac.run
WHERE collection = 'scratch'
  AND state = 'cicd/work/iac-state'
  AND region = 'ap-southeast-2'
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

Facade methods, in call order ([`iac.go`](/internal/stackql/intrinsic/iac.go)):

- `omnisdk.NewResource(key, provider, address, desired, params, arrivals, viaType, viaProgram, identity, addressedBy, correlationParam)`, once per `resources` entry, with each `inbound` entry as an `omnisdk.Arrival{From, As}`.
- `omnisdk.Converge(registryRoot, collection, state, runID, resources, omnisdk.Args{...})`, where `Args.Params` holds the predicates other than `collection`, `resources`, `blueprint`, `state` and `run_id`.
- `Plan.Open(ctx)`, which performs the run; its `Rows` are the report.

From a blueprint:

```bash
./build/stackql exec --preview '{"unstable":true}' \
  "SELECT * FROM stackql_preview.iac.run
   WHERE collection = 'scratch' AND blueprint = 'aws-vpc-subnet'
     AND region = 'ap-southeast-2' AND vpc_cidr = '10.42.0.0/16'
     AND subnet_cidr = '10.42.1.0/24';"
```

- `omnisdk.BlueprintFor(handle)`, falling back to matching `Blueprint.Handle()` across `omnisdk.Blueprints()`.
- `Blueprint.Params()`, to select the predicates the blueprint declares as inputs.
- `Blueprint.Resources(inputs)`.
- `omnisdk.Converge(registryRoot, collection, state, runID, resources, omnisdk.Args{...})`.
- `Plan.Open(ctx)`.

Catalogue:

```bash
./build/stackql exec --preview '{"unstable":true}' "SHOW RESOURCES IN stackql_preview.iac;"               # Blueprints(), Handle(), Summary()
./build/stackql exec --preview '{"unstable":true}' "DESCRIBE stackql_preview.iac.aws_vpc_subnet;"         # BlueprintFor(handle), Params()
```
