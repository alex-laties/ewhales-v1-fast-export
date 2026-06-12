import sqlite3
import sys
import os

if len(sys.argv) != 3:
    print("Usage: python3 anonymize.py <input.sql> <output.sql>")
    print("Example: python3 anonymize.py test_data.sql test_data_anon.sql")
    sys.exit(1)

input_file = sys.argv[1]
output_file = sys.argv[2]

if not os.path.isfile(input_file):
    print(f"Error: Input file '{input_file}' not found.")
    sys.exit(1)

with open(input_file, 'r', encoding='utf-8') as f:
    sql = f.read()

idx = sql.find('INSERT INTO')
if idx == -1:
    print("Error: Could not find any INSERT INTO statements in the input file.")
    sys.exit(1)

schema = sql[:idx]
inserts = sql[idx:]

# Strip MySQL-specific prefixes to parse in SQLite
inserts_clean = inserts.replace('test.logswp_postmeta', 'logswp_postmeta')
inserts_clean = inserts_clean.replace("\\'", "''")

conn = sqlite3.connect(':memory:')
c = conn.cursor()
c.execute("CREATE TABLE logswp_postmeta (meta_id INT, post_id INT, meta_key VARCHAR, meta_value VARCHAR);")
c.executescript(inserts_clean)

c.execute("SELECT meta_id, post_id, meta_key, meta_value FROM logswp_postmeta")
rows = c.fetchall()

updates = []
for meta_id, post_id, meta_key, meta_value in rows:
    new_value = meta_value
    if meta_value is None:
        continue
        
    if meta_key == 'logbook_id':
        # Pure integers are logbook pointers, anything else is a logbook name to be anonymized
        if not str(meta_value).isdigit():
            new_value = f"Logbook-{post_id}"
    else:
        # Anonymize all other string text to hide PII, keeping numerics intact
        if isinstance(meta_value, str):
            if str(meta_value).isdigit():
                new_value = meta_value
            else:
                new_value = 'Anonymized'
            
    if new_value != meta_value:
        updates.append((new_value, meta_id))

# Batch update all fields
c.executemany("UPDATE logswp_postmeta SET meta_value = ? WHERE meta_id = ?", updates)

with open(output_file, 'w', encoding='utf-8') as f:
    f.write(schema)
    c.execute("SELECT meta_id, post_id, meta_key, meta_value FROM logswp_postmeta ORDER BY meta_id")
    for meta_id, post_id, meta_key, meta_value in c.fetchall():
        if meta_value is None:
            val = 'NULL'
        else:
            # Re-escape quotes for MySQL syntax
            val = "'" + str(meta_value).replace("'", "\\'") + "'"
        f.write(f"INSERT INTO test.logswp_postmeta (meta_id, post_id, meta_key, meta_value) VALUES ({meta_id}, {post_id}, '{meta_key}', {val});\n")

print(f"Successfully anonymized {input_file} to {output_file}")
