package service

import (
	"context"

	adminV1 "github.com/antsurge/weaver-admin/api/gen/go/admin/service/v1"
	lowcodeV1 "github.com/antsurge/weaver-admin/api/gen/go/lowcode/service/v1"
	"github.com/antsurge/weaver-admin/app/admin/service/internal/biz"
	"github.com/antsurge/weaver-admin/pkg/utils/copierx"
	"github.com/jinzhu/copier"
	"google.golang.org/protobuf/types/known/emptypb"
	"google.golang.org/protobuf/types/known/structpb"
	"google.golang.org/protobuf/types/known/timestamppb"
)

type LowcodeService struct {
	adminV1.UnimplementedLowcodeServer

	formSchemaUc     *biz.FormSchemaUsecase
	formSubmissionUc *biz.FormSubmissionUsecase
	codegenUc        *biz.CodegenUsecase
}

func NewLowcodeService(
	formSchemaUc *biz.FormSchemaUsecase,
	formSubmissionUc *biz.FormSubmissionUsecase,
	codegenUc *biz.CodegenUsecase,
) *LowcodeService {
	return &LowcodeService{
		formSchemaUc:     formSchemaUc,
		formSubmissionUc: formSubmissionUc,
		codegenUc:        codegenUc,
	}
}

// ====== 表单构建器 ======

func (s *LowcodeService) ListFormSchema(ctx context.Context, req *lowcodeV1.ListFormSchemaRequest) (*lowcodeV1.ListFormSchemaResponse, error) {
	input := biz.ListFormSchemaRequest{}
	var err error
	err = copier.Copy(&input, req)
	if err != nil {
		return nil, err
	}

	res, err := s.formSchemaUc.List(ctx, &input)
	if err != nil {
		return nil, err
	}

	output := &lowcodeV1.ListFormSchemaResponse{}
	if res != nil {
		output.Total = int64(res.Total)
		err = copierx.Copy(&output.Items, &res.Data)
	}

	return output, err
}

func (s *LowcodeService) GetFormSchemaByCode(ctx context.Context, req *lowcodeV1.GetFormSchemaByCodeRequest) (*lowcodeV1.FormSchema, error) {
	form, err := s.formSchemaUc.GetFormSchemaByCode(ctx, req.Code)
	if err != nil {
		return nil, err
	}
	if form == nil {
		return nil, nil
	}

	output := &lowcodeV1.FormSchema{}
	err = copierx.Copy(output, form)
	return output, err
}

func (s *LowcodeService) CreateFormSchema(ctx context.Context, req *lowcodeV1.CreateFormSchemaRequest) (*lowcodeV1.FormSchema, error) {
	input := biz.FormSchema{}
	var err error
	err = copier.Copy(&input, req)
	if err != nil {
		return nil, err
	}

	form, err := s.formSchemaUc.CreateFormSchema(ctx, &input)
	if err != nil {
		return nil, err
	}

	output := &lowcodeV1.FormSchema{}
	err = copierx.Copy(output, form)
	return output, err
}

func (s *LowcodeService) UpdateFormSchema(ctx context.Context, req *lowcodeV1.UpdateFormSchemaRequest) (*lowcodeV1.FormSchema, error) {
	input := biz.FormSchema{}
	var err error
	err = copier.Copy(&input, req)
	if err != nil {
		return nil, err
	}

	form, err := s.formSchemaUc.UpdateFormSchema(ctx, &input)
	if err != nil {
		return nil, err
	}

	output := &lowcodeV1.FormSchema{}
	err = copierx.Copy(output, form)
	return output, err
}

func (s *LowcodeService) UpdateFormSchemaStatus(ctx context.Context, req *lowcodeV1.UpdateFormSchemaStatusRequest) (*emptypb.Empty, error) {
	err := s.formSchemaUc.UpdateFormSchemaStatus(ctx, req.Id, req.Status)
	return nil, err
}

func (s *LowcodeService) DeleteFormSchema(ctx context.Context, req *lowcodeV1.DeleteFormSchemaRequest) (*emptypb.Empty, error) {
	err := s.formSchemaUc.DeleteFormSchema(ctx, req.Ids)
	return nil, err
}

func (s *LowcodeService) SubmitFormData(ctx context.Context, req *lowcodeV1.SubmitFormDataRequest) (*lowcodeV1.SubmitFormDataResponse, error) {
	var data map[string]any
	if req.Data != nil {
		data = req.Data.AsMap()
	}

	submission, err := s.formSubmissionUc.Submit(ctx, req.Code, data)
	if err != nil {
		return nil, err
	}

	return &lowcodeV1.SubmitFormDataResponse{Id: submission.ID}, nil
}

