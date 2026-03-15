package schema

import (
	"time"

	"entgo.io/ent"
	"entgo.io/ent/dialect/entsql"
	"entgo.io/ent/schema"
	"entgo.io/ent/schema/field"
)

type Permission struct {
	ent.Schema
}

func (Permission) Annotations() []schema.Annotation {
	return []schema.Annotation{
		entsql.Annotation{Table: "permission"},
		entsql.WithComments(true),
		schema.Comment("权限表"),
	}
}

func (Permission) Fields() []ent.Field {
	return []ent.Field{
		field.String("id").Unique().Immutable().MaxLen(36),
		field.String("parent_id").Default("").Comment("父权限ID, null表示根节点"),
		field.String("name").NotEmpty().MaxLen(64).Comment("权限名称"),
		field.String("code").NotEmpty().Unique().MaxLen(64).Comment("权限唯一编码"),
		field.String("description").Optional().MaxLen(256).Comment("权限描述"),
		field.String("path").Default("").MaxLen(256).Comment("路由路径"),
		field.String("icon").Default("").MaxLen(128).Comment("图标"),
		field.Enum("type").Values("menu_dir", "menu", "button").Default("menu_dir").Comment("类型:menu_dir=菜单目录,menu=菜单项,button=页面按钮"),
		field.Enum("menu_type").Values("tab", "link", "iframe").Default("tab").Comment("菜单类型:tab=选项卡,link=链接(站外),iframe=Iframe"),
		field.String("url").Default("").MaxLen(256).Comment("URL"),
		field.String("component").Default("").MaxLen(256).Comment("组件路径"),
		field.Int("weight").Default(0).Comment("权重"),
		field.Enum("status").Values("enabled", "disabled").Default("enabled").Comment("状态：enabled=启用 disabled=禁用=禁用"),
		field.Time("created_at").Immutable().Default(time.Now),
		field.Time("updated_at").Default(time.Now).UpdateDefault(time.Now),
	}
}

func (Permission) Edges() []ent.Edge {
	return nil
	//return []ent.Edge{
	//	// 反向：哪些角色拥有此权限
	//	edge.From("roles", Role.Type).
	//		Ref("permissions").
	//		Through("role_permissions", RolePermission.Type),
	//}
}
