package web

import (
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	appservice "github.com/tyler/recipebox/internal/application/service"
	"github.com/tyler/recipebox/internal/interface/web/handler"
)

type Server struct {
	echo *echo.Echo
	addr string
}

func NewServer(svc *appservice.RecipeService, addr string) *Server {
	e := echo.New()
	e.HideBanner = true

	e.Use(middleware.Logger())
	e.Use(middleware.Recover())

	homeHandler := handler.NewHomeHandler(svc)
	recipeHandler := handler.NewRecipeHandler(svc)
	searchHandler := handler.NewSearchHandler(svc)

	e.GET("/", homeHandler.Home)
	e.GET("/recipes", recipeHandler.List)
	e.GET("/recipes/search", recipeHandler.Search)
	e.GET("/recipes/:id", recipeHandler.Detail)
	e.GET("/import", recipeHandler.ImportPage)
	e.POST("/import/submit", recipeHandler.ImportSubmit)
	e.POST("/recipes/:id/notes", recipeHandler.UpdateNotes)
	e.POST("/recipes/:id/delete", recipeHandler.Delete)
	e.GET("/search/results", searchHandler.Results)

	return &Server{echo: e, addr: addr}
}

func (s *Server) Start() error {
	return s.echo.Start(s.addr)
}
