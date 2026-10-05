package schema

import (
	"time"

	"entgo.io/ent"
	"entgo.io/ent/dialect/entsql"
	"entgo.io/ent/schema"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"
)

// OperationLog 操作日志：记录「谁在什么时间对哪个接口做了什么操作」。
// 由 HTTP/gRPC 中间件自动采集，用于审计追溯。
type OperationLog struct {
	ent.Schema
}

func (OperationLog) Annotations() []schema.Annotation {
	return []schema.Annotation{
		entsql.Annotation{Table: "operation_log"},
		entsql.WithComments(true),
		schema.Comment("操作日志表"),
	}
}

// Mixin of the OperationLog.
func (OperationLog) Mixin() []ent.Mixin {
	return []ent.Mixin{TenantMixin{}}
}

func (OperationLog) Fields() []ent.Field {
	return []ent.Field{
		field.String("id").
			MaxLen(36).
			Unique().
			Immutable().
			Comment("日志ID"),

		field.String("admin_id").
			Optional().
			MaxLen(36).
			Comment("操作人ID（未登录操作为空）"),

		field.String("admin_name").
			Optional().
			MaxLen(64).
			Comment("操作人姓名（冗余，避免联表）"),

		field.String("module").
			Optional().
			MaxLen(64).
			Comment("所属模块/服务名，如 Admin、Dictionary"),

		field.String("operation").
			Optional().
			MaxLen(128).
			Comment("操作名，如 CreateAdmin"),

		field.String("method").
			Optional().
			MaxLen(16).
			Comment("HTTP 方法"),

		field.String("path").
			Optional().
			MaxLen(256).
			Comment("请求路径"),

		field.String("summary").
			Optional().
			MaxLen(256).
			Comment("接口中文摘要（来自 OpenAPI）"),

		field.String("ip").
			Optional().
			MaxLen(64).
			Comment("来源IP"),

		field.String("user_agent").
			Optional().
			MaxLen(512).
			Comment("客户端 UA"),

		field.Text("params").
			Optional().
			Comment("请求参数（JSON，已脱敏并截断）"),

		field.Enum("status").
			Values("success", "fail").
			Default("success").
			Comment("结果：success=成功 fail=失败"),

		field.String("code").
			Optional().
			MaxLen(64).
			Comment("失败时的错误码"),

		field.String("message").
			Optional().
			MaxLen(512).
			Comment("失败时的错误信息"),

		field.Int("cost_ms").
			Default(0).
			Comment("耗时（毫秒）"),

		field.Time("created_at").
			Immutable().
			Default(time.Now).
			Comment("操作时间"),
	}
}

func (OperationLog) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("admin_id", "created_at"),
		index.Fields("status", "created_at"),
		index.Fields("created_at"),
	}
}
