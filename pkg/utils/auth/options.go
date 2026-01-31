package auth

import (
	"github.com/golang-jwt/jwt/v5"
	"time"
)

type Options struct {
	Secret        []byte
	Expire        time.Duration
	Issuer        string
	SigningMethod jwt.SigningMethod
}

type Option func(options *Options)

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
