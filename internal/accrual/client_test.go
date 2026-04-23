package accrual_test

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/FeshLig/gophermart/internal/accrual"
	"github.com/stretchr/testify/require"
)

func TestGetOrder_Success(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		require.Equal(t, "/api/orders/123", r.URL.Path)

		w.WriteHeader(http.StatusOK)
		_ = json.NewEncoder(w).Encode(accrual.Response{
			Order:  "123",
			Status: "PROCESSED",
		})
	}))
	defer server.Close()

	client := accrual.NewClient(server.URL, time.Second)

	resp, code, err := client.GetOrder(context.Background(), "123")

	require.NoError(t, err)
	require.Equal(t, http.StatusOK, code)
	require.NotNil(t, resp)
	require.Equal(t, "123", resp.Order)
	require.Equal(t, "PROCESSED", resp.Status)
}

func TestGetOrder_NoContent(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNoContent)
	}))
	defer server.Close()

	client := accrual.NewClient(server.URL, time.Second)

	resp, code, err := client.GetOrder(context.Background(), "123")

	require.NoError(t, err)
	require.Equal(t, http.StatusNoContent, code)
	require.Nil(t, resp)
}

func TestGetOrder_ErrorStatus(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusBadRequest)
	}))
	defer server.Close()

	client := accrual.NewClient(server.URL, time.Second)

	resp, code, err := client.GetOrder(context.Background(), "123")

	require.NoError(t, err)
	require.Equal(t, http.StatusBadRequest, code)
	require.Nil(t, resp)
}

func TestGetOrder_DecodeError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("not-json"))
	}))
	defer server.Close()

	client := accrual.NewClient(server.URL, time.Second)

	resp, code, err := client.GetOrder(context.Background(), "123")

	require.Error(t, err)
	require.Equal(t, http.StatusOK, code)
	require.Nil(t, resp)
}

func TestGetOrder_Timeout(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		time.Sleep(200 * time.Millisecond)
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	client := accrual.NewClient(server.URL, 50*time.Millisecond)

	ctx := context.Background()

	_, _, err := client.GetOrder(ctx, "123")

	require.Error(t, err)
}

func TestRetryLogic(t *testing.T) {
	attempts := 0

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		attempts++

		// первые 2 раза 500, потом успех
		if attempts < 3 {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		w.WriteHeader(http.StatusOK)
		_ = json.NewEncoder(w).Encode(accrual.Response{
			Order:  "123",
			Status: "PROCESSED",
		})
	}))
	defer server.Close()

	client := accrual.NewClient(server.URL, time.Second)

	resp, code, err := client.GetOrder(context.Background(), "123")

	require.NoError(t, err)
	require.Equal(t, http.StatusOK, code)
	require.NotNil(t, resp)
	require.Equal(t, 3, attempts) // проверка retry
}

func TestRetryLogic_FailAll(t *testing.T) {
	attempts := 0

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		attempts++
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer server.Close()

	client := accrual.NewClient(server.URL, time.Second)

	resp, code, err := client.GetOrder(context.Background(), "123")

	require.NoError(t, err)
	require.Equal(t, http.StatusInternalServerError, code)
	require.Nil(t, resp)
	require.Equal(t, 3, attempts)
}

func TestGetOrder_RequestCreationError(t *testing.T) {
	client := accrual.NewClient("http://invalid-url-%", time.Second)

	_, _, err := client.GetOrder(context.Background(), "123")

	require.Error(t, err)
}

func TestGetOrder_ContextCancel(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		time.Sleep(200 * time.Millisecond)
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	client := accrual.NewClient(server.URL, time.Second)

	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	_, _, err := client.GetOrder(ctx, "123")

	require.Error(t, err)
}
