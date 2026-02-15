package scraper

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/PuerkitoBio/goquery"
	"github.com/tyler/recipebox/internal/domain/entity"
)

// JSONLDExtractor extracts recipe data from schema.org/Recipe JSON-LD.
type JSONLDExtractor struct{}

// NewJSONLDExtractor creates a new JSON-LD extractor.
func NewJSONLDExtractor() *JSONLDExtractor {
	return &JSONLDExtractor{}
}

type jsonLDRecipe struct {
	Context            any    `json:"@context"`
	Type               any    `json:"@type"`
	Name               string `json:"name"`
	Description        string `json:"description"`
	Image              any    `json:"image"`
	Author             any    `json:"author"`
	PrepTime           string `json:"prepTime"`
	CookTime           string `json:"cookTime"`
	TotalTime          string `json:"totalTime"`
	RecipeYield        any    `json:"recipeYield"`
	RecipeCategory     any    `json:"recipeCategory"`
	RecipeCuisine      any    `json:"recipeCuisine"`
	RecipeIngredient   []string `json:"recipeIngredient"`
	RecipeInstructions any    `json:"recipeInstructions"`
	Nutrition          any    `json:"nutrition"`
}

// Extract tries to find and parse JSON-LD recipe data from the document.
func (e *JSONLDExtractor) Extract(doc *goquery.Document) (*entity.Recipe, error) {
	var recipe *jsonLDRecipe

	doc.Find(`script[type="application/ld+json"]`).EachWithBreak(func(i int, s *goquery.Selection) bool {
		text := strings.TrimSpace(s.Text())
		r := e.tryParseRecipe(text)
		if r != nil {
			recipe = r
			return false
		}
		return true
	})

	if recipe == nil {
		return nil, fmt.Errorf("no JSON-LD recipe found")
	}

	return e.toEntity(recipe), nil
}

func (e *JSONLDExtractor) tryParseRecipe(text string) *jsonLDRecipe {
	// Try direct object
	var single jsonLDRecipe
	if err := json.Unmarshal([]byte(text), &single); err == nil {
		if e.isRecipeType(single.Type) {
			return &single
		}
	}

	// Try @graph wrapper
	var graphWrapper struct {
		Graph []json.RawMessage `json:"@graph"`
	}
	if err := json.Unmarshal([]byte(text), &graphWrapper); err == nil {
		for _, raw := range graphWrapper.Graph {
			var r jsonLDRecipe
			if err := json.Unmarshal(raw, &r); err == nil {
				if e.isRecipeType(r.Type) {
					return &r
				}
			}
		}
	}

	// Try array
	var arr []json.RawMessage
	if err := json.Unmarshal([]byte(text), &arr); err == nil {
		for _, raw := range arr {
			var r jsonLDRecipe
			if err := json.Unmarshal(raw, &r); err == nil {
				if e.isRecipeType(r.Type) {
					return &r
				}
			}
		}
	}

	return nil
}

func (e *JSONLDExtractor) isRecipeType(t any) bool {
	switch v := t.(type) {
	case string:
		return v == "Recipe" || v == "schema:Recipe" || strings.HasSuffix(v, "/Recipe")
	case []any:
		for _, item := range v {
			if s, ok := item.(string); ok {
				if s == "Recipe" || s == "schema:Recipe" || strings.HasSuffix(s, "/Recipe") {
					return true
				}
			}
		}
	}
	return false
}

func (e *JSONLDExtractor) toEntity(r *jsonLDRecipe) *entity.Recipe {
	recipe := &entity.Recipe{
		Title:       r.Name,
		Description: r.Description,
		PrepTime:    cleanDuration(r.PrepTime),
		CookTime:    cleanDuration(r.CookTime),
		TotalTime:   cleanDuration(r.TotalTime),
		Servings:    extractString(r.RecipeYield),
		Cuisine:     extractStringOrFirst(r.RecipeCuisine),
		Course:      extractStringOrFirst(r.RecipeCategory),
		ImageURL:    extractImageURL(r.Image),
		Author:      extractAuthor(r.Author),
	}

	for _, raw := range r.RecipeIngredient {
		recipe.Ingredients = append(recipe.Ingredients, entity.Ingredient{Raw: strings.TrimSpace(raw)})
	}

	recipe.Instructions = extractInstructions(r.RecipeInstructions)

	if r.Nutrition != nil {
		if b, err := json.Marshal(r.Nutrition); err == nil {
			recipe.Nutrition = string(b)
		}
	}

	return recipe
}

func cleanDuration(d string) string {
	d = strings.TrimPrefix(d, "PT")
	d = strings.TrimPrefix(d, "pt")
	if d == "" {
		return ""
	}
	d = strings.ReplaceAll(d, "H", "h ")
	d = strings.ReplaceAll(d, "M", "m")
	d = strings.ReplaceAll(d, "S", "s")
	return strings.TrimSpace(d)
}

func extractString(v any) string {
	switch val := v.(type) {
	case string:
		return val
	case []any:
		if len(val) > 0 {
			if s, ok := val[0].(string); ok {
				return s
			}
		}
	case float64:
		return fmt.Sprintf("%.0f", val)
	}
	return ""
}

func extractStringOrFirst(v any) string {
	switch val := v.(type) {
	case string:
		return val
	case []any:
		if len(val) > 0 {
			if s, ok := val[0].(string); ok {
				return s
			}
		}
	}
	return ""
}

func extractImageURL(v any) string {
	switch val := v.(type) {
	case string:
		return val
	case []any:
		if len(val) > 0 {
			if s, ok := val[0].(string); ok {
				return s
			}
		}
	case map[string]any:
		if url, ok := val["url"].(string); ok {
			return url
		}
	}
	return ""
}

func extractAuthor(v any) string {
	switch val := v.(type) {
	case string:
		return val
	case map[string]any:
		if name, ok := val["name"].(string); ok {
			return name
		}
	case []any:
		if len(val) > 0 {
			if m, ok := val[0].(map[string]any); ok {
				if name, ok := m["name"].(string); ok {
					return name
				}
			}
		}
	}
	return ""
}

func extractInstructions(v any) []entity.Instruction {
	var instructions []entity.Instruction
	step := 1

	switch val := v.(type) {
	case string:
		for _, line := range strings.Split(val, "\n") {
			line = strings.TrimSpace(line)
			if line != "" {
				instructions = append(instructions, entity.Instruction{StepNumber: step, Text: line})
				step++
			}
		}
	case []any:
		for _, item := range val {
			switch inst := item.(type) {
			case string:
				if t := strings.TrimSpace(inst); t != "" {
					instructions = append(instructions, entity.Instruction{StepNumber: step, Text: t})
					step++
				}
			case map[string]any:
				// HowToStep
				if text, ok := inst["text"].(string); ok {
					instructions = append(instructions, entity.Instruction{StepNumber: step, Text: strings.TrimSpace(text)})
					step++
				}
				// HowToSection
				if items, ok := inst["itemListElement"].([]any); ok {
					for _, subItem := range items {
						if m, ok := subItem.(map[string]any); ok {
							if text, ok := m["text"].(string); ok {
								instructions = append(instructions, entity.Instruction{StepNumber: step, Text: strings.TrimSpace(text)})
								step++
							}
						}
					}
				}
			}
		}
	}

	return instructions
}
