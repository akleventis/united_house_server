package middleware

import (
	"net"
	"net/http"
	"sync"
	"time"

	"golang.org/x/time/rate"
)

const (
	RL5   = 5
	RL10  = 10
	RL30  = 30
	RL50  = 50
	RL100 = 100
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

// Limit IP's: rl = requests per minute
func getIP(ipAddress string, rl int) *rate.Limiter {
	mu.Lock()
	defer mu.Unlock()

	// insert ip into map if not present
	value, ok := ips[ipAddress]
	if !ok {
		rt := rate.Every(time.Minute)
		limiter := rate.NewLimiter(rt, rl)

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
func Limit(next http.HandlerFunc, rl int) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ip, _, err := net.SplitHostPort(r.RemoteAddr)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		limiter := getIP(ip, rl)
		if !limiter.Allow() {
			http.Error(w, http.StatusText(429), http.StatusTooManyRequests)
			return
		}
		next.ServeHTTP(w, r)
	}
}
