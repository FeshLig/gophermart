package middleware_test

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/FeshLig/gophermart/internal/middleware"
	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v4"
	"github.com/stretchr/testify/require"
)

func setupRouter(secret string) *gin.Engine {
	gin.SetMode(gin.TestMode)

	r := gin.New()

	r.Use(middleware.AuthMiddleware(func() string {
		return secret
	}))

	r.GET("/test", func(c *gin.Context) {
		userID, exists := c.Get("userID")
		if !exists {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "no user"})
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"userID": userID,
		})
	})

	return r
}

func generateToken(secret string, userID int64) string {
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"user_id": userID,
		"exp":     time.Now().Add(time.Hour).Unix(),
	})

	str, _ := token.SignedString([]byte(secret))
	return str
}

func TestAuthMiddleware_Success_Bearer(t *testing.T) {
	secret := "test-secret"
	router := setupRouter(secret)

	token := generateToken(secret, 42)

	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	req.Header.Set("Authorization", "Bearer "+token)

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	require.Equal(t, http.StatusOK, w.Code)
	require.Contains(t, w.Body.String(), "42")
}

func TestAuthMiddleware_Success_Cookie(t *testing.T) {
	secret := "test-secret"
	router := setupRouter(secret)

	token := generateToken(secret, 99)

	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	req.AddCookie(&http.Cookie{
		Name:  "token",
		Value: token,
	})

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	require.Equal(t, http.StatusOK, w.Code)
	require.Contains(t, w.Body.String(), "99")
}

func TestAuthMiddleware_NoToken(t *testing.T) {
	router := setupRouter("secret")

	req := httptest.NewRequest(http.MethodGet, "/test", nil)

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	require.Equal(t, http.StatusUnauthorized, w.Code)
}

func TestAuthMiddleware_InvalidToken(t *testing.T) {
	router := setupRouter("secret")

	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	req.Header.Set("Authorization", "Bearer invalid.token.here")

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	require.Equal(t, http.StatusUnauthorized, w.Code)
}

func TestAuthMiddleware_WrongSecret(t *testing.T) {
	secret := "correct"
	router := setupRouter(secret)

	token := generateToken("wrong", 1)

	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	req.Header.Set("Authorization", "Bearer "+token)

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	require.Equal(t, http.StatusUnauthorized, w.Code)
}

func TestAuthMiddleware_NoUserID(t *testing.T) {
	secret := "test-secret"
	router := setupRouter(secret)

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"exp": time.Now().Add(time.Hour).Unix(),
	})

	tokenStr, _ := token.SignedString([]byte(secret))

	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	req.Header.Set("Authorization", "Bearer "+tokenStr)

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	require.Equal(t, http.StatusUnauthorized, w.Code)
}

func TestAuthMiddleware_InvalidUserIDType(t *testing.T) {
	secret := "test-secret"
	router := setupRouter(secret)

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"user_id": "not-a-number",
		"exp":     time.Now().Add(time.Hour).Unix(),
	})

	tokenStr, _ := token.SignedString([]byte(secret))

	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	req.Header.Set("Authorization", "Bearer "+tokenStr)

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	require.Equal(t, http.StatusUnauthorized, w.Code)
}
