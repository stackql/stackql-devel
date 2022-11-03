
## Golang SQL drivers

How to use drivers:

- https://go.dev/doc/database/open-handle

List of drivers:

- https://github.com/golang/go/wiki/SQLDrivers

### Data Source Name (DSN) strings

- [SQLite as per golang](https://github.com/mattn/go-sqlite3#dsn-examples).
- [Postgres URI](https://www.postgresql.org/docs/current/libpq-connect.html#LIBPQ-CONNSTRING).

## Backends

### SQLite

The default implementation is **embedded** SQLite.  SQLite does **not** have a wire protocol or TCP-native version.

### Postgres

#### Postgres over TCP

- [Using golang SQL driver interfaces](https://github.com/jackc/pgx/wiki/Getting-started-with-pgx-through-database-sql#hello-world-from-postgresql).
- [PGX native (improved performance)](https://github.com/jackc/pgx/wiki/Getting-started-with-pgx).

#### Embedded Postgres

https://github.com/fergusstrange/embedded-postgres


#### Setup postgres in docker

```sh
docker run -v "${PWD}/vol/postgres/setup:/docker-entrypoint-initdb.d:ro" -it --entrypoint bash postgres:14.5-bullseye
```

```sh
docker run -v "$(pwd)/vol/postgres/setup:/docker-entrypoint-initdb.d:ro" -p 127.0.0.1:6532:5432/tcp -e POSTGRES_PASSWORD=password postgres:14.5-bullseye
```

#### Setup postgres DB locally

```sql

CREATE database "stackql";

CREATE user stackql with password 'stackql';

GRANT ALL PRIVILEGES on DATABASE stackql to stackql;

```

