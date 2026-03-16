package schema

import (
	"time"

	"entgo.io/ent"
	"entgo.io/ent/dialect/entsql"
	"entgo.io/ent/schema"
	"entgo.io/ent/schema/field"
)

type Organization struct {
	ent.Schema
}

func (Organization) Annotations() []schema.Annotation {
	return []schema.Annotation{
		entsql.Annotation{Table: "organization"},
		entsql.WithComments(true),
		schema.Comment("权限表"),
	}
}

func (Organization) Fields() []ent.Field {
	return []ent.Field{
		field.String("id").Unique().Immutable().MaxLen(36),

		field.String("parent_id").
			Default("").
			Comment("父部门ID，空表示根节点"),

		field.String("name").
			NotEmpty().
			MaxLen(64).
			Comment("部门名称"),

		field.String("code").
			NotEmpty().
			Unique().
			MaxLen(64).
			Comment("部门code"),

		field.Int("weight").
			Default(0).
			Comment("权重"),

		field.Enum("status").
			Values("enabled", "disabled").
			Default("enabled").
			Comment("状态：enabled=启用 disabled=禁用"),

		// 负责人姓名
		field.String("leader_name").
			Optional().
			MaxLen(64).
			Comment("负责人姓名"),

		// 联系电话
		field.String("leader_phone").
			Optional().
			MaxLen(20).
			Comment("联系电话"),

		// 邮箱
		field.String("leader_email").
			Optional().
			MaxLen(128).
			Comment("邮箱"),

		field.Time("created_at").
			Immutable().
			Default(time.Now),

		field.Time("updated_at").
			Default(time.Now).
			UpdateDefault(time.Now),
	}
}

func (Organization) Edges() []ent.Edge {
	return nil
	//return []ent.Edge{
	//	// 反向：哪些角色拥有此权限
	//	edge.From("roles", Role.Type).
	//		Ref("permissions").
	//		Through("role_permissions", RolePermission.Type),
	//}
}
