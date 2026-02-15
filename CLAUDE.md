# CLAUDE.md

## Build & Run

```bash
make build              # generate templ + compile to bin/recipebox
make run                # build then start web server
make test               # run all tests
make generate           # run go generate (templ, etc.)
```

After editing `.templ` files, you must regenerate before building:
```bash
~/go/bin/templ generate && go build -o bin/recipebox .
```

## Architecture

DDD layered architecture — dependencies point inward:

- `internal/domain/` — entities, value objects, repository and scraper interfaces (no external deps)
- `internal/application/` — commands, queries, DTOs, mapper, service orchestrator
- `internal/infrastructure/` — SQLite repository, scraper implementations (JSON-LD + WPRM), HTTP fetcher
- `internal/interface/web/` — Echo HTTP server, handlers, templ templates

Domain interfaces live in `internal/domain/repository/` and `internal/domain/service/`. Infrastructure implements them.

## Key Conventions

- **templ templates** are all in one package: `internal/interface/web/template/`. Do not create sub-packages for components.
- **Generated `_templ.go` files** should not be edited by hand. Edit the `.templ` source files instead.
- **SQLite migrations** are embedded via `//go:embed` from `internal/infrastructure/database/migrations/`. Add new migrations as numbered `.sql` files (e.g., `002_add_tags.sql`).
- **Scraper extractors** try JSON-LD first, then WPRM. The WPRM extractor parses `window.wprm_recipes` JS objects, not just HTML classes.
- **Datastar SSE** is used for import and live search handlers. These handlers write SSE responses directly via `datastar.NewSSE()` — they do not use the standard `Render()` helper. Use `renderSSEFragment()` in `recipe_handler.go` as the pattern.
- **Config** is loaded by Viper from `.recipebox.yaml` (cwd or `$HOME`). Defaults: port 8080, db `recipebox.db`.

## Database

Pure Go SQLite via `modernc.org/sqlite` (no CGO). FTS5 virtual tables (`recipes_fts`, `ingredients_fts`) are kept in sync via triggers defined in the migration. Ingredients and instructions are stored as separate rows, not JSON.

## Testing

```bash
go test ./...
```

No test files exist yet. When adding tests, use `*_test.go` files alongside the code they test.
