package ingress

type concurrencyLimiter struct {
	n  int
	ch chan struct{}
}

func newLimiter(n int) *concurrencyLimiter {
	if n <= 0 {
		n = 1
	}
	ch := make(chan struct{}, n)
	return &concurrencyLimiter{n: n, ch: ch}
}

func (c *concurrencyLimiter) acquire() bool {
	select {
	case c.ch <- struct{}{}:
		return true
	default:
	}
	return false
}

func (c *concurrencyLimiter) release() {
	<-c.ch
}
