---
title: "authelia storage user identifiers import"
description: "Reference for the authelia storage user identifiers import command."
lead: ""
date: 2022-06-15T17:51:47+10:00
draft: false
images: []
menu:
  reference:
    parent: "cli-authelia"
weight: 330
toc: true
---

## authelia storage user identifiers import

Import the identifiers from a YAML file

### Synopsis

Import the identifiers from a YAML file.

This subcommand allows you to import the opaque identifiers for users from a YAML file.

The YAML file can either be automatically generated using the authelia storage user identifiers export command, or
manually provided the file is in the same format.

```
authelia storage user identifiers import [flags]
```

### Examples

```
authelia storage user identifiers import
authelia storage user identifiers import --file export.yaml
authelia storage user identifiers import --file export.yaml --config config.yml
authelia storage user identifiers import --file export.yaml --encryption-key b3453fde-ecc2-4a1f-9422-2707ddbed495 --postgres.host postgres --postgres.password autheliapw
```

### Options

```
  -f, --file string   The file name for the YAML import (default "user-opaque-identifiers.yml")
  -h, --help          help for import
```

### Options inherited from parent commands

```
  -c, --config strings                         configuration files to load (default [configuration.yml])
      --encryption-key string                  the storage encryption key to use
      --mysql.database string                  the MySQL database name (default "authelia")
      --mysql.host string                      the MySQL hostname
      --mysql.password string                  the MySQL password
      --mysql.port int                         the MySQL port (default 3306)
      --mysql.username string                  the MySQL username (default "authelia")
      --postgres.database string               the PostgreSQL database name (default "authelia")
      --postgres.host string                   the PostgreSQL hostname
      --postgres.password string               the PostgreSQL password
      --postgres.port int                      the PostgreSQL port (default 5432)
      --postgres.schema string                 the PostgreSQL schema name (default "public")
      --postgres.ssl.certificate string        the PostgreSQL ssl certificate file location
      --postgres.ssl.key string                the PostgreSQL ssl key file location
      --postgres.ssl.mode string               the PostgreSQL ssl mode (default "disable")
      --postgres.ssl.root_certificate string   the PostgreSQL ssl root certificate file location
      --postgres.username string               the PostgreSQL username (default "authelia")
      --sqlite.path string                     the SQLite database path
```

### SEE ALSO

* [authelia storage user identifiers](authelia_storage_user_identifiers.md)	 - Manage user opaque identifiers
