
## Golang SQL drivers

How to use drivers:

- https://go.dev/doc/database/open-handle

List of drivers:

- https://github.com/golang/go/wiki/SQLDrivers

### SQLite

The default implementation is embedded SQLite.  SQLite is file based and does not have a wire protocol or TCP-native version.

### Postgres

#### Postgres over TCP

https://github.com/jackc/pgx/wiki/Getting-started-with-pgx-through-database-sql#hello-world-from-postgresql

#### Postgres in process

https://github.com/fergusstrange/embedded-postgres

