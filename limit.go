package limit

import (
	"github.com/flamego/flamego"
	"time"
)

var (
	defaultMax        = 10
	defaultExpiration = 5 * time.Second
)

type Limit interface {
	DoLimit(key string) handle
}
type Options struct {
	Max        int
	Expiration time.Duration
}

func Limiter(opts ...Options) flamego.Handler {
	var opt Options
	if len(opts) > 0 {
		opt = opts[0]
	}

	parseOptions := func(opts Options) Options {
		if opts.Max == 0 {
			opts.Max = defaultMax
		}

		if opts.Expiration == 0 {
			opts.Expiration = defaultExpiration
		}
		return opts
	}
	opt = parseOptions(opt)
	l := newLimit(opt)

	return flamego.ContextInvoker(func(c flamego.Context) {
		c.Map(l)
	})
}
