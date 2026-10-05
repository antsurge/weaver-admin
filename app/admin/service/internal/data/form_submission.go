package data

import (
	"context"
	"time"

	"github.com/antsurge/weaver-admin/app/admin/service/internal/biz"
	"github.com/antsurge/weaver-admin/app/admin/service/internal/data/ent"
	"github.com/antsurge/weaver-admin/app/admin/service/internal/data/ent/formsubmission"
	"github.com/antsurge/weaver-admin/pkg/enthelper"
	"github.com/go-kratos/kratos/v2/log"
)

var _ biz.FormSubmissionRepo = (*formSubmissionRepo)(nil)

type formSubmissionRepo struct {
	data *Data
	log  *log.Helper
}

func NewFormSubmissionRepo(data *Data, logger log.Logger) biz.FormSubmissionRepo {
	return &formSubmissionRepo{
		data: data,
		log:  log.NewHelper(logger),
	}
}

func (r *formSubmissionRepo) CreateFormSubmission(ctx context.Context, d *biz.FormSubmission) error {
	_, err := r.data.db.FormSubmission.Create().
		SetID(d.ID).
		SetFormCode(d.FormCode).
		SetNillableFormName(&d.FormName).
		SetSubmitData(d.Data).
		SetCreatedAt(d.CreatedAt).
		SetUpdatedAt(d.UpdatedAt).
		Save(ctx)
	return err
}

func (r *formSubmissionRepo) ListFormSubmission(ctx context.Context, params *biz.ListFormSubmissionRequest, opts ...*biz.ListFormSubmissionOption) (*biz.ListFormSubmissionResponse, error) {
	query := r.data.db.FormSubmission.Query().
		Order(ent.Desc(formsubmission.FieldCreatedAt))

	opt := &biz.ListFormSubmissionOption{}
	if len(opts) > 0 {
		opt = opts[0]
	}

	if opt.OnlyDeleted {
		query = query.Where(formsubmission.DeletedAtNotNil())
	} else if !opt.IncludeDeleted {
		query = query.Where(formsubmission.DeletedAtIsNil())
	}

	if v := params.Code; len(v) > 0 {
		query = query.Where(formsubmission.FormCodeEQ(v))
	}

	res, err := enthelper.Pagination[*ent.FormSubmission, *ent.FormSubmissionQuery](ctx, query, params.PaginationParams)
	if err != nil {
		return nil, err
	}

	data := make([]*biz.FormSubmission, 0, res.Total)
	for _, v := range res.Data {
		data = append(data, &biz.FormSubmission{
			ID:        v.ID,
			FormCode:  v.FormCode,
			FormName:  v.FormName,
			Data:      v.SubmitData,
			CreatedAt: v.CreatedAt,
			UpdatedAt: v.UpdatedAt,
			DeletedAt: v.DeletedAt,
		})
	}

	return &biz.ListFormSubmissionResponse{
		Data:  data,
		Total: res.Total,
	}, nil
}

// DeleteFormSubmission 软删除提交记录（预留：管理员清理数据）
func (r *formSubmissionRepo) DeleteFormSubmission(ctx context.Context, ids []string) error {
	if len(ids) == 0 {
		return nil
	}
	now := time.Now()
	return r.data.db.FormSubmission.
		Update().
		Where(formsubmission.IDIn(ids...)).
		SetDeletedAt(now).
		Exec(ctx)
}
