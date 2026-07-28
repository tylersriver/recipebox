package web

import (
	"database/sql"

	"github.com/gorilla/sessions"
	"github.com/labstack/echo/v4"
	echomw "github.com/labstack/echo/v4/middleware"
	appservice "github.com/tyler/recipebox/internal/application/service"
	infrarepo "github.com/tyler/recipebox/internal/infrastructure/repository"
	"github.com/tyler/recipebox/internal/interface/web/handler"
	"github.com/tyler/recipebox/internal/interface/web/middleware"
)

type Server struct {
	echo *echo.Echo
	addr string
}

func NewServer(svc *appservice.RecipeService, db *sql.DB, addr string, sessionSecret string) *Server {
	e := echo.New()
	e.HideBanner = true

	e.Use(echomw.Logger())
	e.Use(echomw.Recover())
	e.Static("/static", "static")

	// PWA: service worker must be served from the root so it can control
	// the entire site scope. The manifest can stay under /static/.
	e.File("/sw.js", "static/sw.js")

	store := sessions.NewCookieStore([]byte(sessionSecret))

	userRepo := infrarepo.NewSQLiteUserRepository(db)
	authSvc := appservice.NewAuthService(userRepo)

	landingHandler := handler.NewLandingHandler(store)
	authHandler := handler.NewAuthHandler(authSvc, store)
	homeHandler := handler.NewHomeHandler(svc)
	recipeHandler := handler.NewRecipeHandler(svc)
	searchHandler := handler.NewSearchHandler(svc)

	// Public routes
	e.GET("/", landingHandler.Index)
	e.GET("/login", authHandler.LoginPage)
	e.POST("/login", authHandler.LoginSubmit)
	e.GET("/register", authHandler.RegisterPage)
	e.POST("/register", authHandler.RegisterSubmit)
	e.POST("/logout", authHandler.Logout)
	e.GET("/share/:token", recipeHandler.SharedDetail)

	// Protected routes
	auth := e.Group("", middleware.RequireAuth(store))
	auth.GET("/home", homeHandler.Home)
	auth.GET("/recipes", recipeHandler.List)
	auth.GET("/recipes/create", recipeHandler.CreatePage)
	auth.POST("/recipes/create", recipeHandler.CreateSubmit)
	auth.GET("/recipes/search", recipeHandler.Search)
	auth.GET("/recipes/:id", recipeHandler.Detail)
	auth.GET("/recipes/:id/edit", recipeHandler.EditPage)
	auth.POST("/recipes/:id/edit", recipeHandler.EditSubmit)
	auth.GET("/import", recipeHandler.ImportPage)
	auth.POST("/import/submit", recipeHandler.ImportSubmit)
	auth.POST("/recipes/:id/notes", recipeHandler.UpdateNotes)
	auth.POST("/recipes/:id/delete", recipeHandler.Delete)
	auth.POST("/recipes/:id/share", recipeHandler.GenerateShareLink)
	auth.GET("/search/results", searchHandler.Results)

	return &Server{echo: e, addr: addr}
}

func (s *Server) Start() error {
	return s.echo.Start(s.addr)
}
