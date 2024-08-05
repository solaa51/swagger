package singleFlight

import (
	"context"
	"golang.org/x/sync/singleflight"
	"time"
)

type SingleFlight struct {
	g singleflight.Group
}

func (s *SingleFlight) SingleRun(ctx context.Context, key string, method func() (any, error)) (any, error) {
	result := s.g.DoChan(key, func() (interface{}, error) {
		return method()
	})

	//防止 一个出错，全部出错
	go func() {
		time.Sleep(100 * time.Millisecond)
		s.g.Forget(key)
	}()

	select {
	case r := <-result:
		return r.Val, r.Err
	case <-ctx.Done():
		return nil, ctx.Err()
	}
}

var defaultGroup *SingleFlight

func init() {
	defaultGroup = &SingleFlight{}
}

func SingleRun(ctx context.Context, key string, method func() (any, error)) (any, error) {
	return defaultGroup.SingleRun(ctx, key, method)
}
