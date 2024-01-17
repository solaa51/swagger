package limiter

import (
	"context"
	"golang.org/x/time/rate"
	"time"
)

type Limiter struct {
	r         *rate.Limiter
	perSecond float64
	bucket    int
	timeout   int
}

// Allow 不等待 直接返回是否允许通过
func (l *Limiter) Allow() bool {
	if l.timeout == 0 {
		return l.r.Allow()
	}

	//允许等待超时时间 按毫秒计算
	if !l.r.Allow() {
		ctx, cancel := context.WithTimeout(context.Background(), time.Millisecond*(time.Duration(l.timeout)))
		defer cancel()
		if err := l.r.Wait(ctx); err != nil {
			return false
		}
	}

	return true
}

// NewLimiter r为放入桶中的速率/秒 cap位桶容量
//
// r 每秒流入桶内的数量
//
// cap 桶容量
//
// timeout 等待时长[毫秒]
func NewLimiter(r float64, cap int, timeout int) *Limiter {
	rr := rate.Limit(r)
	return &Limiter{
		r:         rate.NewLimiter(rr, cap),
		perSecond: r,
		bucket:    cap,
		timeout:   timeout,
	}
}
