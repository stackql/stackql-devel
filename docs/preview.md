
# Preview functionality

Bleeding edge `stackql` functionality can be exposed using the `--preview` CLI argument.


```bash
./build/stackql shell --preview='{"unstable":true}'
```

For any provider you wish to consume, you will need to pull, for example in support of what follows:

```sql
registry pull google v26.08.00446;

```

## Relevant preview functionality

Streaming high volume queries at low latency has releveance for audit and related use cases.


```sql

./build/stackql exec --preview='{"unstable":true}' --output jsonl \
"select name, location, storageClass, timeCreated
   from stackql_unstable_google.storage.buckets
   where project = 'stackql-demo';"

```