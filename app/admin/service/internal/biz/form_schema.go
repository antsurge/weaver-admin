package biz

import (
	"context"
	"time"

	"github.com/antsurge/weaver-admin/pkg/enthelper"
	"github.com/antsurge/weaver-admin/pkg/utils/uuid"
	"github.com/go-kratos/kratos/v2/errors"
	"github.com/go-kratos/kratos/v2/log"
)

// ErrFormSchemaNotFound 表单不存在
var ErrFormSchemaNotFound = errors.NotFound("FORM_SCHEMA_NOT_FOUND", "表单不存在或已禁用")

// FormSchema 表单定义（表单构建器产物）
type FormSchema struct {
	ID          string         `json:"id"`
	Name        string         `json:"name"`
	Code        string         `json:"code"`
	Description string         `json:"description"`
	SchemaJSON  map[string]any `json:"schemaJson,omitempty"`
	Status      string         `json:"status"`
	Remark      string         `json:"remark"`
	CreatedAt   time.Time      `json:"createdAt"`
	UpdatedAt   time.Time      `json:"updatedAt"`
	DeletedAt   *time.Time     `json:"deletedAt,omitempty"`
}

type ListFormSchemaRequest struct {
	enthelper.PaginationParams
	Name   string `form:"name" query:"name"`
	Code   string `form:"code" query:"code"`
	Status string `form:"status" query:"status"`
}

type ListFormSchemaOption struct {
	enthelper.QueryOption
}

type ListFormSchemaResponse struct {
	Data  []*FormSchema
	Total int
}

type FormSchemaRepo interface {
	ListFormSchema(context.Context, *ListFormSchemaRequest, ...*ListFormSchemaOption) (*ListFormSchemaResponse, error)
	GetFormSchemaByCode(context.Context, string) (*FormSchema, error)
	CreateFormSchema(context.Context, *FormSchema) error
	UpdateFormSchema(context.Context, *FormSchema) error
	DeleteFormSchema(context.Context, []string) error
	UpdateFormSchemaStatus(context.Context, string, string) error
}

type FormSchemaUsecase struct {
	repo FormSchemaRepo
	log  *log.Helper
}

func NewFormSchemaUsecase(repo FormSchemaRepo, logger log.Logger) *FormSchemaUsecase {
	return &FormSchemaUsecase{
		repo: repo,
		log:  log.NewHelper(logger),
	}
}

func (uc *FormSchemaUsecase) List(ctx context.Context, req *ListFormSchemaRequest) (*ListFormSchemaResponse, error) {
	return uc.repo.ListFormSchema(ctx, req)
}

func (uc *FormSchemaUsecase) CreateFormSchema(ctx context.Context, req *FormSchema) (*FormSchema, error) {
	form := req
	form.ID = uuid.GenerateXID()
	form.CreatedAt = time.Now()
	form.UpdatedAt = time.Now()
	if form.Status == "" {
		form.Status = "enabled"
	}

	err := uc.repo.CreateFormSchema(ctx, form)
	return form, err
}

func (uc *FormSchemaUsecase) UpdateFormSchema(ctx context.Context, req *FormSchema) (*FormSchema, error) {
	form := req
	form.UpdatedAt = time.Now()

	err := uc.repo.UpdateFormSchema(ctx, form)
	return form, err
}

func (uc *FormSchemaUsecase) DeleteFormSchema(ctx context.Context, ids []string) error {
	return uc.repo.DeleteFormSchema(ctx, ids)
}

func (uc *FormSchemaUsecase) UpdateFormSchemaStatus(ctx context.Context, id, status string) error {
	return uc.repo.UpdateFormSchemaStatus(ctx, id, status)
}

// GetFormSchemaByCode 按编码查询表单（供运行时渲染引用）
func (uc *FormSchemaUsecase) GetFormSchemaByCode(ctx context.Context, code string) (*FormSchema, error) {
	return uc.repo.GetFormSchemaByCode(ctx, code)
}
