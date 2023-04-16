

```sql

"SET datestyle TO 'ISO'"

BEGIN;

"SELECT t.oid, NULL\nFROM pg_type t JOIN pg_namespace ns\n    ON typnamespace = ns.oid\nWHERE typname = 'hstore'"

ROLLBACK;

BEGIN;

"select pg_catalog.version()"

"select current_schema()"

"show transaction isolation level"

"show standard_conforming_strings"

ROLLBACK;

SET DATESTYLE TO 'ISO';

BEGIN;

"\n            SELECT c.oid\n            FROM pg_catalog.pg_class c\n            LEFT JOIN pg_catalog.pg_namespace n ON n.oid = c.relnamespace\n            WHERE (n.nspname = 'information_schema')\n            AND c.relname = 'attributes' AND c.relkind in\n            ('r', 'v', 'm', 'f', 'p')\n        ",
"\n            SELECT a.attname,\n              pg_catalog.format_type(a.atttypid, a.atttypmod),\n              (\n                SELECT pg_catalog.pg_get_expr(d.adbin, d.adrelid)\n                FROM pg_catalog.pg_attrdef d\n                WHERE d.adrelid = a.attrelid AND d.adnum = a.attnum\n                AND a.atthasdef\n              ) AS DEFAULT,\n              a.attnotnull,\n              a.attrelid as table_oid,\n              pgd.description as comment,\n              a.attgenerated as generated,\n                              (SELECT json_build_object(\n                    'always', a.attidentity = 'a',\n                    'start', s.seqstart,\n                    'increment', s.seqincrement,\n                    'minvalue', s.seqmin,\n                    'maxvalue', s.seqmax,\n                    'cache', s.seqcache,\n                    'cycle', s.seqcycle)\n                FROM pg_catalog.pg_sequence s\n                JOIN pg_catalog.pg_class c on s.seqrelid = c.\"oid\"\n                WHERE c.relkind = 'S'\n                AND a.attidentity != ''\n                AND s.seqrelid = pg_catalog.pg_get_serial_sequence(\n                    a.attrelid::regclass::text, a.attname\n                )::regclass::oid\n                ) as identity_options                \n            FROM pg_catalog.pg_attribute a\n            LEFT JOIN pg_catalog.pg_description pgd ON (\n                pgd.objoid = a.attrelid AND pgd.objsubid = a.attnum)\n            WHERE a.attrelid = '13429'\n            AND a.attnum > 0 AND NOT a.attisdropped\n            ORDER BY a.attnum\n        ",
"\n            SELECT t.typname as \"name\",\n               pg_catalog.format_type(t.typbasetype, t.typtypmod) as \"attype\",\n               not t.typnotnull as \"nullable\",\n               t.typdefault as \"default\",\n               pg_catalog.pg_type_is_visible(t.oid) as \"visible\",\n               n.nspname as \"schema\"\n            FROM pg_catalog.pg_type t\n               LEFT JOIN pg_catalog.pg_namespace n ON n.oid = t.typnamespace\n            WHERE t.typtype = 'd'\n        ",
"\n            SELECT t.typname as \"name\",\n               -- no enum defaults in 8.4 at least\n               -- t.typdefault as \"default\",\n               pg_catalog.pg_type_is_visible(t.oid) as \"visible\",\n               n.nspname as \"schema\",\n               e.enumlabel as \"label\"\n            FROM pg_catalog.pg_type t\n                 LEFT JOIN pg_catalog.pg_namespace n ON n.oid = t.typnamespace\n                 LEFT JOIN pg_catalog.pg_enum e ON t.oid = e.enumtypid\n            WHERE t.typtype = 'e'\n        ORDER BY \"schema\", \"name\", e.oid"

ROLLBACK;

"SET datestyle TO 'ISO'"

BEGIN;

"SELECT t.oid, NULL\nFROM pg_type t JOIN pg_namespace ns\n    ON typnamespace = ns.oid\nWHERE typname = 'hstore'"



```


```sql

SELECT c.oid
FROM pg_catalog.pg_class c
LEFT JOIN pg_catalog.pg_namespace n 
ON n.oid = c.relnamespace
WHERE (n.nspname = 'information_schema')
AND c.relname = 'attributes' AND c.relkind in
('r', 'v', 'm', 'f', 'p')
;

SELECT 
  a.attname,
  pg_catalog.format_type(a.atttypid, a.atttypmod),
  (
    SELECT pg_catalog.pg_get_expr(d.adbin, d.adrelid)
    FROM pg_catalog.pg_attrdef d
    WHERE d.adrelid = a.attrelid AND d.adnum = a.attnum
    AND a.atthasdef
  ) AS DEFAULT,
  a.attnotnull,
  a.attrelid as table_oid,
  pgd.description as comment,
  a.attgenerated as generated,
  (
    SELECT json_build_object(
        'always', a.attidentity = 'a',
        'start', s.seqstart,
        'increment', s.seqincrement,
        'minvalue', s.seqmin,
        'maxvalue', s.seqmax,
        'cache', s.seqcache,
        'cycle', s.seqcycle)
    FROM pg_catalog.pg_sequence s
    JOIN pg_catalog.pg_class c on s.seqrelid = c."oid"
    WHERE c.relkind = 'S'
    AND a.attidentity != ''
    AND s.seqrelid = pg_catalog.pg_get_serial_sequence(
        a.attrelid::regclass::text, a.attname
    )::regclass::oid
        ) as identity_options
    FROM pg_catalog.pg_attribute a
    LEFT JOIN pg_catalog.pg_description pgd ON (
        pgd.objoid = a.attrelid AND pgd.objsubid = a.attnum)
    WHERE a.attrelid = '13429'
    AND a.attnum > 0 AND NOT a.attisdropped
    ORDER BY a.attnum
;


SELECT 
  t.typname as "name",
  pg_catalog.format_type(t.typbasetype, t.typtypmod) as "attype",
  not t.typnotnull as "nullable",
  t.typdefault as "default",
  pg_catalog.pg_type_is_visible(t.oid) as "visible",
  n.nspname as "schema"
FROM pg_catalog.pg_type t
LEFT JOIN pg_catalog.pg_namespace n ON n.oid = t.typnamespace
WHERE t.typtype = 'd'
;


SELECT 
t.typname as "name",
-- no enum defaults in 8.4 at least
-- t.typdefault as "default",
pg_catalog.pg_type_is_visible(t.oid) as "visible",
n.nspname as "schema",
e.enumlabel as "label"
FROM pg_catalog.pg_type t
LEFT JOIN pg_catalog.pg_namespace n ON n.oid = t.typnamespace
LEFT JOIN pg_catalog.pg_enum e ON t.oid = e.enumtypid
WHERE t.typtype = 'e'
ORDER BY "schema", "name", e.oid
;


```