package app

import (
	"io"
	"log"
	"math/rand"
	"net/http"
	"time"
)

// retryTransport retries requests that fail with a transient status code
// (rate limiting, a Cloudflare challenge, or a server error) using capped
// exponential backoff with jitter, before giving up and returning the
// response to the caller.
type retryTransport struct {
	next       http.RoundTripper
	maxRetries int
}

var retryableStatusCodes = map[int]bool{
	http.StatusTooManyRequests:     true,
	http.StatusForbidden:           true,
	http.StatusInternalServerError: true,
	http.StatusBadGateway:          true,
	http.StatusServiceUnavailable:  true,
	http.StatusGatewayTimeout:      true,
}

func (t *retryTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	var resp *http.Response
	var err error
	for attempt := 0; ; attempt++ {
		resp, err = t.next.RoundTrip(req)
		if err != nil || !retryableStatusCodes[resp.StatusCode] || attempt >= t.maxRetries {
			return resp, err
		}
		_, _ = io.Copy(io.Discard, resp.Body)
		_ = resp.Body.Close()
		wait := time.Duration(200*(1<<attempt))*time.Millisecond + time.Duration(rand.Intn(150))*time.Millisecond
		log.Printf("retrying request after status=%d url=%s attempt=%d wait=%s", resp.StatusCode, req.URL, attempt+1, wait)
		time.Sleep(wait)
	}
}
