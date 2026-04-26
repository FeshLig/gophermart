package middleware_test

import (
	"compress/gzip"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/FeshLig/gophermart/internal/middleware"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/require"
)

func TestGzipMiddleware(t *testing.T) {
	gin.SetMode(gin.TestMode)

	t.Run("without gzip support", func(t *testing.T) {
		r := gin.New()
		r.Use(middleware.Gzip())

		r.GET("/", func(c *gin.Context) {
			c.String(http.StatusOK, "hello")
		})

		req := httptest.NewRequest(http.MethodGet, "/", nil)

		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)

		require.Equal(t, http.StatusOK, w.Code)
		require.Empty(t, w.Header().Get("Content-Encoding"))
		require.Equal(t, "hello", w.Body.String())
	})

	t.Run("with gzip support", func(t *testing.T) {
		r := gin.New()
		r.Use(middleware.Gzip())

		r.GET("/", func(c *gin.Context) {
			c.String(http.StatusOK, "hello")
		})

		req := httptest.NewRequest(http.MethodGet, "/", nil)
		req.Header.Set("Accept-Encoding", "gzip")

		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)

		require.Equal(t, http.StatusOK, w.Code)
		require.Equal(t, "gzip", w.Header().Get("Content-Encoding"))

		gr, err := gzip.NewReader(w.Body)
		require.NoError(t, err)
		defer gr.Close()

		body, err := io.ReadAll(gr)
		require.NoError(t, err)

		require.Equal(t, "hello", string(body))
	})
}
