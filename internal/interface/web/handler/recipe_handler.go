package handler

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/a-h/templ"
	"github.com/labstack/echo/v4"
	"github.com/starfederation/datastar-go/datastar"
	"github.com/tyler/recipebox/internal/application/command"
	"github.com/tyler/recipebox/internal/application/query"
	appservice "github.com/tyler/recipebox/internal/application/service"
	"github.com/tyler/recipebox/internal/interface/web/middleware"
	tmpl "github.com/tyler/recipebox/internal/interface/web/template"
)

type RecipeHandler struct {
	svc        *appservice.RecipeService
	uploadsDir string
}

func NewRecipeHandler(svc *appservice.RecipeService, uploadsDir string) *RecipeHandler {
	return &RecipeHandler{svc: svc, uploadsDir: uploadsDir}
}

func (h *RecipeHandler) List(c echo.Context) error {
	userID := middleware.GetUserID(c)
	offset, _ := strconv.Atoi(c.QueryParam("offset"))
	limit, _ := strconv.Atoi(c.QueryParam("limit"))
	if limit <= 0 {
		limit = 12
	}

	result, err := h.svc.ListRecipes(c.Request().Context(), userID, query.ListRecipesQuery{
		Offset: offset,
		Limit:  limit,
	})
	if err != nil {
		return c.String(http.StatusInternalServerError, err.Error())
	}

	return Render(c, http.StatusOK, tmpl.RecipeList(*result, middleware.GetUserEmail(c)))
}

func (h *RecipeHandler) Detail(c echo.Context) error {
	userID := middleware.GetUserID(c)
	id := c.Param("id")
	result, err := h.svc.GetRecipe(c.Request().Context(), userID, query.GetRecipeByIDQuery{ID: id})
	if err != nil {
		return c.String(http.StatusNotFound, "Recipe not found")
	}

	return Render(c, http.StatusOK, tmpl.RecipeDetail(*result, middleware.GetUserEmail(c)))
}

func (h *RecipeHandler) CreatePage(c echo.Context) error {
	return Render(c, http.StatusOK, tmpl.RecipeCreate(middleware.GetUserEmail(c)))
}

func (h *RecipeHandler) CreateSubmit(c echo.Context) error {
	userID := middleware.GetUserID(c)

	if err := c.Request().ParseMultipartForm(maxImageSize); err != nil {
		if err2 := c.Request().ParseForm(); err2 != nil {
			return c.String(http.StatusBadRequest, "Invalid form data")
		}
	}

	title := strings.TrimSpace(c.FormValue("title"))
	if title == "" {
		return Render(c, http.StatusUnprocessableEntity, tmpl.RecipeCreate(middleware.GetUserEmail(c)))
	}

	var ingredients []command.IngredientInput
	for _, raw := range c.Request().Form["ingredients"] {
		raw = strings.TrimSpace(raw)
		if raw != "" {
			ingredients = append(ingredients, command.IngredientInput{Raw: raw})
		}
	}

	var instructions []string
	for _, text := range c.Request().Form["instructions"] {
		text = strings.TrimSpace(text)
		if text != "" {
			instructions = append(instructions, text)
		}
	}

	// Handle optional file upload
	var imagePath string
	file, header, err := c.Request().FormFile("image")
	if err == nil {
		defer file.Close()
		imagePath, err = saveUploadedImage(file, header, "", h.uploadsDir)
		if err != nil {
			return c.String(http.StatusBadRequest, "Image upload failed: "+err.Error())
		}
	}

	cmd := command.CreateRecipeCommand{
		Title:        title,
		Description:  strings.TrimSpace(c.FormValue("description")),
		PrepTime:     strings.TrimSpace(c.FormValue("prep_time")),
		CookTime:     strings.TrimSpace(c.FormValue("cook_time")),
		TotalTime:    strings.TrimSpace(c.FormValue("total_time")),
		Servings:     strings.TrimSpace(c.FormValue("servings")),
		Cuisine:      strings.TrimSpace(c.FormValue("cuisine")),
		Course:       strings.TrimSpace(c.FormValue("course")),
		ImageURL:     strings.TrimSpace(c.FormValue("image_url")),
		ImagePath:    imagePath,
		Ingredients:  ingredients,
		Instructions: instructions,
	}

	result, err := h.svc.CreateRecipe(c.Request().Context(), userID, cmd)
	if err != nil {
		return c.String(http.StatusInternalServerError, "Failed to create recipe: "+err.Error())
	}

	return c.Redirect(http.StatusSeeOther, "/recipes/"+result.ID)
}

