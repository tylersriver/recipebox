package handler

import (
	"bytes"
	"context"
	"fmt"
	"net/http"
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
	svc *appservice.RecipeService
}

func NewRecipeHandler(svc *appservice.RecipeService) *RecipeHandler {
	return &RecipeHandler{svc: svc}
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
	if err := h.svc.DeleteRecipe(c.Request().Context(), userID, command.DeleteRecipeCommand{ID: id}); err != nil {
		if strings.Contains(err.Error(), "not found") {
			return c.String(http.StatusNotFound, "recipe not found")
		}
		return c.String(http.StatusInternalServerError, err.Error())
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

func renderSSEFragment(sse *datastar.ServerSentEventGenerator, component templ.Component) error {
	var buf bytes.Buffer
	if err := component.Render(context.Background(), &buf); err != nil {
		return err
	}
	sse.PatchElements(buf.String())
	return nil
}
