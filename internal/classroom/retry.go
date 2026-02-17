package classroom

import (
	"io"
	"math/rand"
	"net/http"
	"strconv"
	"time"
)

type retryTransport struct {
	base       http.RoundTripper
	maxRetries int
}

func newRetryTransport(base http.RoundTripper) http.RoundTripper {
	if base == nil {
		base = http.DefaultTransport
	}
	return &retryTransport{base: base, maxRetries: 4}
}

func (t *retryTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	var lastResp *http.Response
	var lastErr error

	for attempt := 0; attempt <= t.maxRetries; attempt++ {
		reqCopy := req.Clone(req.Context())
		if attempt > 0 && req.GetBody != nil {
			body, err := req.GetBody()
			if err != nil {
				return nil, err
			}
			reqCopy.Body = body
		}
		resp, err := t.base.RoundTrip(reqCopy)
		if !shouldRetry(resp, err, attempt, t.maxRetries) {
			return resp, err
		}
		if resp != nil && resp.Body != nil {
			_, _ = io.Copy(io.Discard, resp.Body)
			_ = resp.Body.Close()
		}
		lastResp = resp
		lastErr = err
		time.Sleep(backoff(resp, attempt))
	}
	return lastResp, lastErr
}

func shouldRetry(resp *http.Response, err error, attempt, max int) bool {
	if attempt >= max {
		return false
	}
	if err != nil {
		return true
	}
	if resp == nil {
		return true
	}
	if resp.StatusCode == http.StatusTooManyRequests {
		return true
	}
	return resp.StatusCode >= 500
}

func backoff(resp *http.Response, attempt int) time.Duration {
	if resp != nil {
		retryAfter := resp.Header.Get("Retry-After")
		if retryAfter != "" {
			if sec, err := strconv.Atoi(retryAfter); err == nil && sec > 0 {
				return time.Duration(sec) * time.Second
			}
		}
	}
	base := 300 * time.Millisecond
	mult := 1 << attempt
	jitter := time.Duration(rand.Intn(200)) * time.Millisecond
	d := time.Duration(mult)*base + jitter
	max := 8 * time.Second
	if d > max {
		return max
	}
	return d
}
