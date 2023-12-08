package limiter

import (
	"swagger/internal/appConfig"
	"sync"
	"time"
)

// 全局限流器

var zeroLimiter = Limiter{}
var globalRate *Limiter
var mu sync.Mutex

// 控制全局的限流器
func init() {
	parseGlobalRate()

	go func() {
		t := time.NewTicker(time.Second * 5)
		for {
			select {
			case <-t.C:
				if globalRate.bucket != appConfig.Info().Rate.Bucket || globalRate.perSecond != appConfig.Info().Rate.PerSecond || globalRate.timeout != appConfig.Info().Rate.WaitMillisecond {
					parseGlobalRate()
				}
			}
		}
	}()
}

func parseGlobalRate() {
	mu.Lock()
	defer mu.Unlock()

	r := appConfig.Info().Rate
	if r.PerSecond > 0 && r.Bucket > 0 {
		globalRate = NewLimiter(r.PerSecond, r.Bucket, r.WaitMillisecond)
	} else {
		globalRate = &zeroLimiter
	}
}

func Allow() bool {
	if globalRate == &zeroLimiter {
		return true
	}

	return globalRate.Allow()
}
