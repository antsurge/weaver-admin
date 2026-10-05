package schema

import (
	"time"

	"entgo.io/ent"
	"entgo.io/ent/dialect/entsql"
	"entgo.io/ent/schema"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"
)

// LoginLog 登录日志：记录每次登录尝试（成功与失败），用于安全审计与异常排查。
type LoginLog struct {
	ent.Schema
}

func (LoginLog) Annotations() []schema.Annotation {
	return []schema.Annotation{
		entsql.Annotation{Table: "login_log"},
		entsql.WithComments(true),
		schema.Comment("登录日志表"),
	}
}

// Mixin of the LoginLog.
func (LoginLog) Mixin() []ent.Mixin {
	return []ent.Mixin{TenantMixin{}}
}

func (LoginLog) Fields() []ent.Field {
	return []ent.Field{
		field.String("id").
			MaxLen(36).
			Unique().
			Immutable().
			Comment("日志ID"),

		field.String("admin_id").
			Optional().
			MaxLen(36).
			Comment("用户ID（登录成功时记录）"),

		field.String("username").
			Optional().
			MaxLen(64).
			Comment("登录用户名（失败时也记录，便于排查撞库）"),

		field.String("ip").
			Optional().
			MaxLen(64).
			Comment("来源IP"),

		field.String("user_agent").
			Optional().
			MaxLen(512).
			Comment("客户端 UA"),

		field.Enum("status").
			Values("success", "fail").
			Default("success").
			Comment("结果：success=成功 fail=失败"),

		field.String("reason").
			Optional().
			MaxLen(128).
			Comment("失败原因（错误码或简要说明）"),

		field.Time("created_at").
			Immutable().
			Default(time.Now).
			Comment("登录时间"),
	}
}

func (LoginLog) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("username", "created_at"),
		index.Fields("status", "created_at"),
		index.Fields("created_at"),
	}
}
