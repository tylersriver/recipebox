package handler

import (
	"net/http"

	"github.com/labstack/echo/v4"
	"github.com/tyler/recipebox/internal/application/query"
	appservice "github.com/tyler/recipebox/internal/application/service"
	"github.com/tyler/recipebox/internal/interface/web/template"
)

type HomeHandler struct {
	svc *appservice.RecipeService
}

func NewHomeHandler(svc *appservice.RecipeService) *HomeHandler {
	return &HomeHandler{svc: svc}
}

func (h *HomeHandler) Home(c echo.Context) error {
	result, err := h.svc.ListRecipes(c.Request().Context(), query.ListRecipesQuery{Limit: 6})
	if err != nil {
		return c.String(http.StatusInternalServerError, err.Error())
	}

	return Render(c, http.StatusOK, template.Home(result.Recipes))
}
