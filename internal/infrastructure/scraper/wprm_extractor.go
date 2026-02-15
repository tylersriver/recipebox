package scraper

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/PuerkitoBio/goquery"
	"github.com/tyler/recipebox/internal/domain/entity"
)

// WPRMExtractor extracts recipe data from WP Recipe Maker HTML.
type WPRMExtractor struct{}

// NewWPRMExtractor creates a new WPRM extractor.
func NewWPRMExtractor() *WPRMExtractor {
	return &WPRMExtractor{}
}

// Extract tries to find and parse WPRM recipe data from the document.
// It first tries the window.wprm_recipes JS object, then falls back to HTML scraping.
func (e *WPRMExtractor) Extract(doc *goquery.Document) (*entity.Recipe, error) {
	// Try extracting from window.wprm_recipes JavaScript object
	recipe, err := e.extractFromJS(doc)
	if err == nil && recipe.Title != "" {
		return recipe, nil
	}

	// Fallback: try HTML scraping with various container classes
	return e.extractFromHTML(doc)
}

var wprmJSPrefix = "window.wprm_recipes"

type wprmJSData struct {
	Name               string            `json:"name"`
	ImageURL           string            `json:"image_url"`
	OriginalServings   string            `json:"originalServings"`
	Ingredients        []wprmJSIngredient `json:"ingredients"`
	Rating             json.RawMessage   `json:"rating"`
	Collection         wprmJSCollection  `json:"collection"`
}

type wprmJSIngredient struct {
	Amount      string          `json:"amount"`
	Unit        string          `json:"unit"`
	Name        string          `json:"name"`
	Notes       string          `json:"notes"`
	UnitSystems json.RawMessage `json:"unit_systems"`
}

type wprmJSCollection struct {
	Name      string `json:"name"`
	Image     string `json:"image"`
	Servings  int    `json:"servings"`
	ParentURL string `json:"parent_url"`
}

func (e *WPRMExtractor) extractFromJS(doc *goquery.Document) (*entity.Recipe, error) {
	var jsContent string
	doc.Find("script").Each(func(i int, s *goquery.Selection) {
		text := s.Text()
		if strings.Contains(text, "window.wprm_recipes") {
			jsContent = text
		}
	})

	if jsContent == "" {
		return nil, fmt.Errorf("no wprm_recipes JS object found")
	}

	jsonStr := extractJSONFromJS(jsContent, wprmJSPrefix)
	if jsonStr == "" {
		return nil, fmt.Errorf("could not parse wprm_recipes JS object")
	}

	// The outer object is keyed by recipe ID, parse each value
	var outer map[string]json.RawMessage
	if err := json.Unmarshal([]byte(jsonStr), &outer); err != nil {
		return nil, fmt.Errorf("parsing wprm_recipes JSON: %w", err)
	}

	for _, raw := range outer {
		var data wprmJSData
		if err := json.Unmarshal(raw, &data); err != nil {
			continue
		}
		if data.Name == "" {
			continue
		}

		recipe := &entity.Recipe{
			Title:    data.Name,
			ImageURL: data.ImageURL,
			Servings: data.OriginalServings,
		}

		for _, ing := range data.Ingredients {
			recipe.Ingredients = append(recipe.Ingredients, entity.Ingredient{
				Amount: ing.Amount,
				Unit:   ing.Unit,
				Name:   ing.Name,
				Notes:  ing.Notes,
				Raw:    strings.TrimSpace(fmt.Sprintf("%s %s %s", ing.Amount, ing.Unit, ing.Name)),
			})
		}

		// Also try to get instructions from HTML since JS object often doesn't have them
		e.extractInstructionsFromHTML(doc, recipe)

		// Extract additional metadata from HTML
		e.extractMetadataFromHTML(doc, recipe)

		return recipe, nil
	}

	return nil, fmt.Errorf("no valid recipe in wprm_recipes JS data")
}

func (e *WPRMExtractor) extractInstructionsFromHTML(doc *goquery.Document, recipe *entity.Recipe) {
	step := 1
	doc.Find(".wprm-recipe-instruction").Each(func(i int, s *goquery.Selection) {
		text := strings.TrimSpace(s.Find(".wprm-recipe-instruction-text").Text())
		if text == "" {
			text = strings.TrimSpace(s.Text())
		}
		if text != "" {
			recipe.Instructions = append(recipe.Instructions, entity.Instruction{StepNumber: step, Text: text})
			step++
		}
	})

	// If no instructions found via class, try list items in instruction groups
	if len(recipe.Instructions) == 0 {
		doc.Find(".wprm-recipe-instruction-group li").Each(func(i int, s *goquery.Selection) {
			text := strings.TrimSpace(s.Text())
			if text != "" {
				recipe.Instructions = append(recipe.Instructions, entity.Instruction{StepNumber: step, Text: text})
				step++
			}
		})
	}
}

func (e *WPRMExtractor) extractMetadataFromHTML(doc *goquery.Document, recipe *entity.Recipe) {
	if v := wprmValue(doc, ".wprm-recipe-prep_time-container"); v != "" {
		recipe.PrepTime = v
	}
	if v := wprmValue(doc, ".wprm-recipe-cook_time-container"); v != "" {
		recipe.CookTime = v
	}
	if v := wprmValue(doc, ".wprm-recipe-total_time-container"); v != "" {
		recipe.TotalTime = v
	}
	if v := wprmValue(doc, ".wprm-recipe-course-container"); v != "" {
		recipe.Course = v
	}
	if v := wprmValue(doc, ".wprm-recipe-cuisine-container"); v != "" {
		recipe.Cuisine = v
	}
	if v := wprmValue(doc, ".wprm-recipe-author-container"); v != "" {
		recipe.Author = v
	}
	if v := textByClass(doc, ".wprm-recipe-summary"); v != "" {
		recipe.Description = v
	}
}

