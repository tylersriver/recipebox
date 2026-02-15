# Web UI

The RecipeBox web interface is built with [Bulma CSS](https://bulma.io/) for styling and [Datastar](https://data-star.dev/) for reactive SSE-powered interactions.

## Pages

### Home (`/`)

The home page features a live search bar and a grid of recently imported recipes. Search results update as you type with a 300ms debounce.

### Browse (`/recipes`)

A paginated card grid showing all recipes in your collection, ordered by import date (newest first).

### Recipe Detail (`/recipes/:id`)

Full recipe view showing:

- Title, author, and source link
- Prep time, cook time, total time, and servings
- Cuisine and course tags
- Ingredient list
- Numbered instructions
- Recipe image (if available)

### Import (`/import`)

Paste a recipe URL and click **Import Recipe**. The button shows a loading spinner during import, and displays a success message with a link to the imported recipe on completion.

## How It Works

### Datastar SSE

The import and search features use [Datastar](https://data-star.dev/) for server-sent event (SSE) driven updates. Instead of traditional form submissions or AJAX calls, the browser opens an SSE connection and the server streams HTML fragment updates back.

**Import flow:**

1. User clicks Import — Datastar sends a POST to `/import/submit` with the URL signal
2. Server scrapes the recipe, saves it, and streams back an SSE fragment updating `#import-result`
3. The button's loading state is managed via Datastar signals (`$importing`)

**Search flow:**

1. User types in the search box — Datastar sends a GET to `/search/results` after 300ms debounce
2. Server queries FTS5 and streams back an SSE fragment updating `#search-results`

### templ Templates

All HTML is generated server-side using [templ](https://templ.guide/), providing type-safe Go templates. The templates live in `internal/interface/web/template/` as `.templ` files which are compiled to Go code.
