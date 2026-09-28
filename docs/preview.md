
# Preview functionality

Bleeding edge `stackql` functionality can be exposed using the `--preview` CLI argument.


```bash
./build/stackql shell --preview='{"unstable":true}'
```

## Relevant preview functionality

Streaming high volume queries at low latency has releveance for audit and related use cases.


```bash

./build/stackql exec --preview='{"unstable":true}' --output=jsonl \
  "select name, location, storageClass, timeCreated
   from stackql_unstable_google.storage.buckets
   where project = 'stackql-demo'"

```