// wprmValue extracts the value from a WPRM metadata container,
// preferring the .wprm-block-text-normal span (the value) over
// the full container text (which includes the label).
func wprmValue(doc *goquery.Document, selector string) string {
	container := doc.Find(selector).First()
	if container.Length() == 0 {
		return ""
	}
	// Try the value span first
	if val := container.Find(".wprm-block-text-normal").First(); val.Length() > 0 {
		return strings.TrimSpace(val.Text())
	}
	// Try the recipe-time span
	if val := container.Find(".wprm-recipe-time").First(); val.Length() > 0 {
		return strings.TrimSpace(val.Text())
	}
	// Fallback: strip common label prefixes from full text
	text := strings.TrimSpace(container.Text())
	for _, prefix := range []string{"Author:", "Cuisine:", "Course:", "Prep Time:", "Cook Time:", "Total Time:"} {
		text = strings.TrimPrefix(text, prefix)
	}
	return strings.TrimSpace(text)
}

func textByClass(doc *goquery.Document, selector string) string {
	return strings.TrimSpace(doc.Find(selector).First().Text())
}

// extractJSONFromJS finds a JSON object assigned to a JS variable by matching balanced braces.
func extractJSONFromJS(js, prefix string) string {
	idx := strings.Index(js, prefix)
	if idx == -1 {
		return ""
	}
	// Find the first '{' after the prefix
	start := strings.Index(js[idx:], "{")
	if start == -1 {
		return ""
	}
	start += idx

	depth := 0
	inString := false
	escape := false
	for i := start; i < len(js); i++ {
		c := js[i]
		if escape {
			escape = false
			continue
		}
		if c == '\\' && inString {
			escape = true
			continue
		}
		if c == '"' {
			inString = !inString
			continue
		}
		if inString {
			continue
		}
		if c == '{' {
			depth++
		} else if c == '}' {
			depth--
			if depth == 0 {
				return js[start : i+1]
			}
		}
	}
	return ""
}

func (e *WPRMExtractor) extractFromHTML(doc *goquery.Document) (*entity.Recipe, error) {
	// Try multiple container class patterns
	var container *goquery.Selection
	for _, sel := range []string{".wprm-recipe-container", "[class*='wprm-recipe-template']", ".wprm-recipe"} {
		container = doc.Find(sel).First()
		if container.Length() > 0 {
			break
		}
	}

	if container == nil || container.Length() == 0 {
		return nil, fmt.Errorf("no WPRM recipe container found")
	}

	recipe := &entity.Recipe{}

	recipe.Title = strings.TrimSpace(container.Find(".wprm-recipe-name").First().Text())
	if recipe.Title == "" {
		return nil, fmt.Errorf("no recipe title found in WPRM")
	}

	recipe.Description = strings.TrimSpace(container.Find(".wprm-recipe-summary").First().Text())
	recipe.PrepTime = strings.TrimSpace(container.Find(".wprm-recipe-prep_time-container .wprm-recipe-time").First().Text())
	recipe.CookTime = strings.TrimSpace(container.Find(".wprm-recipe-cook_time-container .wprm-recipe-time").First().Text())
	recipe.TotalTime = strings.TrimSpace(container.Find(".wprm-recipe-total_time-container .wprm-recipe-time").First().Text())
	recipe.Servings = strings.TrimSpace(container.Find(".wprm-recipe-servings").First().Text())
	recipe.Course = strings.TrimSpace(container.Find(".wprm-recipe-course-container .wprm-block-text-normal").First().Text())
	recipe.Cuisine = strings.TrimSpace(container.Find(".wprm-recipe-cuisine-container .wprm-block-text-normal").First().Text())
	recipe.Author = strings.TrimSpace(container.Find(".wprm-recipe-author-container .wprm-block-text-normal").First().Text())

	if imgSrc, exists := container.Find(".wprm-recipe-image img").First().Attr("src"); exists {
		recipe.ImageURL = imgSrc
	}

	container.Find(".wprm-recipe-ingredient").Each(func(i int, s *goquery.Selection) {
		ing := entity.Ingredient{}
		ing.Amount = strings.TrimSpace(s.Find(".wprm-recipe-ingredient-amount").Text())
		ing.Unit = strings.TrimSpace(s.Find(".wprm-recipe-ingredient-unit").Text())
		ing.Name = strings.TrimSpace(s.Find(".wprm-recipe-ingredient-name").Text())
		ing.Notes = strings.TrimSpace(s.Find(".wprm-recipe-ingredient-notes").Text())
		ing.Raw = strings.TrimSpace(s.Text())
		recipe.Ingredients = append(recipe.Ingredients, ing)
	})

	step := 1
	container.Find(".wprm-recipe-instruction").Each(func(i int, s *goquery.Selection) {
		text := strings.TrimSpace(s.Find(".wprm-recipe-instruction-text").Text())
		if text == "" {
			text = strings.TrimSpace(s.Text())
		}
		if text != "" {
			recipe.Instructions = append(recipe.Instructions, entity.Instruction{StepNumber: step, Text: text})
			step++
		}
	})

	return recipe, nil
}
