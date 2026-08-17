# omnisdk: document-derived plans do not bind caller inputs to the request

Observed against `github.com/stackql-labs/omnisdk v0.1.1-alpha02`.

## Summary

`NewFromCatalog` accepts `Args.Params`, validates them, and then builds a request
that ignores them. Caller inputs never reach the URL, so every document-derived
`SELECT` that needs a parameter fails at the provider.

Two independent gaps produce this:

1. **`{name}` placeholders in the URL template are never substituted**, though
   `aot.Request.URL()` documents exactly that contract.
2. **Query parameters cannot be expressed at all** — `aot.Request.Params()` is
   defined as body parameters, and there is no `in: query` representation.

The hand-authored catalog does both itself (`gcpIAMBindingsSpec` uses
`/projects/{project}:getIamPolicy`, `awsPerUserSpec` binds `"UserName":
"{principal}"`), so this is specific to the document path.

## Reproduction

Four lines, no consumer involved. `bundleDir` is any stackql provider bundle —
a directory holding `provider.yaml` and `services/`.

```go
pl, err := omnisdk.NewFromCatalog(
    bundleDir, // e.g. ~/.stackql/src/googleapis.com/v25.12.00357
    "stackql_unstable_google.storage.buckets",
    omnisdk.Args{Params: map[string]string{"project": "stackql-demo"}},
)
rows, err := pl.Open(context.Background())
for rows.Next() {
}
// rows.Err()
```

### Actual

```
GET https://storage.googleapis.com/storage/v1/b/
400 {"error":{"code":400,"message":"Required parameter: project",
     "errors":[{"reason":"required","locationType":"parameter","location":"project"}]}}
```

Note the trailing `/b/`: an empty path segment where a placeholder was left
unbound. The `project` value is absent from both path and query.

### Expected

```
GET https://storage.googleapis.com/storage/v1/b?project=stackql-demo
```

## Where it goes wrong

`pkg/omnisdk/omnisdk.go` — `docInputs` copies the caller's params faithfully:

```go
func docInputs(args Args) map[string]any {
	out := map[string]any{}
	for k, v := range args.Params {
		out[k] = v
	}
	return out
}
```

`internal/system_g/exchange/docx/docx.go` — `PlanFor` uses that map for two
things only: presence validation, and as bound values on the plan. Neither
reaches request construction.

```go
for _, in := range ex.Inputs() {
	if v, ok := inputs[in]; !ok || v == "" {
		return nil, fmt.Errorf("docx: exchange %q requires input %q", ex.Name(), in)
	}
}
return plan.NewPlan([]plan.ExchangeSpec{spec}, nil, nil, inputs, nil, encoder.NewJSONLEncoder())
```

`Spec` builds the request from the declaration alone, and folds
`req.Params()` into the **body** regardless of where the document says those
parameters belong:

```go
url := retarget(req.URL(), o.baseURL)          // template, never substituted
hreq := httpx.Request{Method: req.Method(), URL: url}
if params := req.Params(); len(params) > 0 {
	body := make(map[string]any, len(params))
	for k, v := range params {
		body[k] = v
	}
	hreq.Body = httpx.Body{Encoding: encodingOf(req.MediaType()), Params: body}
}
```

`Spec` has no access to `inputs` at all — it takes only `(ex, reg, opts...)`.

## What appears to be needed

1. **Substitute `{name}` from inputs into `Request.URL()`.** This is already the
   documented contract:

   > `// URL is a template; {name} placeholders are bound from Inputs.`

   It requires giving `Spec` the inputs, or binding at plan time before the
   request is issued.

2. **Represent parameter location in the AOT model.** `Request.Params()` is
   documented as "the body parameters the document declares", so a
   query parameter has nowhere to live. Source documents distinguish
   `in: query` / `in: path` / `in: header` / body, and that distinction is lost
   before `docx` sees it. Without it, `?project=` cannot be emitted no matter
   what `docx` does.

Point 2 is the blocking one for `google.storage.buckets`, whose `project` is
`in: query`. Point 1 blocks any resource whose path carries a placeholder,
which is most of them.

## Consumer context

stackql exposes these as `stackql_unstable_<provider>` relations. `SHOW
SERVICES`, `SHOW RESOURCES` and `SHOW METHODS` work against real bundles today —
`DocCatalog`, `DocResources` and `DocMethods` are unaffected. Only the
`SELECT` path is blocked, and the fix is entirely within omnisdk: stackql
already extracts equality predicates from the `WHERE` clause and hands them over
as `Args.Params`.
