package handler

import (
	"net/http"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/labstack/echo/v5"
	"github.com/yannick2009/yily/internal/services"
)

const (
	AuthorizationHeader = "Authorization" // AuthorizationHeader is the HTTP header carrying the Bearer PAT.
	BearerScheme        = "Bearer"        // BearerScheme is the authentication scheme prefix for PATs.
)

// PATHandler defines the interface for Personal Access Token HTTP handlers.
type PATHandler interface {
	Create(c *echo.Context) error
	GetAll(c *echo.Context) error
	GetByID(c *echo.Context) error
	Revoke(c *echo.Context) error
}

// patHandler handles HTTP requests for Personal Access Tokens.
type patHandler struct {
	patService services.PATService
}

// NewPATHandler creates a new PATHandler.
func NewPATHandler(patService services.PATService) PATHandler {
	return &patHandler{patService: patService}
}

// createPATRequest represents the JSON body for creating a PAT.
type createPATRequest struct {
	Name          string `json:"name" validate:"required"`
	ExpiresInDays *int   `json:"expires_in_days,omitempty"`
}

// CreatePATResponse represents the one-time response returned when a PAT is created.
type CreatePATResponse struct {
	ID        uuid.UUID `json:"id"`
	UserID    uuid.UUID `json:"user_id"`
	Name      string    `json:"name"`
	Token     string    `json:"token"`
	CreatedAt time.Time `json:"created_at"`
	ExpiresAt time.Time `json:"expires_at"`
}

// Create handles POST /pats to generate a new PAT and return the raw token once.
func (h *patHandler) Create(c *echo.Context) error {
	user, ok := services.AuthenticatedUserFromContext(c.Request().Context())
	if !ok {
		return NewHTTPResponse[any](c, http.StatusUnauthorized, services.ErrUnauthorized.Error(), nil)
	}

	var req createPATRequest
	if !bindAndValidate(c, &req) {
		return nil
	}

	var ttl time.Duration
	if req.ExpiresInDays != nil {
		if *req.ExpiresInDays <= 0 {
			return NewHTTPResponse[any](c, http.StatusBadRequest, services.ErrInvalidPATExpiration.Error(), nil)
		}
		ttl = time.Duration(*req.ExpiresInDays) * 24 * time.Hour
	}

	rawToken, pat, err := h.patService.Create(c.Request().Context(), user.ID, req.Name, ttl)
	if err != nil {
		return NewHTTPResponse[any](c, httpStatus(err), err.Error(), nil)
	}

	resp := CreatePATResponse{
		ID:        pat.ID,
		UserID:    pat.UserID,
		Name:      pat.Name,
		Token:     rawToken,
		CreatedAt: pat.CreatedAt,
		ExpiresAt: pat.ExpiresAt,
	}

	return NewHTTPResponse(c, http.StatusCreated, "Personal access token created successfully", resp)
}

// GetAll handles GET /pats to list the authenticated user's PAT metadata without raw tokens.
func (h *patHandler) GetAll(c *echo.Context) error {
	user, ok := services.AuthenticatedUserFromContext(c.Request().Context())
	if !ok {
		return NewHTTPResponse[any](c, http.StatusUnauthorized, services.ErrUnauthorized.Error(), nil)
	}

	pats, err := h.patService.GetAll(c.Request().Context(), user.ID)
	if err != nil {
		return NewHTTPResponse[any](c, httpStatus(err), err.Error(), nil)
	}

	return NewHTTPResponse(c, http.StatusOK, "Personal access tokens retrieved successfully", pats)
}

// GetByID handles GET /pats/:id to retrieve a single PAT's metadata without the raw token.
func (h *patHandler) GetByID(c *echo.Context) error {
	user, ok := services.AuthenticatedUserFromContext(c.Request().Context())
	if !ok {
		return NewHTTPResponse[any](c, http.StatusUnauthorized, services.ErrUnauthorized.Error(), nil)
	}

	id, ok := uuidParam(c, "id")
	if !ok {
		return nil
	}

	pat, err := h.patService.GetByID(c.Request().Context(), user.ID, id)
	if err != nil {
		return NewHTTPResponse[any](c, httpStatus(err), err.Error(), nil)
	}

	return NewHTTPResponse(c, http.StatusOK, "Personal access token retrieved successfully", pat)
}

// Revoke handles DELETE /pats/:id to revoke a PAT belonging to the authenticated user.
func (h *patHandler) Revoke(c *echo.Context) error {
	user, ok := services.AuthenticatedUserFromContext(c.Request().Context())
	if !ok {
		return NewHTTPResponse[any](c, http.StatusUnauthorized, services.ErrUnauthorized.Error(), nil)
	}

	id, ok := uuidParam(c, "id")
	if !ok {
		return nil
	}

	if err := h.patService.Revoke(c.Request().Context(), user.ID, id); err != nil {
		return NewHTTPResponse[any](c, httpStatus(err), err.Error(), nil)
	}

	return NewHTTPResponse[any](c, http.StatusOK, "Personal access token revoked successfully", nil)
}

// RequirePAT validates a Bearer PAT from the Authorization header and attaches the identity to context.
func RequirePAT(patService services.PATService) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c *echo.Context) error {
			rawToken, ok := extractBearerToken(c.Request().Header.Get(AuthorizationHeader))
			if !ok {
				return NewHTTPResponse[any](c, http.StatusUnauthorized, services.ErrUnauthorized.Error(), nil)
			}

			user, pat, err := patService.Validate(c.Request().Context(), rawToken)
			if err != nil {
				return NewHTTPResponse[any](c, http.StatusUnauthorized, services.ErrUnauthorized.Error(), nil)
			}

			ctx := services.WithAuthenticatedPAT(c.Request().Context(), user, pat)
			c.SetRequest(c.Request().WithContext(ctx))

			return next(c)
		}
	}
}

// extractBearerToken extracts the raw token from an "Authorization: Bearer <token>" header.
func extractBearerToken(authHeader string) (string, bool) {
	authHeader = strings.TrimSpace(authHeader)
	if authHeader == "" {
		return "", false
	}

	scheme, token, found := strings.Cut(authHeader, " ")
	if !found || !strings.EqualFold(scheme, BearerScheme) {
		return "", false
	}

	token = strings.TrimSpace(token)
	if token == "" || strings.Contains(token, " ") {
		return "", false
	}

	return token, true
}

