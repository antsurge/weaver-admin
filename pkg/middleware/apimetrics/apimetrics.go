// Package apimetrics 采集 HTTP 接口请求指标（请求量/失败数/耗时），
// 数据写入 pkg/metrics 的进程内采集器，供「接口监控」页面展示。
package apimetrics

import (
	"context"
	"net/http"
	"strings"
	"time"

	"github.com/antsurge/weaver-admin/pkg/metrics"
	"github.com/go-kratos/kratos/v2/middleware"
	"github.com/go-kratos/kratos/v2/transport"
	khttp "github.com/go-kratos/kratos/v2/transport/http"
)

// Options 中间件配置
type Options struct {
	collector        *metrics.Collector
	skipPathPrefixes []string
}

// Option 中间件配置项
type Option func(*Options)

// WithCollector 设置采集器（未设置时中间件直接放行）
func WithCollector(c *metrics.Collector) Option {
	return func(o *Options) {
		o.collector = c
	}
}

// WithSkipPathPrefix 跳过指定路径前缀（如长连接 SSE）
func WithSkipPathPrefix(prefixes ...string) Option {
	return func(o *Options) {
		o.skipPathPrefixes = append(o.skipPathPrefixes, prefixes...)
	}
}

// Server 构建接口指标采集中间件。
// 仅统计 HTTP 请求；OPTIONS 预检与配置跳过的前缀不统计。
func Server(opts ...Option) middleware.Middleware {
	o := &Options{}
	for _, opt := range opts {
		opt(o)
	}

	return func(handler middleware.Handler) middleware.Handler {
		return func(ctx context.Context, req any) (any, error) {
			if o.collector == nil {
				return handler(ctx, req)
			}

			tr, ok := transport.FromServerContext(ctx)
			if !ok {
				return handler(ctx, req)
			}
			ht, ok := tr.(*khttp.Transport)
			if !ok || ht.Request() == nil {
				return handler(ctx, req)
			}
			if ht.Request().Method == http.MethodOptions {
				return handler(ctx, req)
			}

			path := ht.PathTemplate()
			if path == "" {
				path = ht.Request().URL.Path
			}
			for _, p := range o.skipPathPrefixes {
				if strings.HasPrefix(path, p) {
					return handler(ctx, req)
				}
			}

			start := time.Now()
			resp, err := handler(ctx, req)
			o.collector.Record(strings.ToUpper(ht.Request().Method), path, time.Since(start), err != nil)
			return resp, err
		}
	}
}
