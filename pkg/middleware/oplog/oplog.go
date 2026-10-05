// Package oplog 提供操作日志自动采集中间件。
//
// 中间件本身不依赖任何存储：它把采集到的日志交给调用方注入的 Recorder，
// 由业务侧决定是同步写库、异步批量写还是投递到消息队列。
package oplog

import (
	"context"
	"encoding/json"
	"net"
	"net/http"
	"strings"
	"time"

	"github.com/antsurge/weaver-admin/pkg/metadata"
	"github.com/antsurge/weaver-admin/pkg/tenant"
	"github.com/go-kratos/kratos/v2/errors"
	"github.com/go-kratos/kratos/v2/middleware"
	"github.com/go-kratos/kratos/v2/transport"
	khttp "github.com/go-kratos/kratos/v2/transport/http"
)

// 日志结果
const (
	StatusSuccess = "success"
	StatusFail    = "fail"
)

// Entry 一条操作日志。
type Entry struct {
	TenantID  string
	AdminID   string
	AdminName string
	Module    string
	Operation string
	Method    string
	Path      string
	Summary   string
	IP        string
	UserAgent string
	Params    string
	Status    string
	Code      string
	Message   string
	CostMS    int64
}

// Recorder 日志落地抽象。实现必须保证非阻塞或快速返回，不能拖慢请求。
type Recorder interface {
	Record(*Entry)
}

// 需要脱敏的字段（小写比较）
var sensitiveKeys = map[string]struct{}{
	"password":        {},
	"oldpassword":     {},
	"newpassword":     {},
	"confirmpassword": {},
	"passwordhash":    {},
	"token":           {},
	"accesstoken":     {},
	"refreshtoken":    {},
	"secret":          {},
	"accesssecret":    {},
	"refreshsecret":   {},
	"accesskeysecret": {},
}

const maskValue = "******"

type options struct {
	skipOperations  map[string]struct{}
	skipPathPrefix  []string
	maxParamsLen    int
	summaryResolver func(method, path string) string
}

// Option 配置采集行为。
type Option func(*options)

// WithSkipOperations 跳过指定 operation（如验证码、SSE 长连接等高频/无意义请求）。
func WithSkipOperations(ops ...string) Option {
	return func(o *options) {
		for _, op := range ops {
			o.skipOperations[op] = struct{}{}
		}
	}
}

// WithSkipPathPrefix 跳过指定 HTTP 路径前缀。
func WithSkipPathPrefix(prefixes ...string) Option {
	return func(o *options) {
		o.skipPathPrefix = append(o.skipPathPrefix, prefixes...)
	}
}

// WithMaxParamsLen 设置请求参数的最大记录长度，超出截断。
func WithMaxParamsLen(n int) Option {
	return func(o *options) {
		if n > 0 {
			o.maxParamsLen = n
		}
	}
}

// WithSummaryResolver 注入「HTTP方法 + 路径 → 中文摘要」的解析器（如来自 OpenAPI）。
func WithSummaryResolver(f func(method, path string) string) Option {
	return func(o *options) {
		if f != nil {
			o.summaryResolver = f
		}
	}
}

func newOptions(opts ...Option) *options {
	o := &options{
		skipOperations: make(map[string]struct{}),
		maxParamsLen:   2000,
	}
	for _, opt := range opts {
		if opt != nil {
			opt(o)
		}
	}
	return o
}

// Server 返回操作日志采集中间件。
// 建议放在 recovery 之后、业务逻辑之前，这样能拿到真实耗时与错误。
func Server(rec Recorder, opts ...Option) middleware.Middleware {
	if rec == nil {
		return func(next middleware.Handler) middleware.Handler { return next }
	}
	o := newOptions(opts...)

	return func(handler middleware.Handler) middleware.Handler {
		return func(ctx context.Context, req any) (resp any, err error) {
			start := time.Now()
			resp, err = handler(ctx, req)
			o.emit(ctx, req, err, time.Since(start).Milliseconds(), rec)
			return resp, err
		}
	}
}

