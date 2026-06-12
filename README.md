# eWHALES v1 Fast Export

This repository contains tools for exporting data from the eWHALES v1 database. There are two primary workflows supported by these tools:

## Backup Processing

The `my.cnf` and `convert_inserts.py` files are intended for use with database backups. They facilitate the conversion and processing of data dumps when a live database connection is not available or desired.

## Live Database Export (Golang Tool)

The Go-based tool in this repository is intended to work directly with a live database. It extracts the required records and produces a CSV output in the exact format expected for the v1 eWHALES data processing pipeline. 

### Configuration

The Golang exporter uses `config.json` to define database connection credentials, specify the output CSV filename, and configure field mappings. Ensure your configuration is correctly set up before running the tool against the live database.
