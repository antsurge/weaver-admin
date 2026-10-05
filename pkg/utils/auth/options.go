package auth

import (
	"context"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

// TokenStore 用于校验 token 是否仍然有效。
// 配合 Redis 存储实现登出/强制下线后的 token 主动失效。
type TokenStore interface {
	Exists(ctx context.Context, token string) (bool, error)
}

type Options struct {
	Secret        []byte
	Expire        time.Duration
	Issuer        string
	SigningMethod jwt.SigningMethod
	// TokenStore 可选；设置后每次请求都会校验 token 是否仍存在于存储中
	TokenStore TokenStore
}

type Option func(options *Options)

// NewOptions 应用所有选项，返回配置对象
func NewOptions(opts ...Option) *Options {
	o := defaultOptions()
	for _, opt := range opts {
		opt(o)
	}
	return o
}

func defaultOptions() *Options {
	return &Options{
		Expire:        24 * time.Hour,
		SigningMethod: jwt.SigningMethodHS256,
	}
}

func WithSecret(secret []byte) Option {
	return func(o *Options) {
		o.Secret = secret
	}
}

func WithExpire(expire time.Duration) Option {
	return func(o *Options) {
		o.Expire = expire
	}
}

func WithIssuer(issuer string) Option {
	return func(o *Options) {
		o.Issuer = issuer
	}
}

func WithSigningMethod(method jwt.SigningMethod) Option {
	return func(o *Options) {
		o.SigningMethod = method
	}
}

// WithTokenStore 设置 token 存储校验（登出/强制下线后 token 立即失效）
func WithTokenStore(store TokenStore) Option {
	return func(o *Options) {
		o.TokenStore = store
	}
}
