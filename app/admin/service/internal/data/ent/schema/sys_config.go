package schema

import (
	"time"

	"entgo.io/ent"
	"entgo.io/ent/dialect/entsql"
	"entgo.io/ent/schema"
	"entgo.io/ent/schema/field"
)

type SysConfig struct {
	ent.Schema
}

func (SysConfig) Annotations() []schema.Annotation {
	return []schema.Annotation{
		entsql.Annotation{Table: "sys_config"},
		entsql.WithComments(true),
		schema.Comment("系统参数配置表"),
	}
}

// Mixin of the SysConfig.
func (SysConfig) Mixin() []ent.Mixin {
	return []ent.Mixin{TenantMixin{}}
}

func (SysConfig) Fields() []ent.Field {
	return []ent.Field{
		field.String("id").
			MaxLen(36).
			Unique().
			Immutable().
			Comment("参数ID"),

		field.String("name").
			NotEmpty().
			MaxLen(64).
			Comment("参数名称"),

		field.String("key").
			NotEmpty().
			MaxLen(100).
			Unique().
			Comment("参数键名"),

		field.String("value").
			NotEmpty().
			MaxLen(512).
			Comment("参数键值"),

		field.Enum("config_type").
			Values("Y", "N").
			Default("N").
			Comment("系统内置（Y=内置，N=外置）"),

		field.Enum("status").
			Values("enabled", "disabled").
			Default("enabled").
			Comment("状态"),

		field.String("remark").
			Optional().
			MaxLen(256).
			Comment("备注"),

		field.String("group").
			Optional().
			MaxLen(50).
			Comment("分组名称（可为空，即单一参数）"),

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

func (SysConfig) Edges() []ent.Edge {
	return nil
}
