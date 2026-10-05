package message

import (
	"context"
	"time"
)

// Middleware 装饰 Handler，用于织入日志、重试、panic 恢复等横切逻辑。
type Middleware func(Handler) Handler

// Chain 按顺序包装 Handler：Chain(h, a, b) 的执行顺序为 a → b → h。
func Chain(h Handler, mws ...Middleware) Handler {
	if h == nil {
		return nil
	}
	for i := len(mws) - 1; i >= 0; i-- {
		if mws[i] == nil {
			continue
		}
		h = mws[i](h)
	}
	return h
}

// RecoverMW 捕获 handler 中的 panic，避免单个消息打挂整个消费者。
func RecoverMW(onPanic func(ctx context.Context, msg *Message, v any)) Middleware {
	return func(next Handler) Handler {
		return func(ctx context.Context, msg *Message) (err error) {
			defer func() {
				if v := recover(); v != nil {
					if onPanic != nil {
						onPanic(ctx, msg, v)
					}
				}
			}()
			return next(ctx, msg)
		}
	}
}

// RetryMW 处理失败时重试。backoff 为 nil 时使用固定 100ms 退避。
// maxRetry <= 0 时不做任何包装。
func RetryMW(maxRetry int, backoff func(attempt int) time.Duration) Middleware {
	if maxRetry <= 0 {
		return func(next Handler) Handler { return next }
	}
	if backoff == nil {
		backoff = func(int) time.Duration { return 100 * time.Millisecond }
	}
	return func(next Handler) Handler {
		return func(ctx context.Context, msg *Message) error {
			var err error
			for attempt := 0; attempt <= maxRetry; attempt++ {
				err = next(ctx, msg)
				if err == nil {
					return nil
				}
				if attempt == maxRetry {
					break
				}
				select {
				case <-ctx.Done():
					return ctx.Err()
				case <-time.After(backoff(attempt)):
				}
			}
			return err
		}
	}
}
