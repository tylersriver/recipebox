# CLI Reference

RecipeBox provides command-line tools for managing your recipe collection.

## Global Flags

| Flag | Description | Default |
|------|-------------|---------|
| `--db` | Database file path | `recipebox.db` |

## Commands

### `serve`

Start the web server.

```bash
recipebox serve
```

The server binds to the host and port configured in `.recipebox.yaml` (default `0.0.0.0:8080`).

### `list`

List all stored recipes.

```bash
recipebox list
```

Output:

```
Found 2 recipe(s):

  019c6302  Crock-Pot Chicken Tortilla Soup
  019c6302  No Bake Protein Balls
```

### `search`

Search recipes by keyword. Searches across titles, descriptions, ingredients, cuisine, course, and author.

```bash
recipebox search <query>
```

Example:

```bash
recipebox search chicken
```

Output:

```
Found 1 recipe(s) matching 'chicken':

  019c6302  Crock-Pot Chicken Tortilla Soup
```

### `show`

Display full details of a recipe. Accepts a full UUID or a unique prefix.

```bash
recipebox show <id>
```

Example:

```bash
recipebox show 019c6302
```

Output:

```
Title:       No Bake Protein Balls
Author:      Joy Shull
Source:      https://example.com/recipe
ID:          019c6302-1fc1-7b5b-b07e-300fc66b27a6

Servings:    27

Cuisine:     American
Course:      Snacks

Ingredients
----------------------------------------
  - 1 1/2 cups no stir creamy peanut butter
  - 1/2 cup honey
  ...

Instructions
----------------------------------------
  1. Add the peanut butter, honey, ...

  2. Use a mini cookie scoop ...
```

!!! tip "Partial IDs"
    You don't need the full UUID. Any unique prefix works — `019c63` is usually enough.
