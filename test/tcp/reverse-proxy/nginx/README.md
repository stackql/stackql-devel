

## Runnning nginx as a pass through TCP proxy

### Example of proxied stackql.io website

This is a brittle example with ordering super important.

From the root of the repository:

```bash

nginx -c $(pwd)/test/tcp/reverse-proxy/nginx/tls-pass-through.conf

```

**After** this, if I add this line to `/etc/hosts`:

```
127.0.0.1       stackql.io
```

And then, 

Then:

```bash

curl -vvv https://stackql.io:9900/docs

```

Then I get the response from stackql.io... which is exactly what is desired.

To stop `nginx`:

```bash

nginx -s stop

```