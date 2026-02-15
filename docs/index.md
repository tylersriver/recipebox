# RecipeBox

A self-hosted web application to import recipes from URLs, store them locally, and search/browse your collection.

## Features

- **Recipe Import** — Scrape recipes from any site using JSON-LD (schema.org/Recipe) or WP Recipe Maker markup
- **Full-Text Search** — Search across recipe titles, descriptions, ingredients, and more powered by SQLite FTS5
- **Browse & View** — Card-based recipe browsing with detailed recipe pages
- **Live Search** — Debounced live search via Datastar SSE
- **CLI Tools** — List, search, and view recipes from the command line
- **Single Binary** — No external dependencies beyond the compiled binary

## Tech Stack

| Component | Technology |
|-----------|-----------|
| Language | Go 1.24+ |
| Architecture | DDD (Domain-Driven Design) |
| Database | SQLite (pure Go, no CGO) |
| CLI | Cobra + Viper |
| HTTP Server | Echo v4 |
| Templates | templ |
| Frontend | Datastar (SSE) + Bulma CSS |
| Scraping | goquery |

## Quick Start

```bash
# Build
make build

# Start the web server
./bin/recipebox serve

# Open http://localhost:8080
```

See the [Getting Started](getting-started.md) guide for full setup instructions.