func (o *options) emit(ctx context.Context, req any, err error, costMS int64, rec Recorder) {
	tr, ok := transport.FromServerContext(ctx)
	if !ok {
		return
	}
	operation := tr.Operation()
	if operation == "" {
		return
	}
	if _, skip := o.skipOperations[operation]; skip {
		return
	}

	tid, _ := tenant.From(ctx)
	e := &Entry{
		TenantID:  tid,
		Operation: operation,
		Status:    StatusSuccess,
		CostMS:    costMS,
		AdminID:   metadata.GetAdminID(ctx),
	}

	// 从 operation 解析模块名：/admin.service.v1.Admin/CreateAdmin → Admin / CreateAdmin
	if idx := strings.LastIndex(operation, "/"); idx >= 0 {
		tail := operation[idx+1:]
		if parts := strings.SplitN(tail, "/", 2); len(parts) == 2 {
			e.Module, e.Operation = parts[0], parts[1]
		} else {
			e.Operation = tail
		}
		if slashIdx := strings.LastIndex(operation[:idx], "."); slashIdx >= 0 {
			e.Module = operation[slashIdx+1 : idx]
		}
	}

	// HTTP 细节（gRPC 场景拿不到，保持为空）
	if r, httpOK := khttp.RequestFromServerContext(ctx); httpOK && r != nil {
		e.Method = r.Method
		e.Path = r.URL.Path
		for _, prefix := range o.skipPathPrefix {
			if strings.HasPrefix(e.Path, prefix) {
				return
			}
		}
		e.IP = clientIP(r)
		e.UserAgent = r.Header.Get("User-Agent")
	}
	if e.IP == "" {
		e.IP = firstHeader(tr, "X-Forwarded-For", "X-Real-IP")
	}
	if e.UserAgent == "" {
		e.UserAgent = tr.RequestHeader().Get("User-Agent")
	}

	if o.summaryResolver != nil {
		e.Summary = o.summaryResolver(e.Method, e.Path)
	}

	if err != nil {
		e.Status = StatusFail
		if se := errors.FromError(err); se != nil {
			e.Code = se.Reason
			e.Message = se.Message
		} else {
			e.Code = "UNKNOWN"
			e.Message = err.Error()
		}
		if len(e.Message) > 512 {
			e.Message = e.Message[:512]
		}
	}

	e.Params = marshalParams(req, o.maxParamsLen)

	rec.Record(e)
}

func firstHeader(tr transport.Transporter, keys ...string) string {
	for _, k := range keys {
		if v := tr.RequestHeader().Get(k); v != "" {
			// X-Forwarded-For 可能是逗号分隔的链路
			if idx := strings.Index(v, ","); idx > 0 {
				return strings.TrimSpace(v[:idx])
			}
			return strings.TrimSpace(v)
		}
	}
	return ""
}

// clientIP 从 HTTP 请求中解析客户端 IP。
func clientIP(r *http.Request) string {
	if v := r.Header.Get("X-Forwarded-For"); v != "" {
		if idx := strings.Index(v, ","); idx > 0 {
			return strings.TrimSpace(v[:idx])
		}
		return strings.TrimSpace(v)
	}
	if v := r.Header.Get("X-Real-IP"); v != "" {
		return strings.TrimSpace(v)
	}
	if host, _, err := net.SplitHostPort(r.RemoteAddr); err == nil {
		return host
	}
	return r.RemoteAddr
}

// marshalParams 序列化请求参数并脱敏敏感字段。
func marshalParams(req any, maxLen int) string {
	if req == nil {
		return ""
	}
	raw, err := json.Marshal(req)
	if err != nil {
		return ""
	}
	var decoded any
	if err := json.Unmarshal(raw, &decoded); err != nil {
		// 非对象（数组/标量）无法脱敏，直接截断
		return truncate(string(raw), maxLen)
	}
	maskSensitive(decoded)
	masked, err := json.Marshal(decoded)
	if err != nil {
		return truncate(string(raw), maxLen)
	}
	return truncate(string(masked), maxLen)
}

func maskSensitive(v any) {
	switch node := v.(type) {
	case map[string]any:
		for k, val := range node {
			if _, sensitive := sensitiveKeys[strings.ToLower(k)]; sensitive {
				node[k] = maskValue
				continue
			}
			maskSensitive(val)
		}
	case []any:
		for _, item := range node {
			maskSensitive(item)
		}
	}
}

func truncate(s string, maxLen int) string {
	if maxLen <= 0 || len(s) <= maxLen {
		return s
	}
	return s[:maxLen] + "...(truncated)"
}
