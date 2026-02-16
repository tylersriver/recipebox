package middleware

import (
	"net/http"

	"github.com/gorilla/sessions"
	"github.com/labstack/echo/v4"
)

const (
	sessionName    = "recipebox"
	sessionUserID  = "user_id"
	sessionEmail   = "email"
	contextUserID  = "user_id"
	contextEmail   = "user_email"
)

// RequireAuth returns Echo middleware that enforces authentication.
func RequireAuth(store sessions.Store) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			sess, err := store.Get(c.Request(), sessionName)
			if err != nil {
				return c.Redirect(http.StatusSeeOther, "/login")
			}

			userID, ok := sess.Values[sessionUserID].(string)
			if !ok || userID == "" {
				return c.Redirect(http.StatusSeeOther, "/login")
			}

			email, _ := sess.Values[sessionEmail].(string)
			c.Set(contextUserID, userID)
			c.Set(contextEmail, email)
			return next(c)
		}
	}
}

// GetUserID returns the authenticated user's ID from the Echo context.
func GetUserID(c echo.Context) string {
	id, _ := c.Get(contextUserID).(string)
	return id
}

// GetUserEmail returns the authenticated user's email from the Echo context.
func GetUserEmail(c echo.Context) string {
	email, _ := c.Get(contextEmail).(string)
	return email
}

// SetSession stores user info in the session.
func SetSession(c echo.Context, store sessions.Store, userID, email string) error {
	sess, err := store.Get(c.Request(), sessionName)
	if err != nil {
		sess, _ = store.New(c.Request(), sessionName)
	}
	sess.Values[sessionUserID] = userID
	sess.Values[sessionEmail] = email
	return sess.Save(c.Request(), c.Response().Writer)
}

// ClearSession removes user info from the session.
func ClearSession(c echo.Context, store sessions.Store) error {
	sess, err := store.Get(c.Request(), sessionName)
	if err != nil {
		return nil
	}
	sess.Options.MaxAge = -1
	return sess.Save(c.Request(), c.Response().Writer)
}
