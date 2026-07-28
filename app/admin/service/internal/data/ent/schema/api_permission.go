package schema

import (
	"entgo.io/ent"
	"entgo.io/ent/dialect/entsql"
	"entgo.io/ent/schema"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
)

// ApiPermission 菜单按钮可绑定的接口权限
type ApiPermission struct {
	ent.Schema
}

func (ApiPermission) Annotations() []schema.Annotation {
	return []schema.Annotation{
		entsql.Annotation{Table: "api_permission"},
		entsql.WithComments(true),
		schema.Comment("接口权限表（菜单按钮绑定）"),
	}
}

func (ApiPermission) Fields() []ent.Field {
	return []ent.Field{
		field.String("id").Unique().Immutable().MaxLen(36),
		field.String("service").NotEmpty().MaxLen(128).Comment("服务全限定名，如 admin.service.v1.PermissionService"),
		field.String("method").NotEmpty().MaxLen(10).Comment("HTTP method: GET/POST/PUT/DELETE"),
		field.String("path").NotEmpty().MaxLen(256).Comment("接口路径，如 /admin/v1/menu"),
		field.String("summary").Default("").MaxLen(256).Comment("接口描述"),
		// 业务唯一键：service + " " + method + " " + path（去重+回查用）
		field.String("code").Unique().NotEmpty().MaxLen(512).Comment("业务唯一键 service|method|path"),
	}
}

func (ApiPermission) Edges() []ent.Edge {
	return []ent.Edge{
		// 反向关联菜单（many-to-many）
		edge.From("menus", Menu.Type).
			Ref("api_permissions"),
	}
}