func (h *RecipeHandler) ImportPage(c echo.Context) error {
	return Render(c, http.StatusOK, tmpl.RecipeImport(middleware.GetUserEmail(c)))
}

type importSignals struct {
	URL       string `json:"url"`
	Importing bool   `json:"importing"`
}

func (h *RecipeHandler) ImportSubmit(c echo.Context) error {
	userID := middleware.GetUserID(c)

	// ReadSignals MUST be called before NewSSE (SSE consumes the body)
	var signals importSignals
	if err := datastar.ReadSignals(c.Request(), &signals); err != nil {
		sse := datastar.NewSSE(c.Response().Writer, c.Request())
		return renderSSEFragment(sse, tmpl.ImportError("Invalid request"))
	}

	sse := datastar.NewSSE(c.Response().Writer, c.Request())

	if signals.URL == "" {
		return renderSSEFragment(sse, tmpl.ImportError("Please enter a URL"))
	}

	// Set importing state
	sse.MarshalAndPatchSignals(importSignals{URL: signals.URL, Importing: true})

	result, err := h.svc.ImportRecipe(c.Request().Context(), userID, command.ImportRecipeCommand{URL: signals.URL})
	if err != nil {
		sse.MarshalAndPatchSignals(importSignals{URL: signals.URL, Importing: false})
		return renderSSEFragment(sse, tmpl.ImportError("Failed to import: "+err.Error()))
	}

	sse.MarshalAndPatchSignals(importSignals{URL: "", Importing: false})
	return renderSSEFragment(sse, tmpl.ImportSuccess(result.Title, result.ID))
}

func (h *RecipeHandler) Delete(c echo.Context) error {
	userID := middleware.GetUserID(c)
	id := c.Param("id")

	// Fetch recipe first to get image path for cleanup
	result, _ := h.svc.GetRecipe(c.Request().Context(), userID, query.GetRecipeByIDQuery{ID: id})

	if err := h.svc.DeleteRecipe(c.Request().Context(), userID, command.DeleteRecipeCommand{ID: id}); err != nil {
		if strings.Contains(err.Error(), "not found") {
			return c.String(http.StatusNotFound, "recipe not found")
		}
		return c.String(http.StatusInternalServerError, err.Error())
	}

	// Clean up uploaded image file (best-effort)
	if result != nil && result.ImagePath != "" {
		os.Remove(filepath.Join(h.uploadsDir, result.ImagePath))
	}

	return c.Redirect(http.StatusSeeOther, "/recipes")
}

func (h *RecipeHandler) UpdateNotes(c echo.Context) error {
	userID := middleware.GetUserID(c)
	id := c.Param("id")
	notes := c.FormValue("notes")
	if err := h.svc.UpdateNotes(c.Request().Context(), userID, command.UpdateNotesCommand{ID: id, Notes: notes}); err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
	}
	return c.JSON(http.StatusOK, map[string]string{"message": "Notes saved"})
}

type browseSearchSignals struct {
	Search string `json:"search"`
}

func (h *RecipeHandler) Search(c echo.Context) error {
	userID := middleware.GetUserID(c)

	var signals browseSearchSignals
	if err := datastar.ReadSignals(c.Request(), &signals); err != nil {
		return err
	}

	sse := datastar.NewSSE(c.Response().Writer, c.Request())

	if signals.Search == "" {
		result, err := h.svc.ListRecipes(c.Request().Context(), userID, query.ListRecipesQuery{
			Offset: 0,
			Limit:  12,
		})
		if err != nil {
			return err
		}
		return renderSSEFragment(sse, tmpl.BrowseResults(*result))
	}

	result, err := h.svc.SearchRecipes(c.Request().Context(), userID, query.SearchRecipesQuery{
		Query: signals.Search,
		Limit: 30,
	})
	if err != nil {
		return err
	}
	return renderSSEFragment(sse, tmpl.BrowseSearchResults(result.Recipes))
}

