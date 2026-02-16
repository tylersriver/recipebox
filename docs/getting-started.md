# Getting Started

## Prerequisites

- Go 1.24+
- [templ](https://templ.guide) CLI

Install the templ code generator:

```bash
go install github.com/a-h/templ/cmd/templ@latest
```

## Build

```bash
make build
```

This runs `templ generate` to produce Go code from `.templ` files, then compiles the binary to `bin/recipebox`.

## Run the Web Server

```bash
./bin/recipebox serve
```

Open [http://localhost:8080](http://localhost:8080) in your browser. From there you can:

1. **Import** a recipe by navigating to the Import page and pasting a recipe URL
2. **Browse** your saved recipes on the Browse page
3. **Search** for recipes using the search bar on the Home page

## Import a Recipe

Navigate to [http://localhost:8080/import](http://localhost:8080/import), paste a recipe URL, and click **Import Recipe**. RecipeBox supports:

- Any site with **JSON-LD** structured data (schema.org/Recipe)
- Sites using **WP Recipe Maker** (WPRM) plugin

!!! tip "WPRM Print URLs"
    For WP Recipe Maker sites, the print URL format (`/wprm_print/recipe-slug`) often works best for scraping.

## Docker

You can also run RecipeBox via Docker:

```bash
docker-compose up recipebox
```

This builds and starts the app on port 8080 with a persistent volume for the database.

## Development

For hot-reloading during development, use [Air](https://github.com/air-verse/air):

```bash
make dev
```

Air watches `.go`, `.templ`, `.sql`, and `.yaml` files, automatically regenerates templ templates, and rebuilds on changes.

You can also manually regenerate and rebuild:

```bash
make build
```
