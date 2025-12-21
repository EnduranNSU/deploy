package httpin

import (
	"encoding/json"
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"

	"github.com/EnduranNSU/trainings/internal/adapter/in/http/dto"
)

// AuthMiddleware ходит в Auth-сервис и валидирует access-токен.
type AuthMiddleware struct {
	client   *http.Client
	authBase string
}

func NewAuthMiddleware(authBase string) *AuthMiddleware {
	return &AuthMiddleware{
		client: &http.Client{
			Timeout: 2 * time.Second,
		},
		authBase: strings.TrimRight(authBase, "/"),
	}
}

type validateResponse struct {
	UserID string `json:"user_id"`
}

func (m *AuthMiddleware) Handle(c *gin.Context) {
	h := c.GetHeader("Authorization")
	if len(h) < 7 || !strings.EqualFold(h[:7], "bearer ") {
		c.AbortWithStatusJSON(http.StatusUnauthorized, dto.ErrorResponse{Error: "no_bearer"})
		return
	}

	req, err := http.NewRequestWithContext(
		c.Request.Context(),
		http.MethodGet,
		m.authBase+"/api/v1/validate",
		nil,
	)
	if err != nil {
		c.AbortWithStatusJSON(http.StatusUnauthorized, dto.ErrorResponse{Error: "auth_unavailable"})
		return
	}
	req.Header.Set("Authorization", h)

	resp, err := m.client.Do(req)
	if err != nil {
		c.AbortWithStatusJSON(http.StatusUnauthorized, dto.ErrorResponse{Error: "auth_unavailable"})
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		c.AbortWithStatusJSON(http.StatusUnauthorized, dto.ErrorResponse{Error: "invalid_token"})
		return
	}

	var body validateResponse
	if err := json.NewDecoder(resp.Body).Decode(&body); err != nil || body.UserID == "" {
		c.AbortWithStatusJSON(http.StatusUnauthorized, dto.ErrorResponse{Error: "bad_auth_response"})
		return
	}

	c.Set("userID", body.UserID)

	c.Next()
}
