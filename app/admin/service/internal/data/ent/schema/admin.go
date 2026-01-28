package schema

import (
	"entgo.io/ent"
	"entgo.io/ent/dialect/entsql"
	"entgo.io/ent/schema"
	"entgo.io/ent/schema/field"
	"time"
)

// Admin holds the schema definition for the Admin entity.
type Admin struct {
	ent.Schema
}

// Annotations 指定表名
func (Admin) Annotations() []schema.Annotation {
	return []schema.Annotation{
		entsql.Annotation{Table: "admin"}, // 3. 显式指定表名为 admin
	}
}

// Fields of the Admin.
func (Admin) Fields() []ent.Field {
	return []ent.Field{
		field.String("id").Unique(),          // 对应 Proto id
		field.String("name").Default(""),     // 对应 Proto name
		field.String("username").Unique(),    // 对应 Proto username
		field.String("email").Optional(),     // 对应 Proto email
		field.String("phone").Optional(),     // 对应 Proto phone
		field.String("avatar").Optional(),    // 对应 Proto avatar
		field.String("password").Sensitive(), // 对应 Proto password (Sensitive 不会被 JSON 序列化)
		field.Time("create_time").
			Default(time.Now).Immutable(), // 对应 Proto createTime
		field.Time("update_time").
			Default(time.Now).UpdateDefault(time.Now), // 对应 Proto updateTime
	}
}

// Edges of the Admin.
func (Admin) Edges() []ent.Edge {
	return nil // 稍后可以增加 Role 的关联
}
