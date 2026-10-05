package schema

import (
	"time"

	"entgo.io/ent"
	"entgo.io/ent/dialect/entsql"
	"entgo.io/ent/schema"
	"entgo.io/ent/schema/field"
)

// Tenant 租户表（平台级，不参与租户隔离）。
type Tenant struct {
	ent.Schema
}

func (Tenant) Annotations() []schema.Annotation {
	return []schema.Annotation{
		entsql.Annotation{Table: "tenant"},
		entsql.WithComments(true),
		schema.Comment("租户表"),
	}
}

func (Tenant) Fields() []ent.Field {
	return []ent.Field{
		field.String("id").
			MaxLen(36).
			Unique().
			Immutable().
			Comment("租户ID（默认租户固定为 default）"),

		field.String("code").
			NotEmpty().
			MaxLen(64).
			Unique().
			Immutable().
			Comment("租户编码（登录用，创建后不可修改）"),

		field.String("name").
			NotEmpty().
			MaxLen(64).
			Comment("租户名称"),

		field.Enum("status").
			Values("enabled", "disabled").
			Default("enabled").
			Comment("状态：enabled=启用 disabled=禁用"),

		field.Time("expire_at").
			Optional().
			Nillable().
			Comment("到期时间，空表示永久"),

		field.Int("max_users").
			Default(0).
			Comment("用户数上限，0 表示不限"),

		field.Int("max_roles").
			Default(0).
			Comment("角色数上限，0 表示不限"),

		field.String("contact_name").
			Optional().
			MaxLen(64).
			Comment("联系人"),

		field.String("contact_phone").
			Optional().
			MaxLen(20).
			Comment("联系电话"),

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

		field.Time("deleted_at").
			Optional().
			Nillable().
			Comment("删除时间"),
	}
}

func (Tenant) Edges() []ent.Edge {
	return nil
}
