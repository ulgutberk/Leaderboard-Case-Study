package middleware

import (
    "context"
    "net/http"
    "strings"
    "time"

    "github.com/go-redis/redis/v8"
)

// Atomic fixed-window counter via Lua (INCR + EXPIRE).
// Returns 1 if allowed, 0 if limit exceeded.
var rateLimitScript = redis.NewScript(`
local key    = KEYS[1]
local limit  = tonumber(ARGV[1])
local window = tonumber(ARGV[2])
local n = redis.call('INCR', key)
if n == 1 then
    redis.call('EXPIRE', key, window)
end
if n > limit then
    return 0
end
return 1
`)

// RateLimiter limits each IP to `limit` requests per `window`.
// Backed by Redis — works across multiple service instances.
func RateLimiter(rdb *redis.Client, limit int, window time.Duration) func(http.Handler) http.Handler {
    windowSecs := int(window.Seconds())
    return func(next http.Handler) http.Handler {
        return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
            ip := clientIP(r)
            allowed, err := rateLimitScript.Run(
                context.Background(), rdb,
                []string{"rate_limit:" + ip},
                limit, windowSecs,
            ).Int()
            if err != nil {
                // Redis hiccup — fail open, don't block traffic.
                next.ServeHTTP(w, r)
                return
            }
            if allowed == 0 {
                w.Header().Set("Content-Type", "application/json")
                w.Header().Set("Retry-After", "60")
                w.WriteHeader(http.StatusTooManyRequests)
                w.Write([]byte(`{"error":"rate limit exceeded"}`))
                return
            }
            next.ServeHTTP(w, r)
        })
    }
}

// clientIP returns the real client IP, honouring X-Forwarded-For.
func clientIP(r *http.Request) string {
    if xff := r.Header.Get("X-Forwarded-For"); xff != "" {
        if i := strings.Index(xff, ","); i != -1 {
            return strings.TrimSpace(xff[:i])
        }
        return strings.TrimSpace(xff)
    }
    addr := r.RemoteAddr
    if i := strings.LastIndex(addr, ":"); i != -1 {
        return addr[:i]
    }
    return addr
}