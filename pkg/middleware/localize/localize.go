package localize

import (
	"context"
	"embed"
	"github.com/BurntSushi/toml"
	"github.com/go-kratos/kratos/v2/errors"
	"github.com/go-kratos/kratos/v2/middleware"
	"github.com/go-kratos/kratos/v2/transport"
	"github.com/nicksnyder/go-i18n/v2/i18n"
	"golang.org/x/text/language"
	"io/fs"
	"strings"
)

//go:embed lang/**/*.toml
var LocaleFS embed.FS

var bundle = i18n.NewBundle(language.English)

func init() {
	bundle.RegisterUnmarshalFunc("toml", toml.Unmarshal)

	// 遍历 embed FS 里的文件
	_ = fs.WalkDir(LocaleFS, "lang", func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if !d.IsDir() {
			bundle.LoadMessageFileFS(LocaleFS, path)
		}
		return nil
	})
}

// 默认语言
const defaultLang = "zh-cn"

// localizerKey 用于把语言本地化器注入上下文，供链路内其他中间件（如 bufvalidate）复用
type localizerKey struct{}

// normalizeLang 将客户端语言标识规范化为翻译文件目录名（zh-cn / en-us）
func normalizeLang(lang string) string {
	lang = strings.ToLower(strings.ReplaceAll(lang, "_", "-"))
	lang = strings.TrimSpace(strings.Split(lang, ",")[0])
	if i := strings.Index(lang, ";"); i > 0 {
		lang = lang[:i]
	}
	switch {
	case strings.HasPrefix(lang, "zh"):
		return "zh-cn"
	case strings.HasPrefix(lang, "en"):
		return "en-us"
	default:
		return lang
	}
}

// LangFromContext 从上下文获取已解析的语言标识（缺失时返回默认语言）
func LangFromContext(ctx context.Context) string {
	if l, ok := ctx.Value(localizerKey{}).(string); ok && l != "" {
		return l
	}
	return defaultLang
}

// resolveLang 从传输层解析客户端语言
// HTTP：Accept-Language 请求头；gRPC：metadata 中的 accept-language
func resolveLang(ctx context.Context) string {
	if tr, ok := transport.FromServerContext(ctx); ok {
		header := tr.RequestHeader()
		if header != nil {
			for _, key := range []string{"Accept-Language", "accept-language"} {
				if v := header.Get(key); v != "" {
					return normalizeLang(v)
				}
			}
		}
	}
	return defaultLang
}

// Translate 根据请求语言翻译指定的消息 key；
// 命中返回翻译文本，未命中返回 ok=false（由调用方决定回退策略）。
func Translate(ctx context.Context, key string) (string, bool) {
	lang := LangFromContext(ctx)
	localizer := i18n.NewLocalizer(bundle, lang)
	msg, err := localizer.Localize(&i18n.LocalizeConfig{MessageID: key})
	if err != nil {
		return "", false
	}
	return msg, true
}

// I18N 返回国际化中间件：
//  1. 解析请求语言并注入上下文；
//  2. 下游返回 Kratos 错误时，以错误码（Reason）为 key 查翻译文件，
//     命中则把错误 Message 替换为对应语言的文案。
func I18N() middleware.Middleware {
	return func(handler middleware.Handler) middleware.Handler {
		return func(ctx context.Context, req interface{}) (reply interface{}, err error) {
			lang := resolveLang(ctx)
			ctx = context.WithValue(ctx, localizerKey{}, lang)

			reply, err = handler(ctx, req)

			if err == nil {
				return reply, nil
			}

			// 错误国际化：以 Reason（错误码）为翻译 key
			if e := errors.FromError(err); e != nil {
				if translated, ok := Translate(ctx, e.Reason); ok && translated != "" {
					// 直接修改原 error 的 Message，保留 cause/metadata 等信息
					if ke, ok := err.(*errors.Error); ok {
						ke.Message = translated
					} else {
						err = errors.New(int(e.Code), e.Reason, translated)
					}
				}
			}
			return reply, err
		}
	}
}
