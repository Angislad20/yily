package handler

import (
	"net/http"
	"strings"
	"time"

	"github.com/labstack/echo/v5"
	"github.com/yannick2009/yily/internal/domain/model"
	"github.com/yannick2009/yily/internal/services"
)

const (
	SessionCookieName = "yily_session" // SessionCookieName is the HTTP cookie name for the session token.
)

// AuthHandler defines the interface for authentication HTTP handlers.
type AuthHandler interface {
	Setup(c *echo.Context) error
	Login(c *echo.Context) error
	Logout(c *echo.Context) error
	Me(c *echo.Context) error
}

// authHandler handles HTTP requests for authentication and sessions.
type authHandler struct {
	authService services.AuthService
}

// NewAuthHandler creates a new AuthHandler.
func NewAuthHandler(authService services.AuthService) AuthHandler {
	return &authHandler{authService: authService}
}

// authCredentialsRequest represents the JSON body for setup and login.
type authCredentialsRequest struct {
	Email    string `json:"email" validate:"required,email"`
	Password string `json:"password" validate:"required"`
}

// Setup handles POST /auth/setup to create the initial Super Admin account.
func (h *authHandler) Setup(c *echo.Context) error {
	var req authCredentialsRequest
	if !bindAndValidate(c, &req) {
		return nil
	}

	user, err := h.authService.Setup(c.Request().Context(), req.Email, req.Password)
	if err != nil {
		return NewHTTPResponse[any](c, httpStatus(err), err.Error(), nil)
	}

	return NewHTTPResponse(c, http.StatusCreated, "Super Admin created successfully", user)
}

// Login handles POST /auth/login to authenticate the Super Admin and set the session cookie.
func (h *authHandler) Login(c *echo.Context) error {
	var req authCredentialsRequest
	if err := c.Bind(&req); err != nil {
		return NewHTTPResponse[any](c, http.StatusBadRequest, "invalid request body", nil)
	}
	if strings.TrimSpace(req.Email) == "" || req.Password == "" {
		return NewHTTPResponse[any](c, http.StatusUnauthorized, services.ErrInvalidCredentials.Error(), nil)
	}

	var existingToken string
	if cookie, err := c.Cookie(SessionCookieName); err == nil && cookie != nil {
		existingToken = cookie.Value
	}

	rawToken, session, user, err := h.authService.Login(c.Request().Context(), req.Email, req.Password, existingToken)
	if err != nil {
		return NewHTTPResponse[any](c, httpStatus(err), err.Error(), nil)
	}

	setSessionCookie(c, rawToken, session)

	return NewHTTPResponse(c, http.StatusOK, "Login successful", user)
}

// Logout handles POST /auth/logout to revoke the session and clear the cookie.
func (h *authHandler) Logout(c *echo.Context) error {
	var rawToken string
	if cookie, err := c.Cookie(SessionCookieName); err == nil && cookie != nil {
		rawToken = cookie.Value
	}

	err := h.authService.Logout(c.Request().Context(), rawToken)
	clearSessionCookie(c)

	if err != nil {
		return NewHTTPResponse[any](c, httpStatus(err), err.Error(), nil)
	}

	return NewHTTPResponse[any](c, http.StatusOK, "Logout successful", nil)
}

// Me handles GET /auth/me to return the authenticated Super Admin from context.
func (h *authHandler) Me(c *echo.Context) error {
	user, ok := services.AuthenticatedUserFromContext(c.Request().Context())
	if !ok {
		return NewHTTPResponse[any](c, http.StatusUnauthorized, services.ErrUnauthorized.Error(), nil)
	}

	return NewHTTPResponse(c, http.StatusOK, "Authenticated user retrieved successfully", user)
}

// RequireAuth validates the session cookie or Bearer PAT and attaches the authenticated identity to the request context.
func RequireAuth(authService services.AuthService, patServices ...services.PATService) echo.MiddlewareFunc {
	var patService services.PATService
	if len(patServices) > 0 {
		patService = patServices[0]
	}

	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c *echo.Context) error {
			if patService != nil && strings.TrimSpace(c.Request().Header.Get(AuthorizationHeader)) != "" {
				return RequirePAT(patService)(next)(c)
			}

			cookie, err := c.Cookie(SessionCookieName)
			if err != nil || cookie == nil || strings.TrimSpace(cookie.Value) == "" {
				return NewHTTPResponse[any](c, http.StatusUnauthorized, services.ErrUnauthorized.Error(), nil)
			}

			user, session, err := authService.ValidateSession(c.Request().Context(), cookie.Value)
			if err != nil {
				clearSessionCookie(c)
				return NewHTTPResponse[any](c, http.StatusUnauthorized, services.ErrUnauthorized.Error(), nil)
			}

			ctx := services.WithAuthenticatedIdentity(c.Request().Context(), user, session)
			c.SetRequest(c.Request().WithContext(ctx))

			return next(c)
		}
	}
}

// setSessionCookie writes the session token to a secure HttpOnly cookie.
func setSessionCookie(c *echo.Context, rawToken string, session *model.Session) {
	maxAge := int(session.ExpiresAt.Sub(session.CreatedAt).Seconds())
	if maxAge <= 0 {
		maxAge = int(services.DefaultSessionTTL.Seconds())
	}

	http.SetCookie(c.Response(), &http.Cookie{
		Name:     SessionCookieName,
		Value:    rawToken,
		Path:     "/",
		Expires:  session.ExpiresAt,
		MaxAge:   maxAge,
		HttpOnly: true,
		Secure:   isHTTPSRequest(c),
		SameSite: http.SameSiteLaxMode,
	})
}

// clearSessionCookie expires and clears the session cookie on the client.
func clearSessionCookie(c *echo.Context) {
	http.SetCookie(c.Response(), &http.Cookie{
		Name:     SessionCookieName,
		Value:    "",
		Path:     "/",
		Expires:  time.Unix(0, 0).UTC(),
		MaxAge:   -1,
		HttpOnly: true,
		Secure:   isHTTPSRequest(c),
		SameSite: http.SameSiteLaxMode,
	})
}

// isHTTPSRequest reports whether the request uses HTTPS directly or via X-Forwarded-Proto.
func isHTTPSRequest(c *echo.Context) bool {
	if c.Request().TLS != nil {
		return true
	}

	if forwardedProto := c.Request().Header.Get("X-Forwarded-Proto"); forwardedProto != "" {
		firstProto, _, _ := strings.Cut(forwardedProto, ",")
		if strings.EqualFold(strings.TrimSpace(firstProto), "https") {
			return true
		}
	}

	return strings.EqualFold(c.Scheme(), "https")
}
