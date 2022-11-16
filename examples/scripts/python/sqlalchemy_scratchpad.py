import sqlalchemy

eng = sqlalchemy.create_engine('postgresql://stackql:stackql@127.0.0.1:5474/stackql')

## this is the sticking point for now
conn = eng.raw_connection()

curs = conn.cursor()

curs.execute("show transaction isolation level")

rv = curs.fetchall()

for entry in rv:
    print(entry)