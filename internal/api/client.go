package api

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"

	"golang.org/x/oauth2"
	"google.golang.org/api/googleapi"
)

const (
	baseURL      = "https://classroom.googleapis.com/v1"
	defaultRetry = 3
	initialDelay = time.Second
	maxDelay     = 32 * time.Second
)

type Client struct {
	httpClient  *http.Client
	tokenSource oauth2.TokenSource
	retries     int
	backoff     time.Duration
}

type Option func(*Client)

func WithRetries(n int) Option {
	return func(c *Client) {
		c.retries = n
	}
}

func WithBackoff(d time.Duration) Option {
	return func(c *Client) {
		c.backoff = d
	}
}

func NewClient(ctx context.Context, ts oauth2.TokenSource, opts ...Option) (*Client, error) {
	httpClient := oauth2.NewClient(ctx, ts)

	client := &Client{
		httpClient:  httpClient,
		tokenSource: ts,
		retries:     defaultRetry,
		backoff:     initialDelay,
	}

	for _, opt := range opts {
		opt(client)
	}

	return client, nil
}

func NewClientFromToken(ctx context.Context, cfg *oauth2.Config, token *oauth2.Token, opts ...Option) (*Client, error) {
	ts := cfg.TokenSource(ctx, token)
	return NewClient(ctx, ts, opts...)
}

type APIError struct {
	Code    int    `json:"code,omitempty"`
	Message string `json:"message,omitempty"`
	Status  string `json:"status,omitempty"`
}

func (e *APIError) Error() string {
	return fmt.Sprintf("API error %d: %s (%s)", e.Code, e.Message, e.Status)
}

func IsNotFound(err error) bool {
	var apiErr *APIError
	if errors.As(err, &apiErr) {
		return apiErr.Code == 404 || apiErr.Status == "NOT_FOUND"
	}
	var gae *googleapi.Error
	if errors.As(err, &gae) {
		return gae.Code == http.StatusNotFound
	}
	return false
}

func IsForbidden(err error) bool {
	var apiErr *APIError
	if errors.As(err, &apiErr) {
		return apiErr.Code == 403 || apiErr.Status == "PERMISSION_DENIED"
	}
	var gae *googleapi.Error
	if errors.As(err, &gae) {
		return gae.Code == http.StatusForbidden
	}
	return false
}

func IsRateLimited(err error) bool {
	var apiErr *APIError
	if errors.As(err, &apiErr) {
		return apiErr.Code == 429 || apiErr.Status == "RESOURCE_EXHAUSTED"
	}
	var gae *googleapi.Error
	if errors.As(err, &gae) {
		return gae.Code == http.StatusTooManyRequests
	}
	return false
}

func (c *Client) parseError(resp *http.Response) error {
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("failed to read response body: %w", err)
	}

	var apiErr APIError
	if err := json.Unmarshal(body, &apiErr); err != nil {
		return fmt.Errorf("request failed with status %d: %s", resp.StatusCode, string(body))
	}

	if apiErr.Code == 0 {
		apiErr.Code = resp.StatusCode
	}

	if apiErr.Message == "" {
		apiErr.Message = string(body)
	}

	return &apiErr
}

func (c *Client) doRequest(ctx context.Context, method, url string, body io.Reader) (*http.Response, error) {
	req, err := http.NewRequestWithContext(ctx, method, url, body)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("request failed: %w", err)
	}

	return resp, nil
}

func (c *Client) doRequestWithRetry(ctx context.Context, method, url string, body io.Reader) (*http.Response, error) {
	var lastErr error
	backoff := c.backoff

	for i := 0; i <= c.retries; i++ {
		resp, err := c.doRequest(ctx, method, url, body)
		if err != nil {
			return nil, err
		}

		if resp.StatusCode == 429 {
			resp.Body.Close()
			if i < c.retries {
				select {
				case <-ctx.Done():
					return nil, ctx.Err()
				case <-time.After(backoff):
					backoff *= 2
					if backoff > maxDelay {
						backoff = maxDelay
					}
					continue
				}
			}
			return nil, c.parseError(resp)
		}

		if resp.StatusCode < 400 {
			return resp, nil
		}

		if resp.StatusCode == 404 || resp.StatusCode == 403 {
			return resp, c.parseError(resp)
		}

		if resp.StatusCode >= 500 {
			resp.Body.Close()
			if i < c.retries {
				select {
				case <-ctx.Done():
					return nil, ctx.Err()
				case <-time.After(backoff):
					backoff *= 2
					if backoff > maxDelay {
						backoff = maxDelay
					}
					continue
				}
			}
		}

		return resp, c.parseError(resp)
	}

	return nil, lastErr
}

func (c *Client) get(ctx context.Context, endpoint string, params url.Values) ([]byte, error) {
	url := baseURL + endpoint
	if len(params) > 0 {
		url += "?" + params.Encode()
	}

	resp, err := c.doRequestWithRetry(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		return nil, c.parseError(resp)
	}

	return io.ReadAll(resp.Body)
}

func (c *Client) patch(ctx context.Context, endpoint string, params url.Values, body []byte) ([]byte, error) {
	url := baseURL + endpoint
	if len(params) > 0 {
		url += "?" + params.Encode()
	}

	resp, err := c.doRequestWithRetry(ctx, http.MethodPatch, url, strings.NewReader(string(body)))
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		return nil, c.parseError(resp)
	}

	return io.ReadAll(resp.Body)
}

type ListResponse struct {
	NextPageToken string          `json:"nextPageToken"`
	Coursework    json.RawMessage `json:"courseWork,omitempty"`
	Courses       json.RawMessage `json:"courses,omitempty"`
	Submissions   json.RawMessage `json:"studentSubmissions,omitempty"`
	Announcements json.RawMessage `json:"announcements,omitempty"`
}

func parsePageToken(resp []byte) string {
	var lr ListResponse
	if err := json.Unmarshal(resp, &lr); err != nil {
		return ""
	}
	return lr.NextPageToken
}

func buildParams(pairs ...string) url.Values {
	params := url.Values{}
	for i := 0; i < len(pairs)-1; i += 2 {
		if pairs[i+1] != "" {
			params.Set(pairs[i], pairs[i+1])
		}
	}
	return params
}

func buildListParams(pageSize int, pageToken string) url.Values {
	params := url.Values{}
	if pageSize > 0 {
		params.Set("pageSize", strconv.Itoa(pageSize))
	}
	if pageToken != "" {
		params.Set("pageToken", pageToken)
	}
	return params
}
