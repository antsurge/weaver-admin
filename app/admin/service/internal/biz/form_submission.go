package biz

import (
	"context"
	"time"

	"github.com/antsurge/weaver-admin/pkg/enthelper"
	"github.com/antsurge/weaver-admin/pkg/utils/uuid"
	"github.com/go-kratos/kratos/v2/errors"
	"github.com/go-kratos/kratos/v2/log"
)

// FormSubmission 表单提交记录
type FormSubmission struct {
	ID        string         `json:"id"`
	FormCode  string         `json:"formCode"`
	FormName  string         `json:"formName,omitempty"`
	Data      map[string]any `json:"data,omitempty"`
	CreatedAt time.Time      `json:"createdAt"`
	UpdatedAt time.Time      `json:"updatedAt"`
	DeletedAt *time.Time     `json:"deletedAt,omitempty"`
}

type ListFormSubmissionRequest struct {
	enthelper.PaginationParams
	Code string `form:"code" query:"code"`
}

type ListFormSubmissionOption struct {
	enthelper.QueryOption
}

type ListFormSubmissionResponse struct {
	Data  []*FormSubmission
	Total int
}

// FormSchemaField schemaJson.fields 中的单个字段定义
type FormSchemaField struct {
	Key     string   `json:"key"`
	Label   string   `json:"label"`
	Type    string   `json:"type"`
	Options []string `json:"options,omitempty"`
}

// FormSchemaJSON schemaJson 顶层结构
type FormSchemaJSON struct {
	Version int               `json:"version"`
	Fields  []FormSchemaField `json:"fields"`
}

type FormSubmissionRepo interface {
	CreateFormSubmission(context.Context, *FormSubmission) error
	ListFormSubmission(context.Context, *ListFormSubmissionRequest, ...*ListFormSubmissionOption) (*ListFormSubmissionResponse, error)
}

type FormSubmissionUsecase struct {
	formSchemaRepo FormSchemaRepo
	repo           FormSubmissionRepo
	log            *log.Helper
}

func NewFormSubmissionUsecase(repo FormSubmissionRepo, formSchemaRepo FormSchemaRepo, logger log.Logger) *FormSubmissionUsecase {
	return &FormSubmissionUsecase{
		repo:           repo,
		formSchemaRepo: formSchemaRepo,
		log:            log.NewHelper(logger),
	}
}

// Submit 校验并保存表单提交数据
func (uc *FormSubmissionUsecase) Submit(ctx context.Context, code string, data map[string]any) (*FormSubmission, error) {
	if code == "" {
		return nil, errors.BadRequest("FORM_CODE_REQUIRED", "请填写表单编码")
	}

	form, err := uc.formSchemaRepo.GetFormSchemaByCode(ctx, code)
	if err != nil {
		return nil, err
	}
	if form == nil {
		return nil, ErrFormSchemaNotFound
	}

	// 校验必填字段
	if err := uc.validateRequired(form, data); err != nil {
		return nil, err
	}

	submission := &FormSubmission{
		ID:        uuid.GenerateXID(),
		FormCode:  form.Code,
		FormName:  form.Name,
		Data:      data,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}
	if err := uc.repo.CreateFormSubmission(ctx, submission); err != nil {
		return nil, err
	}
	return submission, nil
}

// validateRequired 根据 schemaJson 字段定义校验必填字段
// 当前设计器未提供"必填"标记，默认全部字段可选；此方法预留扩展点。
func (uc *FormSubmissionUsecase) validateRequired(form *FormSchema, data map[string]any) error {
	return nil
}

func (uc *FormSubmissionUsecase) List(ctx context.Context, req *ListFormSubmissionRequest) (*ListFormSubmissionResponse, error) {
	if req.Code == "" {
		return nil, errors.BadRequest("FORM_CODE_REQUIRED", "请填写表单编码")
	}
	return uc.repo.ListFormSubmission(ctx, req)
}
