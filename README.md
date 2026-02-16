# RecipeBox

A self-hosted web application to import recipes from URLs, store them locally, and search/browse your collection.

**[Documentation](https://tylersriver.github.io/recipebox/)**

## Features

- Import recipes from any site using JSON-LD (schema.org/Recipe) or WP Recipe Maker markup
- Full-text search across recipe titles, descriptions, ingredients, and more (SQLite FTS5)
- Browse recipes with a card-based UI
- Live search with debounced typing via Datastar SSE
- Dark mode with system preference detection and manual toggle
- CLI for importing, listing, and searching recipes
- Single binary, no external dependencies beyond the binary itself

## Tech Stack

- **Go** with DDD layered architecture
- **SQLite** (pure Go, no CGO) with FTS5 full-text search
- **Cobra/Viper** for CLI and config
- **Echo** for HTTP routing
- **templ** for type-safe HTML templates
- **Datastar** for hypermedia/SSE interactions
- **Bulma CSS** for styling (via CDN)

## Getting Started

### Prerequisites

- Go 1.24+
- [templ](https://templ.guide) CLI (`go install github.com/a-h/templ/cmd/templ@latest`)

### Build & Run

```bash
make build              # generate templ + compile to bin/recipebox
./bin/recipebox serve   # start web server at http://localhost:8080
```

### Docker

```bash
docker-compose up recipebox
```

This builds and runs the app on port 8080 with a persistent data volume.

### Development

For hot-reloading during development (requires [Air](https://github.com/air-verse/air)):

```bash
make dev
```

Air watches `.go`, `.templ`, `.sql`, and `.yaml` files and automatically rebuilds on changes.

### CLI Usage

```bash
# Import a recipe from a URL
./bin/recipebox import "https://example.com/wprm_print/some-recipe"

# List all stored recipes
./bin/recipebox list

# Search recipes by keyword
./bin/recipebox search "chicken"

# Start the web server
./bin/recipebox serve
```

## Configuration

RecipeBox looks for `.recipebox.yaml` in the current directory or `$HOME`. You can also use flags.

```yaml
server:
  host: "0.0.0.0"
  port: 8080

database:
  path: "recipebox.db"
```

Override the database path with `--db`:

```bash
./bin/recipebox --db /path/to/recipes.db serve
```

## Project Structure

```
cmd/recipebox/          CLI entry point (Cobra commands)
internal/
  domain/               Entities, repository interface, scraper interface
  application/          Commands, queries, DTOs, service orchestrator
  infrastructure/       SQLite repository, scraper implementations, migrations
  interface/web/        Echo server, handlers, templ templates
```
