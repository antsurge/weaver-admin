package bufvalidate

import (
	"buf.build/go/protovalidate"
	"context"
	"github.com/go-kratos/kratos/v2/errors"
	"github.com/go-kratos/kratos/v2/middleware"
	"google.golang.org/protobuf/proto"
)

var validatorInstance protovalidate.Validator

func init() {
	v, err := protovalidate.New()
	if err != nil {
		panic(err)
	}
	validatorInstance = v
}

// buf protovalidate
func BufValidator() middleware.Middleware {
	return func(handler middleware.Handler) middleware.Handler {
		return func(ctx context.Context, req interface{}) (interface{}, error) {
			msg, ok := req.(proto.Message)
			if !ok {
				return handler(ctx, req)
			}

			if err := validatorInstance.Validate(msg); err != nil {
				// 如果是 ValidationError 类型，可以获取字段信息
				if verr, ok := err.(*protovalidate.ValidationError); ok {
					var errMsg string
					// 只返回第一个字段错误
					if len(verr.Violations) > 0 {
						v := verr.Violations[0]
						errMsg = v.Proto.GetMessage()
					} else {
						errMsg = verr.Error()
					}

					return nil, errors.BadRequest("VALIDATOR", errMsg).WithCause(err)
				}
				// 不是 ValidationError，直接返回原始错误
				return nil, errors.BadRequest("VALIDATOR", err.Error()).WithCause(err)
			}
			return handler(ctx, req)
		}
	}
}