func (s *LowcodeService) ListFormSubmission(ctx context.Context, req *lowcodeV1.ListFormSubmissionRequest) (*lowcodeV1.ListFormSubmissionResponse, error) {
	input := biz.ListFormSubmissionRequest{}
	err := copier.Copy(&input, req)
	if err != nil {
		return nil, err
	}

	res, err := s.formSubmissionUc.List(ctx, &input)
	if err != nil {
		return nil, err
	}

	output := &lowcodeV1.ListFormSubmissionResponse{}
	if res != nil {
		output.Total = int64(res.Total)
		for _, v := range res.Data {
			item := &lowcodeV1.FormSubmission{
				Id:        v.ID,
				FormCode:  v.FormCode,
				FormName:  v.FormName,
				CreatedAt: timestamppb.New(v.CreatedAt),
				UpdatedAt: timestamppb.New(v.UpdatedAt),
			}
			if v.Data != nil {
				st, err := structpb.NewStruct(v.Data)
				if err != nil {
					return nil, err
				}
				item.SubmitData = st
			}
			output.Items = append(output.Items, item)
		}
	}
	return output, nil
}

// ====== 代码生成器 ======

func (s *LowcodeService) ListGenTable(ctx context.Context, req *lowcodeV1.ListGenTableRequest) (*lowcodeV1.ListGenTableResponse, error) {
	input := biz.ListGenTableRequest{}
	var err error
	err = copier.Copy(&input, req)
	if err != nil {
		return nil, err
	}

	res, err := s.codegenUc.List(ctx, &input)
	if err != nil {
		return nil, err
	}

	output := &lowcodeV1.ListGenTableResponse{}
	if res != nil {
		output.Total = int64(res.Total)
		err = copierx.Copy(&output.Items, &res.Data)
	}

	return output, err
}

func (s *LowcodeService) ListDbTables(ctx context.Context, req *lowcodeV1.ListDbTablesRequest) (*lowcodeV1.ListDbTablesResponse, error) {
	items, err := s.codegenUc.ListDbTables(ctx, req.Keyword)
	if err != nil {
		return nil, err
	}

	output := &lowcodeV1.ListDbTablesResponse{}
	err = copierx.Copy(&output.Items, &items)
	return output, err
}

func (s *LowcodeService) ListDbColumns(ctx context.Context, req *lowcodeV1.ListDbColumnsRequest) (*lowcodeV1.ListDbColumnsResponse, error) {
	items, err := s.codegenUc.ListDbColumns(ctx, req.TableName)
	if err != nil {
		return nil, err
	}

	output := &lowcodeV1.ListDbColumnsResponse{}
	err = copierx.Copy(&output.Items, &items)
	return output, err
}

func (s *LowcodeService) CreateGenTable(ctx context.Context, req *lowcodeV1.CreateGenTableRequest) (*lowcodeV1.GenTable, error) {
	input := biz.GenTable{}
	var err error
	err = copier.Copy(&input, req)
	if err != nil {
		return nil, err
	}
	// copier 无法自动转换 *structpb.Struct -> map[string]any，必须显式转换
	if req.FieldsJson != nil {
		input.FieldsJSON = req.FieldsJson.AsMap()
	}

	table, err := s.codegenUc.CreateGenTable(ctx, &input)
	if err != nil {
		return nil, err
	}

	output := &lowcodeV1.GenTable{}
	err = copierx.Copy(output, table)
	return output, err
}

func (s *LowcodeService) UpdateGenTable(ctx context.Context, req *lowcodeV1.UpdateGenTableRequest) (*lowcodeV1.GenTable, error) {
	input := biz.GenTable{}
	var err error
	err = copier.Copy(&input, req)
	if err != nil {
		return nil, err
	}
	// copier 无法自动转换 *structpb.Struct -> map[string]any，必须显式转换
	if req.FieldsJson != nil {
		input.FieldsJSON = req.FieldsJson.AsMap()
	}

	table, err := s.codegenUc.UpdateGenTable(ctx, &input)
	if err != nil {
		return nil, err
	}

	output := &lowcodeV1.GenTable{}
	err = copierx.Copy(output, table)
	return output, err
}

func (s *LowcodeService) DeleteGenTable(ctx context.Context, req *lowcodeV1.DeleteGenTableRequest) (*emptypb.Empty, error) {
	err := s.codegenUc.DeleteGenTable(ctx, req.Ids)
	return nil, err
}

func (s *LowcodeService) GenerateCode(ctx context.Context, req *lowcodeV1.GenerateCodeRequest) (*lowcodeV1.GenerateCodeResponse, error) {
	input := &biz.GenerateAllInput{
		ID:           req.Id,
		TableName:    req.TableName,
		TableComment: req.TableComment,
		ModuleName:   req.ModuleName,
		BizName:      req.BizName,
		GenType:      req.GenType,
		Status:       req.Status,
		MenuEnabled:  req.GenMenu,
		MenuModule:   req.MenuModule,
		Buttons:      req.Buttons,
	}
	// copier 无法自动转换 *structpb.Struct -> map[string]any，必须显式转换
	if req.FieldsJson != nil {
		input.FieldsJSON = req.FieldsJson.AsMap()
	}

	// 一次调用完成"保存配置 + 生成代码"（大方法编排）
	result, err := s.codegenUc.GenerateAll(ctx, input)
	if err != nil {
		return nil, err
	}

	output := &lowcodeV1.GenerateCodeResponse{
		Id:    result.ID,
		Files: make([]*lowcodeV1.GenCodeFile, 0, len(result.Files)),
	}
	err = copierx.Copy(&output.Files, &result.Files)
	return output, err
}
