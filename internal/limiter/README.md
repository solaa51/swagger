限流器

```go
var limiter = NewIPRateLimiter(1, 5, 0)

rateLimiter := limiter.GetLimiter(r.RemoteAddr)
if !rateLimiter.Allow() {
http.Error(w, http.StatusText(http.StatusTooManyRequests), http.StatusTooManyRequests)
return
}
```