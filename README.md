# XRay

**XRay** is an open-source Go library and CLI tool for database schema extraction and query execution. It supports multiple databases and provides a unified interface for developers and data engineers.

---

## Features

- Extracts schema metadata from popular databases.
- Executes SQL queries across different engines.
- Unified API for multiple backends.
- Easy integration into Go projects.
- CLI for quick inspection and automation.

---

## Installation

### Library

```bash
go get github.com/yindia/xray@latest
```

### CLI

#### MacOS

```bash
brew install yindia/homebrew-tap/xray
```

#### Linux

```bash
curl -sL https://raw.githubusercontent.com/yindia/xray/main/install.sh | sudo bash -s -- -b /usr/local/bin
```

---

## Quick Start

### Go Library Example

```go
package main

import (
    "github.com/yindia/xray"
    "github.com/yindia/xray/config"
)

func main() {
    cfg := config.Config{
        // Fill in your database config here
    }
    client, err := xray.NewClient(cfg)
    if err != nil {
        panic(err)
    }

    schema, err := client.ExtractSchema()
    if err != nil {
        panic(err)
    }

    // Use schema metadata
    fmt.Println(schema)
}
```

See [example/{database}/main.go](./example/) for full working examples for each supported database.

---

## Supported Databases

- MySQL
- PostgreSQL
- Redshift
- BigQuery
- Snowflake
- MSSQL

---

## Integration Guides

- [MySQL Integration](./example/mysql/integration.md)
- [Postgres Integration](./example/postgres/integration.md)
- [Redshift Integration](./example/redshift/integration.md)
- [BigQuery Integration](./example/bigquery/integration.md)
- [Snowflake Integration](./example/snowflake/integration.md)
- [MSSQL Integration](./example/mssql/integration.md)

## Example Projects

- [MySQL Example](./example/mysql/README.md)
- [Postgres Example](./example/postgres/README.md)
- [Redshift Example](./example/redshift/README.md)
- [BigQuery Example](./example/bigquery/README.md)
- [Snowflake Example](./example/snowflake/README.md)
- [MSSQL Example](./example/mssql/README.md)

---

## CLI Usage

See [CLI Getting Started](./cli/README.md) for full documentation.

---

## Documentation

- [GoDoc Reference](https://pkg.go.dev/github.com/yindia/xray)

---

## Contributing

Contributions are welcome! Please open issues or pull requests.

---

## License

[MIT](./LICENSE)

---

## Show Your Support

If you find XRay useful, please consider starring the repository on GitHub!



