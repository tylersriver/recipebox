package handler

import (
	"bytes"
	"context"
	"net/http"
	"strconv"

	"github.com/a-h/templ"
	"github.com/labstack/echo/v4"
	"github.com/starfederation/datastar-go/datastar"
	"github.com/tyler/recipebox/internal/application/command"
	"github.com/tyler/recipebox/internal/application/query"
	appservice "github.com/tyler/recipebox/internal/application/service"
	tmpl "github.com/tyler/recipebox/internal/interface/web/template"
)

type RecipeHandler struct {
	svc *appservice.RecipeService
}

func NewRecipeHandler(svc *appservice.RecipeService) *RecipeHandler {
	return &RecipeHandler{svc: svc}
}

func (h *RecipeHandler) List(c echo.Context) error {
	offset, _ := strconv.Atoi(c.QueryParam("offset"))
	limit, _ := strconv.Atoi(c.QueryParam("limit"))
	if limit <= 0 {
		limit = 12
	}

	result, err := h.svc.ListRecipes(c.Request().Context(), query.ListRecipesQuery{
		Offset: offset,
		Limit:  limit,
	})
	if err != nil {
		return c.String(http.StatusInternalServerError, err.Error())
	}

	return Render(c, http.StatusOK, tmpl.RecipeList(*result))
}

func (h *RecipeHandler) Detail(c echo.Context) error {
	id := c.Param("id")
	result, err := h.svc.GetRecipe(c.Request().Context(), query.GetRecipeByIDQuery{ID: id})
	if err != nil {
		return c.String(http.StatusNotFound, "Recipe not found")
	}

	return Render(c, http.StatusOK, tmpl.RecipeDetail(*result))
}

func (h *RecipeHandler) ImportPage(c echo.Context) error {
	return Render(c, http.StatusOK, tmpl.RecipeImport())
}

type importSignals struct {
	URL       string `json:"url"`
	Importing bool   `json:"importing"`
}

func (h *RecipeHandler) ImportSubmit(c echo.Context) error {
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

	result, err := h.svc.ImportRecipe(c.Request().Context(), command.ImportRecipeCommand{URL: signals.URL})
	if err != nil {
		sse.MarshalAndPatchSignals(importSignals{URL: signals.URL, Importing: false})
		return renderSSEFragment(sse, tmpl.ImportError("Failed to import: "+err.Error()))
	}

	sse.MarshalAndPatchSignals(importSignals{URL: "", Importing: false})
	return renderSSEFragment(sse, tmpl.ImportSuccess(result.Title, result.ID))
}

func (h *RecipeHandler) Delete(c echo.Context) error {
	id := c.Param("id")
	if err := h.svc.DeleteRecipe(c.Request().Context(), command.DeleteRecipeCommand{ID: id}); err != nil {
		return c.String(http.StatusInternalServerError, err.Error())
	}
	return c.Redirect(http.StatusSeeOther, "/recipes")
}

func (h *RecipeHandler) UpdateNotes(c echo.Context) error {
	id := c.Param("id")
	notes := c.FormValue("notes")
	if err := h.svc.UpdateNotes(c.Request().Context(), command.UpdateNotesCommand{ID: id, Notes: notes}); err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
	}
	return c.JSON(http.StatusOK, map[string]string{"message": "Notes saved"})
}

type browseSearchSignals struct {
	Search string `json:"search"`
}

func (h *RecipeHandler) Search(c echo.Context) error {
	var signals browseSearchSignals
	if err := datastar.ReadSignals(c.Request(), &signals); err != nil {
		return err
	}

	sse := datastar.NewSSE(c.Response().Writer, c.Request())

	if signals.Search == "" {
		result, err := h.svc.ListRecipes(c.Request().Context(), query.ListRecipesQuery{
			Offset: 0,
			Limit:  12,
		})
		if err != nil {
			return err
		}
		return renderSSEFragment(sse, tmpl.BrowseResults(*result))
	}

	result, err := h.svc.SearchRecipes(c.Request().Context(), query.SearchRecipesQuery{
		Query: signals.Search,
		Limit: 30,
	})
	if err != nil {
		return err
	}
	return renderSSEFragment(sse, tmpl.BrowseSearchResults(result.Recipes))
}

func renderSSEFragment(sse *datastar.ServerSentEventGenerator, component templ.Component) error {
	var buf bytes.Buffer
	if err := component.Render(context.Background(), &buf); err != nil {
		return err
	}
	sse.PatchElements(buf.String())
	return nil
}
