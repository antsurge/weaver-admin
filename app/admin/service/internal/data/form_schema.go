package data

import (
	"context"
	"time"

	"github.com/antsurge/weaver-admin/app/admin/service/internal/biz"
	"github.com/antsurge/weaver-admin/app/admin/service/internal/data/ent"
	"github.com/antsurge/weaver-admin/app/admin/service/internal/data/ent/formschema"
	"github.com/antsurge/weaver-admin/pkg/enthelper"
	"github.com/go-kratos/kratos/v2/log"
)

var _ biz.FormSchemaRepo = (*formSchemaRepo)(nil)

type formSchemaRepo struct {
	data *Data
	log  *log.Helper
}

func NewFormSchemaRepo(data *Data, logger log.Logger) biz.FormSchemaRepo {
	return &formSchemaRepo{
		data: data,
		log:  log.NewHelper(logger),
	}
}

func (r *formSchemaRepo) ListFormSchema(ctx context.Context, params *biz.ListFormSchemaRequest, opts ...*biz.ListFormSchemaOption) (*biz.ListFormSchemaResponse, error) {
	query := r.data.db.FormSchema.Query().
		Order(ent.Desc(formschema.FieldCreatedAt))

	opt := &biz.ListFormSchemaOption{}
	if len(opts) > 0 {
		opt = opts[0]
	}

	if opt.OnlyDeleted {
		query = query.Where(formschema.DeletedAtNotNil())
	} else if !opt.IncludeDeleted {
		query = query.Where(formschema.DeletedAtIsNil())
	}

	if v := params.Name; len(v) > 0 {
		query = query.Where(formschema.NameContains(v))
	}
	if v := params.Code; len(v) > 0 {
		query = query.Where(formschema.CodeContains(v))
	}
	if v := params.Status; len(v) > 0 {
		query = query.Where(formschema.StatusEQ(formschema.Status(v)))
	}

	res, err := enthelper.Pagination[*ent.FormSchema, *ent.FormSchemaQuery](ctx, query, params.PaginationParams)
	if err != nil {
		return nil, err
	}

	data := make([]*biz.FormSchema, 0, res.Total)
	for _, v := range res.Data {
		data = append(data, &biz.FormSchema{
			ID:          v.ID,
			Name:        v.Name,
			Code:        v.Code,
			Description: v.Description,
			SchemaJSON:  v.SchemaJSON,
			Status:      string(v.Status),
			Remark:      v.Remark,
			CreatedAt:   v.CreatedAt,
			UpdatedAt:   v.UpdatedAt,
			DeletedAt:   v.DeletedAt,
		})
	}

	return &biz.ListFormSchemaResponse{
		Data:  data,
		Total: res.Total,
	}, nil
}

func (r *formSchemaRepo) GetFormSchemaByCode(ctx context.Context, code string) (*biz.FormSchema, error) {
	v, err := r.data.db.FormSchema.Query().
		Where(
			formschema.CodeEQ(code),
			formschema.DeletedAtIsNil(),
			formschema.StatusEQ(formschema.StatusEnabled),
		).
		Only(ctx)
	if err != nil {
		if ent.IsNotFound(err) {
			return nil, nil
		}
		return nil, err
	}

	return &biz.FormSchema{
		ID:          v.ID,
		Name:        v.Name,
		Code:        v.Code,
		Description: v.Description,
		SchemaJSON:  v.SchemaJSON,
		Status:      string(v.Status),
		Remark:      v.Remark,
		CreatedAt:   v.CreatedAt,
		UpdatedAt:   v.UpdatedAt,
		DeletedAt:   v.DeletedAt,
	}, nil
}

func (r *formSchemaRepo) CreateFormSchema(ctx context.Context, d *biz.FormSchema) error {
	_, err := r.data.db.FormSchema.Create().
		SetID(d.ID).
		SetName(d.Name).
		SetCode(d.Code).
		SetNillableDescription(&d.Description).
		SetSchemaJSON(d.SchemaJSON).
		SetStatus(formschema.Status(d.Status)).
		SetNillableRemark(&d.Remark).
		SetCreatedAt(d.CreatedAt).
		SetUpdatedAt(d.UpdatedAt).
		Save(ctx)
	return err
}

func (r *formSchemaRepo) UpdateFormSchema(ctx context.Context, d *biz.FormSchema) error {
	_, err := r.data.db.FormSchema.UpdateOneID(d.ID).
		SetName(d.Name).
		SetCode(d.Code).
		SetNillableDescription(&d.Description).
		SetSchemaJSON(d.SchemaJSON).
		SetStatus(formschema.Status(d.Status)).
		SetNillableRemark(&d.Remark).
		SetUpdatedAt(d.UpdatedAt).
		Save(ctx)
	return err
}

func (r *formSchemaRepo) DeleteFormSchema(ctx context.Context, ids []string) error {
	if len(ids) == 0 {
		return nil
	}

	now := time.Now()
	return r.data.db.FormSchema.
		Update().
		Where(formschema.IDIn(ids...)).
		SetDeletedAt(now).
		Exec(ctx)
}

func (r *formSchemaRepo) UpdateFormSchemaStatus(ctx context.Context, id, status string) error {
	_, err := r.data.db.FormSchema.UpdateOneID(id).
		SetStatus(formschema.Status(status)).
		SetUpdatedAt(time.Now()).
		Save(ctx)
	return err
}
