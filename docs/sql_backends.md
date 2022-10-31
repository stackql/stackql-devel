
## Golang SQL drivers

How to use drivers:

- https://go.dev/doc/database/open-handle

List of drivers:

- https://github.com/golang/go/wiki/SQLDrivers

### SQLite

The default implementation is embedded SQLite.  SQLite does not have a wire protocol or TCP-native version.

### Postgres

#### Postgres over TCP

- [Using golang SQL driver interfaces](https://github.com/jackc/pgx/wiki/Getting-started-with-pgx-through-database-sql#hello-world-from-postgresql).
- [PGX native (improved performance)](https://github.com/jackc/pgx/wiki/Getting-started-with-pgx).

#### Postgres in process

https://github.com/fergusstrange/embedded-postgres

