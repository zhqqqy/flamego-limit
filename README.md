# flamego-limit
flamego's current limiting middleware
# Getting started
```go
func main() {
	f := flamego.Classic()
	f.Use(Limiter(Options{
		Max:        10, 
		Expiration: 5 * time.Second,
	}))
	f.Get("/", func(c flamego.Context, limit Limit) string {
		l := limit.DoLimit(c.Request().Host)
		if l(c) {
			c.ResponseWriter().WriteHeader(http.StatusTooManyRequests)
			return "Too Many Requests"
    }
	return "Hello, Flamego!"
	})
	f.Run()
}
```