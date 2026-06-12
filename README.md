# eWHALES v1 Fast Export

This repository contains tools for exporting data from the eWHALES v1 database. There are two primary workflows supported by these tools:

## Backup Processing

The `my.cnf` and `convert_inserts.py` files are intended for use with database backups. They facilitate the conversion and processing of data dumps when a live database connection is not available or desired.

## Live Database Export (Golang Tool)

The Go-based tool in this repository is intended to work directly with a live database. It extracts the required records and produces a CSV output in the exact format expected for the v1 eWHALES data processing pipeline. 

### Building the Tool

Ensure you have Go installed on your system. To compile the binary, run:
```bash
go build -o exporter
```

### Running the Tool

You can run the compiled binary directly. Use the `--help` flag to see available options:
```bash
./exporter --help
```

Available flags:
- `-config`: Path to the configuration file (default is `config.json`).
- `-progress`: Enable terminal progress bars for the querying and serialization phases.
- `-h` / `--help`: Print help info and exit.

Example usage:
```bash
./exporter -progress -config custom_config.json
```

### Configuration

The Golang exporter uses `config.json` to define database connection credentials, specify the output CSV filename, and configure field mappings. Ensure your configuration is correctly set up before running the tool against the live database.

### Generating Test Data

`query_test.go` uses sql dump files (e.g. `test_multiple_logbook_logbook_entries.sql`) to generate test data.

The following query can be used to extract test data from the live database and save it to a sql file. It is recommended to limit the number of logbooks to a small number for testing purposes. Replace `limit 10` with a larger number to increase the amount of data extracted or specify specific logbooks by replacing `limit 10` with `and where meta_value in ("logbook-name-1", "logbook-name-2")`. The query can look for as many logbooks as you'd like, e.g. `and where meta_value in ("Westward-1978-1979", "A. Houghton (bark) 1853-1857", "T. A. Spofford (bark) 1851-1855")`. Just keep in mind that some logbooks have more entries than others.

```
with logbooks as (select
         post_id
     from logswp_postmeta where
                              meta_key = "logbook_id"
                            and meta_value is not null
                            and meta_value <> ''
                            and meta_value REGEXP '^[a-zA-z].*' limit 10),
    logbook_entries as (select
                            post_id
                        from logswp_postmeta where
                                                 meta_key = "logbook_id"
                                               and meta_value <> ''
                                               and meta_value in (select * from logbooks))
select * from logswp_postmeta where post_id in (select * from logbooks) or post_id in (select * from logbook_entries);
```

Once you've exported the data to a `.sql` dump file, it's highly recommended to anonymize it before committing it to the repository as test data, since logbook entries may contain PII, researcher names, or proprietary notes.

You can anonymize the file using the provided `anonymize.py` script. This script randomizes sensitive text fields while perfectly preserving numeric IDs and `post_id` referential integrity so the Golang tests can still parse the logic.

```bash
# Usage: python3 anonymize.py <input.sql> <output.sql>
python3 anonymize.py test_data.sql test_data_anon.sql
```