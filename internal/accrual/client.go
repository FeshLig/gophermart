package accrual

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"
)

type Client struct {
	baseURL string
	client  *http.Client
}

func NewClient(baseURL string, timeout time.Duration) *Client {
	return &Client{
		baseURL: baseURL,
		client: &http.Client{
			Timeout: timeout,
		},
	}
}

type Response struct {
	Order   string   `json:"order"`
	Status  string   `json:"status"`
	Accrual *float64 `json:"accrual,omitempty"`
}

func (c *Client) GetOrder(ctx context.Context, number string) (*Response, int, http.Header, error) {
	url := fmt.Sprintf("%s/api/orders/%s", c.baseURL, number)

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, 0, nil, err
	}

	resp, err := c.doWithRetry(req)
	if err != nil {
		return nil, 0, nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusNoContent {
		return nil, resp.StatusCode, resp.Header, nil
	}

	if resp.StatusCode != http.StatusOK {
		return nil, resp.StatusCode, resp.Header, nil
	}

	var result Response
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, resp.StatusCode, resp.Header, err
	}

	return &result, resp.StatusCode, resp.Header, nil
}

func (c *Client) doWithRetry(req *http.Request) (*http.Response, error) {
	const maxRetries = 3

	var resp *http.Response
	var err error

	for i := 0; i < maxRetries; i++ {
		resp, err = c.client.Do(req)

		if err == nil && resp.StatusCode < 500 {
			return resp, nil
		}

		if resp != nil {
			resp.Body.Close()
		}

		time.Sleep(time.Duration(i+1) * 100 * time.Millisecond)
	}

	return resp, err
}
