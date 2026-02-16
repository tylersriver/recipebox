package handler

import (
	"bytes"
	"context"

	"github.com/labstack/echo/v4"
	"github.com/starfederation/datastar-go/datastar"
	"github.com/tyler/recipebox/internal/application/query"
	appservice "github.com/tyler/recipebox/internal/application/service"
	"github.com/tyler/recipebox/internal/interface/web/middleware"
	tmpl "github.com/tyler/recipebox/internal/interface/web/template"
)

type SearchHandler struct {
	svc *appservice.RecipeService
}

func NewSearchHandler(svc *appservice.RecipeService) *SearchHandler {
	return &SearchHandler{svc: svc}
}

type searchSignals struct {
	Search string `json:"search"`
}

func (h *SearchHandler) Results(c echo.Context) error {
	userID := middleware.GetUserID(c)

	// ReadSignals before NewSSE
	var signals searchSignals
	if err := datastar.ReadSignals(c.Request(), &signals); err != nil {
		return err
	}

	sse := datastar.NewSSE(c.Response().Writer, c.Request())

	if signals.Search == "" {
		var buf bytes.Buffer
		if err := tmpl.SearchResultsEmpty().Render(context.Background(), &buf); err != nil {
			return err
		}
		sse.PatchElements(buf.String())
		return nil
	}

	result, err := h.svc.SearchRecipes(c.Request().Context(), userID, query.SearchRecipesQuery{
		Query: signals.Search,
		Limit: 10,
	})
	if err != nil {
		return err
	}

	var buf bytes.Buffer
	if err := tmpl.SearchResults(result.Recipes).Render(context.Background(), &buf); err != nil {
		return err
	}
	sse.PatchElements(buf.String())
	return nil
}
