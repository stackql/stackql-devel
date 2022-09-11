

## Setup activities

Centres around `sqlalchemy` / `psycopg2` setup actions:

- [sqlalchemy setup orchestration example](https://github.com/sqlalchemy/sqlalchemy/blob/479dbc99e7fc5a60f846992c0cca8542047a8933/lib/sqlalchemy/engine/default.py#L439)
- [sqlalchemy get server version info (one of the setup steps)](https://github.com/sqlalchemy/sqlalchemy/blob/479dbc99e7fc5a60f846992c0cca8542047a8933/lib/sqlalchemy/dialects/postgresql/base.py#L3082)


## Known queries required to suport


| Query | Example response (`psycopg` client) |
| --- | ----------- |
| `select pg_catalog.version()` | `[('PostgreSQL 14.5 on x86_64-apple-darwin20.6.0, compiled by Apple clang version 13.0.0 (clang-1300.0.29.30), 64-bit',)]` |
| `select current_schema()` | `[('public',)]` |
| `show transaction isolation level` | `[('read committed',)]` |
|  |  |
|  |  |
|  |  |
|  |  |
|  |  |
|  |  |
|  |  |