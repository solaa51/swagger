package limiter

import (
	"sync"
)

type KeyRateLimiter struct {
	keys map[string]*Limiter
	mu   *sync.RWMutex
	r    float64 // 速率
	b    int     // 容量
	t    int     // 等待毫秒数
}

func NewKeyRateLimiter(r float64, b int, t int) *KeyRateLimiter {
	return &KeyRateLimiter{
		keys: make(map[string]*Limiter),
		mu:   &sync.RWMutex{},
		r:    r,
		b:    b,
		t:    t,
	}
}

func (k *KeyRateLimiter) addKey(key string) *Limiter {
	k.mu.Lock()
	defer k.mu.Unlock()

	limiter := NewLimiter(k.r, k.b, k.t)
	k.keys[key] = limiter

	return limiter
}

func (k *KeyRateLimiter) GetLimiter(key string) *Limiter {
	k.mu.Lock()
	limiter, has := k.keys[key]

	if !has {
		k.mu.Unlock()
		return k.addKey(key)
	}

	k.mu.Unlock()
	return limiter
}
