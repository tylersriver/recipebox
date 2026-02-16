package repository

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/tyler/recipebox/internal/domain/entity"
	"github.com/tyler/recipebox/internal/domain/repository"
)

type sqliteRecipeRepository struct {
	db *sql.DB
}

// NewSQLiteRecipeRepository creates a new SQLite-backed recipe repository.
func NewSQLiteRecipeRepository(db *sql.DB) repository.RecipeRepository {
	return &sqliteRecipeRepository{db: db}
}

func (r *sqliteRecipeRepository) Save(ctx context.Context, vr *entity.ValidatedRecipe) error {
	recipe := vr.Recipe()

	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("beginning transaction: %w", err)
	}
	defer tx.Rollback()

	_, err = tx.ExecContext(ctx, `
		INSERT INTO recipes (id, title, description, prep_time, cook_time, total_time, servings, cuisine, course, image_url, source_url, author, nutrition, created_at, updated_at)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
		ON CONFLICT(source_url) DO UPDATE SET
			title=excluded.title, description=excluded.description,
			prep_time=excluded.prep_time, cook_time=excluded.cook_time, total_time=excluded.total_time,
			servings=excluded.servings, cuisine=excluded.cuisine, course=excluded.course,
			image_url=excluded.image_url, author=excluded.author, nutrition=excluded.nutrition,
			updated_at=excluded.updated_at`,
		recipe.ID, recipe.Title, recipe.Description,
		recipe.PrepTime, recipe.CookTime, recipe.TotalTime,
		recipe.Servings, recipe.Cuisine, recipe.Course,
		recipe.ImageURL, recipe.SourceURL, recipe.Author, recipe.Nutrition,
		recipe.CreatedAt, recipe.UpdatedAt,
	)
	if err != nil {
		return fmt.Errorf("upserting recipe: %w", err)
	}

	// Delete existing ingredients and instructions for upsert
	if _, err := tx.ExecContext(ctx, "DELETE FROM ingredients WHERE recipe_id = ?", recipe.ID); err != nil {
		return fmt.Errorf("deleting old ingredients: %w", err)
	}
	if _, err := tx.ExecContext(ctx, "DELETE FROM instructions WHERE recipe_id = ?", recipe.ID); err != nil {
		return fmt.Errorf("deleting old instructions: %w", err)
	}

	for i, ing := range recipe.Ingredients {
		_, err := tx.ExecContext(ctx, `
			INSERT INTO ingredients (recipe_id, amount, unit, name, notes, raw, sort_order)
			VALUES (?, ?, ?, ?, ?, ?, ?)`,
			recipe.ID, ing.Amount, ing.Unit, ing.Name, ing.Notes, ing.Raw, i,
		)
		if err != nil {
			return fmt.Errorf("inserting ingredient: %w", err)
		}
	}

	for _, inst := range recipe.Instructions {
		_, err := tx.ExecContext(ctx, `
			INSERT INTO instructions (recipe_id, step_number, text)
			VALUES (?, ?, ?)`,
			recipe.ID, inst.StepNumber, inst.Text,
		)
		if err != nil {
			return fmt.Errorf("inserting instruction: %w", err)
		}
	}

	return tx.Commit()
}

func (r *sqliteRecipeRepository) FindByID(ctx context.Context, id string) (*entity.Recipe, error) {
	recipe, err := r.scanRecipe(ctx, "SELECT id, title, description, prep_time, cook_time, total_time, servings, cuisine, course, image_url, source_url, author, nutrition, notes, created_at, updated_at FROM recipes WHERE id = ?", id)
	if err != nil {
		return nil, err
	}
	if recipe == nil {
		return nil, nil
	}
	if err := r.loadRelations(ctx, recipe); err != nil {
		return nil, err
	}
	return recipe, nil
}

func (r *sqliteRecipeRepository) FindBySourceURL(ctx context.Context, sourceURL string) (*entity.Recipe, error) {
	recipe, err := r.scanRecipe(ctx, "SELECT id, title, description, prep_time, cook_time, total_time, servings, cuisine, course, image_url, source_url, author, nutrition, notes, created_at, updated_at FROM recipes WHERE source_url = ?", sourceURL)
	if err != nil {
		return nil, err
	}
	if recipe == nil {
		return nil, nil
	}
	if err := r.loadRelations(ctx, recipe); err != nil {
		return nil, err
	}
	return recipe, nil
}

func (r *sqliteRecipeRepository) FindAll(ctx context.Context, offset, limit int) ([]*entity.Recipe, int, error) {
	var total int
	if err := r.db.QueryRowContext(ctx, "SELECT COUNT(*) FROM recipes").Scan(&total); err != nil {
		return nil, 0, fmt.Errorf("counting recipes: %w", err)
	}

	rows, err := r.db.QueryContext(ctx,
		"SELECT id, title, description, prep_time, cook_time, total_time, servings, cuisine, course, image_url, source_url, author, nutrition, notes, created_at, updated_at FROM recipes ORDER BY created_at DESC LIMIT ? OFFSET ?",
		limit, offset,
	)
	if err != nil {
		return nil, 0, fmt.Errorf("querying recipes: %w", err)
	}
	defer rows.Close()

	recipes, err := r.scanRecipes(ctx, rows)
	if err != nil {
		return nil, 0, err
	}
	return recipes, total, nil
}

