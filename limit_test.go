package limit

import (
	"bytes"
	"fmt"
	"github.com/flamego/flamego"
	"github.com/stretchr/testify/assert"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestLimit(t *testing.T) {
	f := flamego.NewWithLogger(&bytes.Buffer{})
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

	singleRequest := func(shouldFail bool) {
		req, err := http.NewRequest(http.MethodGet, "/", nil)
		resp := httptest.NewRecorder()
		f.ServeHTTP(resp, req)
		if shouldFail {
			assert.Equal(t, nil, err)
			assert.Equal(t, http.StatusTooManyRequests, resp.Code)
		} else {
			assert.Equal(t, nil, err)
			assert.Equal(t, http.StatusOK, resp.Code)

		}
	}

	for i := 0; i < 10; i++ {
		singleRequest(false)
	}
	singleRequest(true)
	fmt.Println("================================================================================================================================================================")
	time.Sleep(6500 * time.Millisecond)
	for i := 0; i < 4; i++ {
		singleRequest(false)
	}
	singleRequest(true)
}
