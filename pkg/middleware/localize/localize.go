package localize

import (
	"context"
	"embed"
	"fmt"
	"github.com/BurntSushi/toml"
	"github.com/go-kratos/kratos/v2/errors"
	"github.com/go-kratos/kratos/v2/middleware"
	"github.com/go-kratos/kratos/v2/transport"
	"github.com/nicksnyder/go-i18n/v2/i18n"
	"golang.org/x/text/language"
	"io/fs"
	"log"
	"strings"
)

// If use go:embed
//
//go:embed lang/**/*.toml
var LocaleFS embed.FS

type localizerKey struct {
}

var (
	bundle = i18n.NewBundle(language.English)
)

func init() {
	bundle.RegisterUnmarshalFunc("toml", toml.Unmarshal)

	// 遍历 embed FS 里的文件
	fs.WalkDir(LocaleFS, "lang", func(path string, d fs.DirEntry, err error) error {
		if !d.IsDir() {
			bundle.LoadMessageFileFS(LocaleFS, path)
		}
		return nil
	})

	// 获取加载的语言
	tags := bundle.LanguageTags()
	log.Println("tags", tags)
}

func I18N() middleware.Middleware {
	return func(handler middleware.Handler) middleware.Handler {
		return func(ctx context.Context, req interface{}) (reply interface{}, err error) {
			var localizer *i18n.Localizer
			if tr, ok := transport.FromServerContext(ctx); ok {
				// HTTP
				if tr.Kind() == transport.KindHTTP {
					lang := "zh-cn"

					if tr, ok := transport.FromServerContext(ctx); ok {
						if accept := tr.RequestHeader().Get("Accept-Language"); accept != "" {
							lang = strings.ToLower(accept)
						}
					}
					localizer = i18n.NewLocalizer(bundle, lang)
				}

				// TODO：grpc
			}

			// 调用下游
			reply, err = handler(ctx, req)

			// 错误国际化处理
			if err != nil {
				if e := errors.FromError(err); e != nil {
					// message 翻译
					translated, eerr := localizer.Localize(&i18n.LocalizeConfig{
						MessageID: e.Message,
					})
					fmt.Println(e.Message)
					if eerr == nil {
						err = errors.BadRequest(e.Reason, translated)
					}
				}
			}

			return reply, err
		}
	}
}
