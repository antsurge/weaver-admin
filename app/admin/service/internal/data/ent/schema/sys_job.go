package schema

import (
	"time"

	"entgo.io/ent"
	"entgo.io/ent/dialect/entsql"
	"entgo.io/ent/schema"
	"entgo.io/ent/schema/field"
)

type SysJob struct {
	ent.Schema
}

func (SysJob) Annotations() []schema.Annotation {
	return []schema.Annotation{
		entsql.Annotation{Table: "sys_job"},
		entsql.WithComments(true),
		schema.Comment("定时任务表"),
	}
}

// Mixin of the SysJob.
func (SysJob) Mixin() []ent.Mixin {
	return []ent.Mixin{TenantMixin{}}
}

func (SysJob) Fields() []ent.Field {
	return []ent.Field{
		field.String("id").
			MaxLen(36).
			Unique().
			Immutable().
			Comment("任务ID"),

		field.String("name").
			NotEmpty().
			MaxLen(64).
			Comment("任务名称"),

		field.String("job_group").
			NotEmpty().
			Default("DEFAULT").
			MaxLen(64).
			Comment("任务分组"),

		field.String("job_type").
			Default("http").
			MaxLen(32).
			Comment("任务类型（http=HTTP调用）"),

		field.String("invoke_target").
			NotEmpty().
			MaxLen(512).
			Comment("调用目标（HTTP请求URL）"),

		field.String("cron_expression").
			NotEmpty().
			MaxLen(64).
			Comment("Cron 表达式"),

		field.String("misfire_policy").
			Default("immediately").
			MaxLen(32).
			Comment("错失执行策略（immediately=立即执行，once=执行一次，ignore=放弃执行）"),

		field.Bool("concurrent").
			Default(true).
			Comment("是否允许并发执行"),

		field.Enum("status").
			Values("enabled", "disabled").
			Default("enabled").
			Comment("状态"),

		field.String("remark").
			Optional().
			MaxLen(256).
			Comment("备注"),

		field.Time("next_run_time").
			Optional().
			Nillable().
			Comment("下次执行时间"),

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

func (SysJob) Edges() []ent.Edge {
	return nil
}