func (h *RecipeHandler) SharedDetail(c echo.Context) error {
	token := c.Param("token")
	result, err := h.svc.GetSharedRecipe(c.Request().Context(), token)
	if err != nil {
		return c.String(http.StatusNotFound, "Recipe not found")
	}

	return Render(c, http.StatusOK, tmpl.SharedRecipeDetail(*result))
}

func (h *RecipeHandler) GenerateShareLink(c echo.Context) error {
	userID := middleware.GetUserID(c)
	id := c.Param("id")
	result, err := h.svc.ShareRecipe(c.Request().Context(), userID, command.ShareRecipeCommand{ID: id})
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
	}

	scheme := "https"
	if c.Request().TLS == nil && c.Request().Header.Get("X-Forwarded-Proto") == "" {
		scheme = "http"
	}
	shareURL := fmt.Sprintf("%s://%s/share/%s", scheme, c.Request().Host, result.ShareToken)
	return c.JSON(http.StatusOK, map[string]string{"url": shareURL, "token": result.ShareToken})
}

func (h *RecipeHandler) UploadImage(c echo.Context) error {
	userID := middleware.GetUserID(c)
	recipeID := c.Param("id")

	file, header, err := c.Request().FormFile("image")
	if err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "No image file provided"})
	}
	defer file.Close()

	filename, err := saveUploadedImage(file, header, recipeID, h.uploadsDir)
	if err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": err.Error()})
	}

	if err := h.svc.UploadImage(c.Request().Context(), userID, command.UploadImageCommand{
		RecipeID: recipeID,
		Filename: filename,
	}); err != nil {
		os.Remove(filepath.Join(h.uploadsDir, filename))
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
	}

	return c.Redirect(http.StatusSeeOther, "/recipes/"+recipeID)
}

var allowedImageTypes = map[string]string{
	"image/jpeg": ".jpg",
	"image/png":  ".png",
	"image/webp": ".webp",
}

const maxImageSize = 10 << 20 // 10 MB

func saveUploadedImage(file multipart.File, header *multipart.FileHeader, name string, uploadsDir string) (string, error) {
	if header.Size > maxImageSize {
		return "", fmt.Errorf("image file too large (max 10 MB)")
	}

	// Detect content type from file bytes
	buf := make([]byte, 512)
	n, err := file.Read(buf)
	if err != nil && err != io.EOF {
		return "", fmt.Errorf("reading file: %w", err)
	}
	contentType := http.DetectContentType(buf[:n])

	ext, ok := allowedImageTypes[contentType]
	if !ok {
		return "", fmt.Errorf("unsupported image type: %s (allowed: JPEG, PNG, WebP)", contentType)
	}

	// Seek back to beginning
	if seeker, ok := file.(io.Seeker); ok {
		seeker.Seek(0, io.SeekStart)
	}

	// Use provided name or generate from header filename
	if name == "" {
		base := strings.TrimSuffix(header.Filename, filepath.Ext(header.Filename))
		name = base
	}

	// Remove any old files for this name with different extensions
	for _, oldExt := range allowedImageTypes {
		if oldExt != ext {
			os.Remove(filepath.Join(uploadsDir, name+oldExt))
		}
	}

	filename := name + ext

	if err := os.MkdirAll(uploadsDir, 0755); err != nil {
		return "", fmt.Errorf("creating uploads directory: %w", err)
	}

	dst, err := os.Create(filepath.Join(uploadsDir, filename))
	if err != nil {
		return "", fmt.Errorf("creating file: %w", err)
	}
	defer dst.Close()

	if _, err := io.Copy(dst, file); err != nil {
		return "", fmt.Errorf("saving file: %w", err)
	}

	return filename, nil
}

func renderSSEFragment(sse *datastar.ServerSentEventGenerator, component templ.Component) error {
	var buf bytes.Buffer
	if err := component.Render(context.Background(), &buf); err != nil {
		return err
	}
	sse.PatchElements(buf.String())
	return nil
}
