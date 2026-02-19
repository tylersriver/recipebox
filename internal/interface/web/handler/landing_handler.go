package handler

import (
	"net/http"

	"github.com/gorilla/sessions"
	"github.com/labstack/echo/v4"
	"github.com/tyler/recipebox/internal/interface/web/template"
)

type LandingHandler struct {
	store sessions.Store
}

func NewLandingHandler(store sessions.Store) *LandingHandler {
	return &LandingHandler{store: store}
}

func (h *LandingHandler) Index(c echo.Context) error {
	sess, err := h.store.Get(c.Request(), "recipebox")
	if err == nil {
		if uid, ok := sess.Values["user_id"].(string); ok && uid != "" {
			return c.Redirect(http.StatusSeeOther, "/home")
		}
	}
	return Render(c, http.StatusOK, template.Landing())
}
