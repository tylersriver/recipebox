# Configuration

RecipeBox uses [Viper](https://github.com/spf13/viper) for configuration, loading settings from a `.recipebox.yaml` file.

## Config File Location

Viper searches for `.recipebox.yaml` in:

1. Current working directory
2. Home directory (`$HOME`)

## Options

```yaml
server:
  host: "0.0.0.0"    # Bind address
  port: 8080          # HTTP port

database:
  path: "recipebox.db"  # SQLite database file path
```

## Defaults

All settings have sensible defaults. You can run RecipeBox without any config file.

| Setting | Default | Description |
|---------|---------|-------------|
| `server.host` | `0.0.0.0` | Address the web server binds to |
| `server.port` | `8080` | Port the web server listens on |
| `database.path` | `recipebox.db` | Path to the SQLite database file |

## CLI Overrides

The database path can be overridden via the `--db` flag:

```bash
recipebox --db /path/to/recipes.db serve
```

## Database

The SQLite database is created automatically on first run. It uses:

- **WAL mode** for concurrent read/write performance
- **Foreign keys** enabled for referential integrity
- **Embedded migrations** that run automatically on startup
