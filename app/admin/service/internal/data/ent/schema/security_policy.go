package schema

import (
	"time"

	"entgo.io/ent"
	"entgo.io/ent/dialect/entsql"
	"entgo.io/ent/schema"
	"entgo.io/ent/schema/field"
)

// SecurityPolicy 安全策略（登录策略 + 访问控制），全库单例配置。
type SecurityPolicy struct {
	ent.Schema
}

func (SecurityPolicy) Annotations() []schema.Annotation {
	return []schema.Annotation{
		entsql.Annotation{Table: "security_policy"},
		entsql.WithComments(true),
		schema.Comment("安全策略配置表（单例）"),
	}
}

func (SecurityPolicy) Fields() []ent.Field {
	return []ent.Field{
		field.String("id").
			MaxLen(36).
			Unique().
			Immutable().
			Comment("策略ID（固定为 default）"),

		field.Bool("login_fail_enabled").
			Default(false).
			Comment("是否启用登录失败次数限制"),

		field.Int("login_fail_max").
			Default(5).
			Comment("窗口内允许的最大失败次数"),

		field.Int("login_fail_window_minutes").
			Default(15).
			Comment("失败次数统计窗口（分钟）"),

		field.Int("login_lock_minutes").
			Default(15).
			Comment("达到上限后的锁定时长（分钟）"),

		field.Enum("ip_mode").
			Values("off", "whitelist", "blacklist").
			Default("off").
			Comment("IP 访问控制模式：off=关闭 whitelist=白名单 blacklist=黑名单"),

		field.Text("ip_list").
			Optional().
			Comment("IP 列表（换行或逗号分隔，支持 CIDR）"),

		field.Int64("access_token_ttl_seconds").
			Default(7200).
			Comment("access token 有效期（秒），合法范围 60-2592000，0 回退默认值 7200"),

		field.Int64("refresh_token_ttl_seconds").
			Default(604800).
			Comment("refresh token 有效期（秒），合法范围 300-31536000，0 回退默认值 604800"),

		field.String("remark").
			Optional().
			MaxLen(256).
			Comment("备注"),

		field.Time("created_at").
			Immutable().
			Default(time.Now).
			Comment("创建时间"),

		field.Time("updated_at").
			Default(time.Now).
			UpdateDefault(time.Now).
			Comment("更新时间"),
	}
}

func (SecurityPolicy) Edges() []ent.Edge {
	return nil
}
