package app

import (
	"io"
	"log"
	"math/rand"
	"net/http"
	"sync"
	"time"
)

// retryTransport retries requests that fail with a transient status code
// (rate limiting, a Cloudflare challenge, or a server error) using capped
// exponential backoff with jitter, before giving up and returning the
// response to the caller. It also enforces a minimum spacing between ALL
// requests it sends, regardless of which colly.Collector (or other caller)
// issued them: colly's own per-collector LimitRule only throttles requests
// made through that one collector instance, but makeCollector builds a new
// collector for every scrape call, so most of our request volume (e.g.
// looking up TMDB ids for many watchlist films in a row) would otherwise go
// out back-to-back with no delay at all.
type retryTransport struct {
	next       http.RoundTripper
	maxRetries int

	minInterval time.Duration
	mu          sync.Mutex
	earliest    time.Time // earliest time the next request through this transport may start
}

var retryableStatusCodes = map[int]bool{
	http.StatusTooManyRequests:     true,
	http.StatusForbidden:           true,
	http.StatusInternalServerError: true,
	http.StatusBadGateway:          true,
	http.StatusServiceUnavailable:  true,
	http.StatusGatewayTimeout:      true,
}

// waitTurn blocks until this request is allowed to start, spacing every
// request sent through the transport by minInterval-2*minInterval.
func (t *retryTransport) waitTurn() {
	t.mu.Lock()
	now := time.Now()
	wait := time.Duration(0)
	if t.earliest.After(now) {
		wait = t.earliest.Sub(now)
	}
	jitter := time.Duration(rand.Int63n(int64(t.minInterval) + 1))
	t.earliest = now.Add(wait).Add(t.minInterval + jitter)
	t.mu.Unlock()
	if wait > 0 {
		time.Sleep(wait)
	}
}

func (t *retryTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	var resp *http.Response
	var err error
	for attempt := 0; ; attempt++ {
		t.waitTurn()
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
