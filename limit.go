package main

import (
	"net"
	"net/http"
	"sync"
	"time"

	"golang.org/x/time/rate"
)

type ip struct {
	limiter  *rate.Limiter
	lastSeen time.Time
}

var ips = make(map[string]*ip)
var mu sync.Mutex

func init() {
	go cleanUp()
}

// Limit IP's => 20 requests per minute
func getIP(ipAddress string) *rate.Limiter {
	mu.Lock()
	defer mu.Unlock()

	// insert ip into map if not present
	value, ok := ips[ipAddress]
	if !ok {
		rt := rate.Every(time.Minute)
		limiter := rate.NewLimiter(rt, 20)

		ips[ipAddress] = &ip{limiter, time.Now()}
		return limiter
	}

	// Udpate last seen
	value.lastSeen = time.Now()
	return value.limiter
}

// If ip hasn't been seen for over a minute, remove the map entry
func cleanUp() {
	for {
		time.Sleep(time.Minute)

		mu.Lock()
		for i, value := range ips {
			if time.Since(value.lastSeen) > time.Minute {
				delete(ips, i)
			}
		}
		mu.Unlock()
	}
}

// ip address rate limiting
func limit(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ip, _, err := net.SplitHostPort(r.RemoteAddr)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		limiter := getIP(ip)
		if !limiter.Allow() {
			http.Error(w, http.StatusText(429), http.StatusTooManyRequests)
			return
		}
		next.ServeHTTP(w, r)
	}
}
