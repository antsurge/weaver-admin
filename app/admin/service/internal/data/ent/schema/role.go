package schema

import (
	"time"

	"entgo.io/ent"
	"entgo.io/ent/dialect/entsql"
	"entgo.io/ent/schema"
	"entgo.io/ent/schema/field"
)

type Role struct {
	ent.Schema
}

// Annotations 指定表名
func (Role) Annotations() []schema.Annotation {
	return []schema.Annotation{
		entsql.Annotation{Table: "role"},
		entsql.WithComments(true),
		schema.Comment("系统角色表"),
	}
}

func (Role) Fields() []ent.Field {
	return []ent.Field{
		field.String("id").
			NotEmpty().
			Unique(),
		field.String("name").
			Comment("角色名称"),
		field.String("code").
			Unique().
			Comment("角色编码"),
		field.Int64("weight").
			Default(0).
			Comment("排序权重"),
		field.Enum("status").
			Values("enabled", "disabled").
			Default("enabled").
			Comment("状态"),
		field.String("remark").
			Optional().
			Comment("备注"),
		field.Bool("is_system").
			Default(false).
			Comment("是否系统内置"),

		field.String("data_scope").
			Default("all").
			Comment("数据权限范围: all / dept / self"),

		field.Time("created_at").
			Default(time.Now),
		field.Time("updated_at").
			Default(time.Now).
			UpdateDefault(time.Now),
		field.Time("deleted_at").
			Optional().
			Nillable(),
	}
}

func (Role) Edges() []ent.Edge {
	return nil
	//return []ent.Edge{
	//	// Role → Permissions
	//	edge.To("permissions", Permission.Type).
	//		Through("role_permissions", RolePermission.Type),
	//
	//	// Role ← Admins (反向)
	//	edge.From("admins", Admin.Type).
	//		Ref("roles"), // 对应 Admin.roles
	//}
}
