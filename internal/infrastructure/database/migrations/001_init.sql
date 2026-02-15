CREATE TABLE IF NOT EXISTS recipes (
    id TEXT PRIMARY KEY,
    title TEXT NOT NULL,
    description TEXT NOT NULL DEFAULT '',
    prep_time TEXT NOT NULL DEFAULT '',
    cook_time TEXT NOT NULL DEFAULT '',
    total_time TEXT NOT NULL DEFAULT '',
    servings TEXT NOT NULL DEFAULT '',
    cuisine TEXT NOT NULL DEFAULT '',
    course TEXT NOT NULL DEFAULT '',
    image_url TEXT NOT NULL DEFAULT '',
    source_url TEXT NOT NULL UNIQUE,
    author TEXT NOT NULL DEFAULT '',
    nutrition TEXT NOT NULL DEFAULT '',
    created_at DATETIME NOT NULL,
    updated_at DATETIME NOT NULL
);

CREATE TABLE IF NOT EXISTS ingredients (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    recipe_id TEXT NOT NULL REFERENCES recipes(id) ON DELETE CASCADE,
    amount TEXT NOT NULL DEFAULT '',
    unit TEXT NOT NULL DEFAULT '',
    name TEXT NOT NULL DEFAULT '',
    notes TEXT NOT NULL DEFAULT '',
    raw TEXT NOT NULL DEFAULT '',
    sort_order INTEGER NOT NULL DEFAULT 0
);

CREATE INDEX IF NOT EXISTS idx_ingredients_recipe_id ON ingredients(recipe_id);

CREATE TABLE IF NOT EXISTS instructions (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    recipe_id TEXT NOT NULL REFERENCES recipes(id) ON DELETE CASCADE,
    step_number INTEGER NOT NULL,
    text TEXT NOT NULL DEFAULT ''
);

CREATE INDEX IF NOT EXISTS idx_instructions_recipe_id ON instructions(recipe_id);

CREATE VIRTUAL TABLE IF NOT EXISTS recipes_fts USING fts5(
    title, description, cuisine, course, author,
    content='recipes',
    content_rowid='rowid'
);

CREATE VIRTUAL TABLE IF NOT EXISTS ingredients_fts USING fts5(
    name, raw,
    content='ingredients',
    content_rowid='id'
);

-- Triggers to keep FTS in sync
CREATE TRIGGER IF NOT EXISTS recipes_ai AFTER INSERT ON recipes BEGIN
    INSERT INTO recipes_fts(rowid, title, description, cuisine, course, author)
    VALUES (new.rowid, new.title, new.description, new.cuisine, new.course, new.author);
END;

CREATE TRIGGER IF NOT EXISTS recipes_ad AFTER DELETE ON recipes BEGIN
    INSERT INTO recipes_fts(recipes_fts, rowid, title, description, cuisine, course, author)
    VALUES ('delete', old.rowid, old.title, old.description, old.cuisine, old.course, old.author);
END;

CREATE TRIGGER IF NOT EXISTS recipes_au AFTER UPDATE ON recipes BEGIN
    INSERT INTO recipes_fts(recipes_fts, rowid, title, description, cuisine, course, author)
    VALUES ('delete', old.rowid, old.title, old.description, old.cuisine, old.course, old.author);
    INSERT INTO recipes_fts(rowid, title, description, cuisine, course, author)
    VALUES (new.rowid, new.title, new.description, new.cuisine, new.course, new.author);
END;

CREATE TRIGGER IF NOT EXISTS ingredients_ai AFTER INSERT ON ingredients BEGIN
    INSERT INTO ingredients_fts(rowid, name, raw)
    VALUES (new.id, new.name, new.raw);
END;

CREATE TRIGGER IF NOT EXISTS ingredients_ad AFTER DELETE ON ingredients BEGIN
    INSERT INTO ingredients_fts(ingredients_fts, rowid, name, raw)
    VALUES ('delete', old.id, old.name, old.raw);
END;

CREATE TRIGGER IF NOT EXISTS ingredients_au AFTER UPDATE ON ingredients BEGIN
    INSERT INTO ingredients_fts(ingredients_fts, rowid, name, raw)
    VALUES ('delete', old.id, old.name, old.raw);
    INSERT INTO ingredients_fts(rowid, name, raw)
    VALUES (new.id, new.name, new.raw);
END;
