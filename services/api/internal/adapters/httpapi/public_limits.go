package httpapi

import (
	"net"
	"net/http"
	"sync"
	"time"
)

type publicBucket struct {
	start time.Time
	count int
}
type publicLimiter struct {
	mu      sync.Mutex
	clients map[string]publicBucket
	global  publicBucket
}

func (l *publicLimiter) allow(key string, now time.Time) bool {
	l.mu.Lock()
	defer l.mu.Unlock()
	if now.Sub(l.global.start) >= time.Minute {
		l.global = publicBucket{start: now}
		for key, b := range l.clients {
			if now.Sub(b.start) >= time.Minute {
				delete(l.clients, key)
			}
		}
	}
	if l.global.count >= 400 {
		return false
	}
	l.global.count++
	if l.clients == nil {
		l.clients = map[string]publicBucket{}
	}
	b := l.clients[key]
	if now.Sub(b.start) >= time.Minute {
		b = publicBucket{start: now}
	}
	if b.count >= 20 {
		return false
	}
	b.count++
	l.clients[key] = b
	return true
}
func (l *publicLimiter) guard(w http.ResponseWriter, r *http.Request, origin string) bool {
	if r.URL.Path != "/v1/public/appointment-requests" && r.URL.Path != "/v1/smart-cards/verify" {
		return true
	}
	if r.Method != "GET" && r.Header.Get("Origin") != "" && r.Header.Get("Origin") != origin {
		write(w, 403, map[string]string{"code": "ORIGIN_DENIED", "error": "Request origin is not allowed"})
		return false
	}
	host, _, err := net.SplitHostPort(r.RemoteAddr)
	if err != nil {
		host = r.RemoteAddr
	}
	// Do not trust caller-controlled forwarding headers. A reverse proxy should
	// enforce its own shared client quota before forwarding to this local API.
	if !l.allow(host, time.Now()) {
		w.Header().Set("Retry-After", "60")
		write(w, 429, map[string]string{"code": "RATE_LIMITED", "error": "Too many requests; try again later"})
		return false
	}
	return true
}
