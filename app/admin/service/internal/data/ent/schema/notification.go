package schema

import (
	"time"

	"entgo.io/ent"
	"entgo.io/ent/dialect/entsql"
	"entgo.io/ent/schema"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"
)

// Notification 通知主表（一条通知的正文与投放范围）。
type Notification struct {
	ent.Schema
}

func (Notification) Annotations() []schema.Annotation {
	return []schema.Annotation{
		entsql.Annotation{Table: "notification"},
		entsql.WithComments(true),
		schema.Comment("通知表"),
	}
}

// Mixin of the Notification.
func (Notification) Mixin() []ent.Mixin {
	return []ent.Mixin{TenantMixin{}}
}

func (Notification) Fields() []ent.Field {
	return []ent.Field{
		field.String("id").
			MaxLen(36).
			Unique().
			Immutable().
			Comment("通知ID"),

		field.String("title").
			NotEmpty().
			MaxLen(128).
			Comment("通知标题"),

		field.String("content").
			Optional().
			MaxLen(1024).
			Comment("通知内容"),

		field.Enum("type").
			Values("system", "announce", "audit", "todo", "alert").
			Default("system").
			Comment("通知类型：system=系统 announce=公告 audit=审批 todo=待办 alert=告警"),

		field.Enum("level").
			Values("info", "success", "warn", "error").
			Default("info").
			Comment("重要级别：info=普通 success=成功 warn=警告 error=严重"),

		field.String("biz_type").
			Optional().
			MaxLen(64).
			Comment("业务类型，用于业务方归类（如 order、admin）"),

		field.String("biz_id").
			Optional().
			MaxLen(64).
			Comment("业务ID，配合 biz_type 定位数据来源"),

		field.String("sender_id").
			Optional().
			MaxLen(36).
			Comment("发送人ID，空表示系统发送"),

		field.String("sender_name").
			Optional().
			MaxLen(64).
			Comment("发送人名称（冗余，避免联表）"),

		field.String("link").
			Optional().
			MaxLen(256).
			Comment("点击后跳转的前端路由"),

		field.Enum("target_type").
			Values("all", "user", "role").
			Default("all").
			Comment("投放范围：all=全体用户 user=指定用户 role=指定角色"),

		field.JSON("target_ids", []string{}).
			Optional().
			Comment("投放目标ID集合，target_type 为 all 时为空"),

		field.Enum("status").
			Values("draft", "published", "revoked").
			Default("published").
			Comment("状态：draft=草稿 published=已发布 revoked=已撤回"),

		field.Time("published_at").
			Optional().
			Nillable().
			Comment("发布时间"),

		field.Time("expire_at").
			Optional().
			Nillable().
			Comment("过期时间，为空表示长期有效"),

		field.Time("created_at").
			Immutable().
			Default(time.Now).
			Comment("创建时间"),

		field.Time("updated_at").
			Default(time.Now).
			UpdateDefault(time.Now).
			Comment("更新时间"),

		field.Time("deleted_at").
			Optional().
			Nillable().
			Comment("删除时间"),
	}
}

func (Notification) Edges() []ent.Edge {
	return []ent.Edge{
		// 一条通知扇出成多条用户收件记录
		edge.To("records", NotificationRecord.Type),
	}
}

func (Notification) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("status", "created_at"),
		index.Fields("biz_type", "biz_id"),
	}
}
