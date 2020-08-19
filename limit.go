package main

import (
	"golang.org/x/time/rate"
	"net"
	"net/http"
	"sync"
	"time"
)

// Create a custom visitor struct which holds the rate limiter for each
// visitor and the last time that the visitor was seen.
type visitor struct {
	limiter  *rate.Limiter
	lastSeen time.Time
}

// Change the the map to hold values of the type visitor.
var visitors = make(map[string]*visitor)
var mu sync.Mutex

// Run a background goroutine to remove old entries from the visitors map.
func InitRatelimiter() {
	go cleanupVisitors()
}

func getVisitor(ip string) *rate.Limiter {
	mu.Lock()
	defer mu.Unlock()

	v, exists := visitors[ip]
	if !exists {
		limiter := rate.NewLimiter(1, 2)
		// Include the current time when creating a new visitor.
		visitors[ip] = &visitor{limiter, time.Now()}
		return limiter
	}

	// Update the last seen time for the visitor.
	v.lastSeen = time.Now()
	return v.limiter
}

// Every minute check the map for visitors that haven't been seen for
// more than 3 minutes and delete the entries.
func cleanupVisitors() {
	for {
		time.Sleep(time.Minute)

		mu.Lock()
		for ip, v := range visitors {
			if time.Since(v.lastSeen) > 3*time.Minute {
				delete(visitors, ip)
			}
		}
		mu.Unlock()
	}
}

func limit(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var ip string
		// handle nginx reverse proxy
		forwardHeader := r.Header.Get("x-forwarded-for")
		if forwardHeader != "" {
			ip = forwardHeader
		} else {
			ip2, _, err := net.SplitHostPort(r.RemoteAddr)
			if err != nil {
				SendJSONResponse(&w, ResponseError{
					Status:  1,
					Message: "Could not fetch IP address.",
				}, http.StatusInternalServerError)
				return
			}
			ip = ip2
		}

		limiter := getVisitor(ip)
		if limiter.Allow() == false {
			SendJSONResponse(&w, ResponseError{
				Status:  1,
				Message: "You are being rate limited.",
			}, http.StatusTooManyRequests)
			return
		}

		next.ServeHTTP(w, r)
	})
}
