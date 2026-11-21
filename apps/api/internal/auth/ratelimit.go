package auth

import (
	"net"
	"net/http"
	"sync"
	"time"

	"golang.org/x/time/rate"
)

const visitorExpiry = 3 * time.Minute

type visitor struct {
	limiter  *rate.Limiter
	lastSeen time.Time
}

// RateLimiter enforces a per-client-IP request rate. It's in-memory and
// per-instance, consistent with this API's other in-memory state (veto
// sessions): fine for a single-instance deployment, not for a scaled-out one.
type RateLimiter struct {
	mu       sync.Mutex
	visitors map[string]*visitor
	rate     rate.Limit
	burst    int
}

func NewRateLimiter(requestsPerSecond float64, burst int) *RateLimiter {
	rl := &RateLimiter{
		visitors: make(map[string]*visitor),
		rate:     rate.Limit(requestsPerSecond),
		burst:    burst,
	}
	go rl.evictExpiredVisitors()
	return rl
}

func (rl *RateLimiter) Middleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if !rl.allow(clientIP(r)) {
			http.Error(w, "rate limit exceeded", http.StatusTooManyRequests)
			return
		}
		next.ServeHTTP(w, r)
	})
}

func (rl *RateLimiter) allow(key string) bool {
	rl.mu.Lock()
	v, exists := rl.visitors[key]
	if !exists {
		v = &visitor{limiter: rate.NewLimiter(rl.rate, rl.burst)}
		rl.visitors[key] = v
	}
	v.lastSeen = time.Now()
	limiter := v.limiter
	rl.mu.Unlock()

	return limiter.Allow()
}

func (rl *RateLimiter) evictExpiredVisitors() {
	for range time.Tick(time.Minute) {
		rl.mu.Lock()
		for key, v := range rl.visitors {
			if time.Since(v.lastSeen) > visitorExpiry {
				delete(rl.visitors, key)
			}
		}
		rl.mu.Unlock()
	}
}

// clientIP uses the direct TCP peer address, not X-Forwarded-For: this API
// isn't behind a reverse proxy yet, and trusting that header without one
// would let a client spoof its way around the limit.
func clientIP(r *http.Request) string {
	host, _, err := net.SplitHostPort(r.RemoteAddr)
	if err != nil {
		return r.RemoteAddr
	}
	return host
}
