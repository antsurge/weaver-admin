package schema

import (
	"time"

	"entgo.io/ent"
	"entgo.io/ent/dialect/entsql"
	"entgo.io/ent/schema"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
)

// Admin holds the schema definition for the Admin entity.
type Admin struct {
	ent.Schema
}

// Annotations 指定表名
func (Admin) Annotations() []schema.Annotation {
	return []schema.Annotation{
		entsql.Annotation{Table: "admin"},
		entsql.WithComments(true),
		schema.Comment("管理员表"),
	}
}

// Mixin of the Admin.
func (Admin) Mixin() []ent.Mixin {
	return []ent.Mixin{TenantMixin{}}
}

// Fields of the Admin.
func (Admin) Fields() []ent.Field {
	return []ent.Field{
		field.String("id").Unique(),                          // 对应 Proto id
		field.String("real_name").Default(""),                // 对应 Proto name
		field.String("username"),                             // 对应 Proto username
		field.String("email").Optional(),                     // 对应 Proto email
		field.String("phone").Optional(),                     // 对应 Proto phone
		field.String("avatar").Optional(),                    // 对应 Proto avatar
		field.String("department_id").Optional().Default(""), // 对应 Proto departmentId（归属部门）
		field.String("password").Sensitive(),                 // 对应 Proto password (Sensitive 不会被 JSON 序列化)
		field.Enum("status").
			Values("enabled", "disabled").
			Default("enabled"). // 对应 Proto status：disabled 用户禁止登录
			Comment("状态：enabled=启用 disabled=禁用"),
		field.Time("created_at").
			Default(time.Now).Immutable(), // 对应 Proto createTime
		field.Time("updated_at").
			Default(time.Now).UpdateDefault(time.Now), // 对应 Proto updateTime
		field.Time("deleted_at").
			Optional().
			Nillable(), // 软删除时间
	}
}

// Edges of the Admin.
func (Admin) Edges() []ent.Edge {
	return []ent.Edge{
		// Admin → Roles (多对多，通过 admin_roles 关联表)
		edge.To("roles", Role.Type).
			Through("admin_roles", AdminRole.Type),
		// Admin → DataPermissions (多对多，通过 admin_data_permissions 关联表)
		edge.To("data_permissions", DataPermission.Type).
			Through("admin_data_permissions", AdminDataPermission.Type),
	}
}