func (r *sqliteRecipeRepository) Search(ctx context.Context, query string, offset, limit int) ([]*entity.Recipe, int, error) {
	// Search across both recipe FTS and ingredient FTS, deduplicate by recipe ID
	sqlQuery := `
		SELECT DISTINCT r.id, r.title, r.description, r.prep_time, r.cook_time, r.total_time,
		       r.servings, r.cuisine, r.course, r.image_url, r.source_url, r.author, r.nutrition,
		       r.notes, r.created_at, r.updated_at
		FROM recipes r
		LEFT JOIN ingredients i ON i.recipe_id = r.id
		WHERE r.rowid IN (SELECT rowid FROM recipes_fts WHERE recipes_fts MATCH ?)
		   OR i.id IN (SELECT rowid FROM ingredients_fts WHERE ingredients_fts MATCH ?)
		ORDER BY r.created_at DESC
		LIMIT ? OFFSET ?`

	ftsQuery := query + "*"

	rows, err := r.db.QueryContext(ctx, sqlQuery, ftsQuery, ftsQuery, limit, offset)
	if err != nil {
		return nil, 0, fmt.Errorf("searching recipes: %w", err)
	}
	defer rows.Close()

	recipes, err := r.scanRecipes(ctx, rows)
	if err != nil {
		return nil, 0, err
	}

	// Count total matches
	countQuery := `
		SELECT COUNT(DISTINCT r.id)
		FROM recipes r
		LEFT JOIN ingredients i ON i.recipe_id = r.id
		WHERE r.rowid IN (SELECT rowid FROM recipes_fts WHERE recipes_fts MATCH ?)
		   OR i.id IN (SELECT rowid FROM ingredients_fts WHERE ingredients_fts MATCH ?)`

	var total int
	if err := r.db.QueryRowContext(ctx, countQuery, ftsQuery, ftsQuery).Scan(&total); err != nil {
		return nil, 0, fmt.Errorf("counting search results: %w", err)
	}

	return recipes, total, nil
}

func (r *sqliteRecipeRepository) UpdateNotes(ctx context.Context, id string, notes string) error {
	result, err := r.db.ExecContext(ctx, "UPDATE recipes SET notes = ?, updated_at = datetime('now') WHERE id = ?", notes, id)
	if err != nil {
		return fmt.Errorf("updating notes: %w", err)
	}
	n, _ := result.RowsAffected()
	if n == 0 {
		return fmt.Errorf("recipe not found: %s", id)
	}
	return nil
}

func (r *sqliteRecipeRepository) Delete(ctx context.Context, id string) error {
	result, err := r.db.ExecContext(ctx, "DELETE FROM recipes WHERE id = ?", id)
	if err != nil {
		return fmt.Errorf("deleting recipe: %w", err)
	}
	n, _ := result.RowsAffected()
	if n == 0 {
		return fmt.Errorf("recipe not found: %s", id)
	}
	return nil
}

func (r *sqliteRecipeRepository) scanRecipe(ctx context.Context, query string, args ...any) (*entity.Recipe, error) {
	row := r.db.QueryRowContext(ctx, query, args...)
	recipe := &entity.Recipe{}
	err := row.Scan(
		&recipe.ID, &recipe.Title, &recipe.Description,
		&recipe.PrepTime, &recipe.CookTime, &recipe.TotalTime,
		&recipe.Servings, &recipe.Cuisine, &recipe.Course,
		&recipe.ImageURL, &recipe.SourceURL, &recipe.Author, &recipe.Nutrition,
		&recipe.Notes, &recipe.CreatedAt, &recipe.UpdatedAt,
	)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("scanning recipe: %w", err)
	}
	return recipe, nil
}

func (r *sqliteRecipeRepository) scanRecipes(ctx context.Context, rows *sql.Rows) ([]*entity.Recipe, error) {
	var recipes []*entity.Recipe
	for rows.Next() {
		recipe := &entity.Recipe{}
		err := rows.Scan(
			&recipe.ID, &recipe.Title, &recipe.Description,
			&recipe.PrepTime, &recipe.CookTime, &recipe.TotalTime,
			&recipe.Servings, &recipe.Cuisine, &recipe.Course,
			&recipe.ImageURL, &recipe.SourceURL, &recipe.Author, &recipe.Nutrition,
			&recipe.Notes, &recipe.CreatedAt, &recipe.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("scanning recipe row: %w", err)
		}
		if err := r.loadRelations(ctx, recipe); err != nil {
			return nil, err
		}
		recipes = append(recipes, recipe)
	}
	return recipes, rows.Err()
}

func (r *sqliteRecipeRepository) loadRelations(ctx context.Context, recipe *entity.Recipe) error {
	// Load ingredients
	rows, err := r.db.QueryContext(ctx,
		"SELECT amount, unit, name, notes, raw FROM ingredients WHERE recipe_id = ? ORDER BY sort_order", recipe.ID)
	if err != nil {
		return fmt.Errorf("loading ingredients: %w", err)
	}
	defer rows.Close()
	for rows.Next() {
		var ing entity.Ingredient
		if err := rows.Scan(&ing.Amount, &ing.Unit, &ing.Name, &ing.Notes, &ing.Raw); err != nil {
			return fmt.Errorf("scanning ingredient: %w", err)
		}
		recipe.Ingredients = append(recipe.Ingredients, ing)
	}
	if err := rows.Err(); err != nil {
		return err
	}

	// Load instructions
	rows2, err := r.db.QueryContext(ctx,
		"SELECT step_number, text FROM instructions WHERE recipe_id = ? ORDER BY step_number", recipe.ID)
	if err != nil {
		return fmt.Errorf("loading instructions: %w", err)
	}
	defer rows2.Close()
	for rows2.Next() {
		var inst entity.Instruction
		if err := rows2.Scan(&inst.StepNumber, &inst.Text); err != nil {
			return fmt.Errorf("scanning instruction: %w", err)
		}
		recipe.Instructions = append(recipe.Instructions, inst)
	}
	return rows2.Err()
}
