

# Analytics with stackql

The canonical pattern is a postgres backend.  To meaningfully develop analytics capability, **real** authenticated access to providers plus a postgres backend is needed. Therefore for local development:

- Ensure that all env var secrets are exported from the `.gitignore`d file `cicd/vol/vendor-secrets/secrets.sh`.
- Run and kill development containers with `docker compose -f docker-compose-live.yml down --volumes` / `docker compose -f docker-compose-live.yml up --force-recreate`.
- Connect and develop queries with `psql "postgresql://stackql:stackql@127.0.0.1:8632/stackql"`.

