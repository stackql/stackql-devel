

SELECT a.attname, (SELECT s.seqstart FROM pg_catalog.pg_sequence s AND a.attidentity != '' AND s.seqrelid = pg_catalog.pg_get_serial_sequence( a.attrelid::regclass::text, a.attname )::regclass::oid ) as identity_options FROM pg_catalog.pg_attribute a ORDER BY a.attnum ;