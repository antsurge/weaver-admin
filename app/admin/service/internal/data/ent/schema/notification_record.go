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

// NotificationRecord 用户收件箱：通知扇出到每个接收人后的一条记录。
// 这是「谁收到什么、是否已读」的真相源，实时推送只是加速通道。
type NotificationRecord struct {
	ent.Schema
}

func (NotificationRecord) Annotations() []schema.Annotation {
	return []schema.Annotation{
		entsql.Annotation{Table: "notification_record"},
		entsql.WithComments(true),
		schema.Comment("用户通知收件表"),
	}
}

// Mixin of the NotificationRecord.
func (NotificationRecord) Mixin() []ent.Mixin {
	return []ent.Mixin{TenantMixin{}}
}

func (NotificationRecord) Fields() []ent.Field {
	return []ent.Field{
		field.String("id").
			MaxLen(36).
			Unique().
			Immutable().
			Comment("记录ID"),

		field.String("notification_id").
			MaxLen(36).
			Comment("通知ID"),

		field.String("user_id").
			MaxLen(36).
			Comment("接收人ID"),

		field.Bool("is_read").
			Default(false).
			Comment("是否已读"),

		field.Time("read_at").
			Optional().
			Nillable().
			Comment("已读时间"),

		field.Bool("is_deleted").
			Default(false).
			Comment("是否已删除（回收站，用户侧软删）"),

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

func (NotificationRecord) Edges() []ent.Edge {
	return []ent.Edge{
		edge.From("notification", Notification.Type).
			Ref("records").
			Field("notification_id").
			Unique().
			Required(),
	}
}

func (NotificationRecord) Indexes() []ent.Index {
	return []ent.Index{
		// 防止同一用户对同一通知重复收件
		index.Fields("notification_id", "user_id").Unique(),
		// 未读数与收件箱列表的高频查询路径
		index.Fields("user_id", "is_read", "is_deleted"),
	}
}
