package metadata

import (
	"context"
	"github.com/go-kratos/kratos/v2/transport"
)

const (
	AdminIDKey = "X-Admin-ID"
)

func SetAdminID(ctx context.Context, userID string) context.Context {
	if tr, ok := transport.FromServerContext(ctx); ok {
		tr.RequestHeader().Set(AdminIDKey, userID)
	}
	return ctx
}

func GetAdminID(ctx context.Context) string {
	adminID := ""
	if tr, ok := transport.FromServerContext(ctx); ok {
		adminID = tr.RequestHeader().Get(AdminIDKey)
	}
	return adminID
}
