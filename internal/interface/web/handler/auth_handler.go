package handler

import (
	"net/http"

	"github.com/gorilla/sessions"
	"github.com/labstack/echo/v4"
	"github.com/tyler/recipebox/internal/application/command"
	appservice "github.com/tyler/recipebox/internal/application/service"
	"github.com/tyler/recipebox/internal/interface/web/middleware"
	tmpl "github.com/tyler/recipebox/internal/interface/web/template"
)

type AuthHandler struct {
	authSvc *appservice.AuthService
	store   sessions.Store
}

func NewAuthHandler(authSvc *appservice.AuthService, store sessions.Store) *AuthHandler {
	return &AuthHandler{authSvc: authSvc, store: store}
}

func (h *AuthHandler) LoginPage(c echo.Context) error {
	return Render(c, http.StatusOK, tmpl.Login(""))
}

func (h *AuthHandler) LoginSubmit(c echo.Context) error {
	email := c.FormValue("email")
	password := c.FormValue("password")

	if email == "" || password == "" {
		return Render(c, http.StatusOK, tmpl.Login("Email and password are required"))
	}

	user, err := h.authSvc.Login(c.Request().Context(), command.LoginCommand{
		Email:    email,
		Password: password,
	})
	if err != nil {
		return Render(c, http.StatusOK, tmpl.Login("Invalid email or password"))
	}

	if err := middleware.SetSession(c, h.store, user.ID, user.Email); err != nil {
		return Render(c, http.StatusOK, tmpl.Login("Failed to create session"))
	}

	return c.Redirect(http.StatusSeeOther, "/")
}

func (h *AuthHandler) RegisterPage(c echo.Context) error {
	return Render(c, http.StatusOK, tmpl.Register(""))
}

func (h *AuthHandler) RegisterSubmit(c echo.Context) error {
	email := c.FormValue("email")
	password := c.FormValue("password")
	confirm := c.FormValue("confirm")

	if email == "" || password == "" {
		return Render(c, http.StatusOK, tmpl.Register("Email and password are required"))
	}

	if password != confirm {
		return Render(c, http.StatusOK, tmpl.Register("Passwords do not match"))
	}

	if len(password) < 8 {
		return Render(c, http.StatusOK, tmpl.Register("Password must be at least 8 characters"))
	}

	user, err := h.authSvc.Register(c.Request().Context(), command.RegisterCommand{
		Email:    email,
		Password: password,
	})
	if err != nil {
		return Render(c, http.StatusOK, tmpl.Register(err.Error()))
	}

	if err := middleware.SetSession(c, h.store, user.ID, user.Email); err != nil {
		return Render(c, http.StatusOK, tmpl.Register("Failed to create session"))
	}

	return c.Redirect(http.StatusSeeOther, "/")
}

func (h *AuthHandler) Logout(c echo.Context) error {
	middleware.ClearSession(c, h.store)
	return c.Redirect(http.StatusSeeOther, "/login")
}
