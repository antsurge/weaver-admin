package biz

import (
	"bytes"
	"context"
	"strconv"
	"time"

	"github.com/go-kratos/kratos/v2/errors"

	"github.com/go-kratos/kratos/v2/log"
	"github.com/hypercoze/kratos-admin/pkg/enthelper"
	"github.com/hypercoze/kratos-admin/pkg/metadata"
	"github.com/hypercoze/kratos-admin/pkg/utils/excel"
	uuid "github.com/hypercoze/kratos-admin/pkg/utils/uuid"
	"github.com/xuri/excelize/v2"
)

// 职务
type Position struct {
	ID        string    `json:"id"`
	Name      string    `json:"name"`
	Code      string    `json:"code"`
	Weight    int       `json:"weight"`
	Status    string    `json:"status"`
	Remark    string    `json:"remark"`
	CreatedAt time.Time `json:"createdAt"`
	CreatedBy string    `json:"createdBy"`
	UpdatedAt time.Time `json:"updatedAt"`
	UpdatedBy string    `json:"updatedBy"`
}

type PositionRepo interface {
	ListPosition(context.Context, *ListPositionRequest, ...*ListPositionOption) (*ListPositionResponse, error)
	GetPosition(ctx context.Context, id string) (*Position, error)
	CreatePosition(ctx context.Context, o *Position) error
	BatchCreatePosition(ctx context.Context, o []*Position) error
	UpdatePosition(ctx context.Context, o *Position) error
	DeletePosition(ctx context.Context, ids []string) error
	UpdatePositionStatus(ctx context.Context, id, status string) error
	IsPositionNameExists(ctx context.Context, name, id string) (bool, error)
	IsPositionCodeExists(ctx context.Context, code, id string) (bool, error)
}

type ListPositionRequest struct {
	enthelper.PaginationParams
	Name   string   `form:"name" query:"name"`
	Code   string   `form:"code" query:"code"`
	Status string   `form:"status" query:"status"`
	Names  []string `form:"names" query:"names"`
	Codes  []string `form:"codes" query:"codes"`
}

type ListPositionResponse struct {
	Items []*Position
	Total int
}

type ListPositionOption struct {
	enthelper.QueryOption
}

type PositionUsecase struct {
	repo PositionRepo
	log  *log.Helper
}

func NewPositionUsecase(repo PositionRepo, logger log.Logger) *PositionUsecase {
	return &PositionUsecase{
		repo: repo,
		log:  log.NewHelper(logger),
	}
}

func (uc *PositionUsecase) ListPosition(ctx context.Context, req *ListPositionRequest, opts ...*ListPositionOption) (*ListPositionResponse, error) {
	return uc.repo.ListPosition(ctx, req, opts...)
}

func (uc *PositionUsecase) GetPosition(ctx context.Context, id string) (*Position, error) {
	return uc.repo.GetPosition(ctx, id)
}

func (uc *PositionUsecase) CreatePosition(ctx context.Context, position *Position) (*Position, error) {
	now := time.Now()
	adminID := metadata.GetAdminID(ctx)
	position.ID = uuid.GenerateXID()
	position.CreatedAt = now
	position.UpdatedAt = now
	position.CreatedBy = adminID
	position.UpdatedBy = adminID

	// 校验唯一性
	if err := uc.validateUnique(ctx, position); err != nil {
		return nil, err
	}

	err := uc.repo.CreatePosition(ctx, position)
	return position, err
}

func (uc *PositionUsecase) UpdatePosition(ctx context.Context, position *Position) (*Position, error) {
	adminID := metadata.GetAdminID(ctx)
	position.UpdatedAt = time.Now()
	position.UpdatedBy = adminID

	// 校验唯一性
	if err := uc.validateUnique(ctx, position); err != nil {
		return nil, err
	}

	err := uc.repo.UpdatePosition(ctx, position)

	return position, err
}

func (uc *PositionUsecase) DeletePosition(ctx context.Context, ids []string) error {
	return uc.repo.DeletePosition(ctx, ids)
}

func (uc *PositionUsecase) UpdatePositionStatus(ctx context.Context, id, status string) error {
	return uc.repo.UpdatePositionStatus(ctx, id, status)
}

func (uc *PositionUsecase) IsPositionNameExists(ctx context.Context, name, id string) (bool, error) {
	return uc.repo.IsPositionNameExists(ctx, name, id)
}

func (uc *PositionUsecase) IsPositionCodeExists(ctx context.Context, code, id string) (bool, error) {
	return uc.repo.IsPositionCodeExists(ctx, code, id)
}

func (uc *PositionUsecase) validateUnique(ctx context.Context, position *Position) error {
	nameExists, err := uc.IsPositionNameExists(ctx, position.Name, position.ID)
	if err != nil {
		return err
	}
	if nameExists {
		return errors.New(400, "position name already exists", "")
	}

	codeExists, err := uc.IsPositionCodeExists(ctx, position.Code, position.ID)
	if err != nil {
		return err
	}
	if codeExists {
		return errors.New(400, "position code already exists", "")
	}
	return nil
}

func (uc *PositionUsecase) ExportPosition(ctx context.Context, req *ListPositionRequest, opts ...*ListPositionOption) ([]byte, error) {
	// 不用执行分页
	req.PaginationParams.PageSize = 0

	res, err := uc.repo.ListPosition(ctx, req, opts...)
	if err != nil {
		return nil, err
	}
	data := res.Items

	// 创建 Excel
	f := excelize.NewFile()
	sheet := "Sheet1"
	f.SetSheetName("Sheet1", sheet)

	// 创建样式
	headerStyle, _ := excel.NewHeaderStyle(f, excel.AlignCenter, true) // 表头居中加粗
	contentStyle, _ := excel.NewContentStyle(f, excel.AlignCenter)     // 内容居中
	if err := excel.ApplySheetTemaplte(f, sheet, "标题", "导出人", &excel.SheetTemplateOptions{ColCount: 5}); err != nil {
		return nil, err
	}

	// 表头
	headers := []string{"职务名称", "职务编码", "职务级别", "状态", "备注"}
	for index, header := range headers {
		cell, _ := excelize.CoordinatesToCellName(index+1, 3)
		f.SetCellValue(sheet, cell, header)
		f.SetCellStyle(sheet, cell, cell, headerStyle)
	}

	// 数据
	for i, item := range data {
		row := i + 4
		cols := []interface{}{
			item.Name,
			item.Code,
			item.Weight,
			item.Status,
			item.Remark,
		}

		for colIdx, val := range cols {
			cell, _ := excelize.CoordinatesToCellName(colIdx+1, row)
			f.SetCellValue(sheet, cell, val)
			f.SetCellStyle(sheet, cell, cell, contentStyle) // 自动应用内容样式
		}
	}
	_ = excel.AutoAdjustColumnWidth(f, sheet, 1, 0)

	// 写入 buffer
	var buf bytes.Buffer
	if err := f.Write(&buf); err != nil {
		return nil, err
	}

	return buf.Bytes(), nil
}

func (uc *PositionUsecase) ImportPosition(ctx context.Context, data [][]string) error {
	// 处理name和编码
	var names, codes []string
	for _, v := range data {
		names = append(names, v[0])
		codes = append(codes, v[1])
	}

	// 查询编码和名称是否重复
	codesExistsResult, err := uc.repo.ListPosition(ctx, &ListPositionRequest{
		Codes: codes,
	})
	if err != nil {
		return err
	}
	namesExistsResult, err := uc.repo.ListPosition(ctx, &ListPositionRequest{
		Names: names,
	})
	if err != nil {
		return err
	}

	codeMap := make(map[string]bool, len(codesExistsResult.Items))
	for _, item := range codesExistsResult.Items {
		codeMap[item.Code] = true
	}

	// name -> true
	nameMap := make(map[string]bool, len(namesExistsResult.Items))
	for _, item := range namesExistsResult.Items {
		nameMap[item.Name] = true
	}

	now := time.Now()
	adminID := metadata.GetAdminID(ctx)

	insertData := make([]*Position, 0, len(data))
	for _, item := range data {
		name := item[0]
		code := item[1]
		weight := item[2]
		status := item[3]
		remark := item[4]
		weightInt, err := strconv.Atoi(weight)
		if err != nil {
			continue
		}

		if codeMap[code] {
			continue
		}
		if nameMap[name] {
			continue
		}

		insertData = append(insertData, &Position{
			ID:        uuid.GenerateXID(),
			Name:      name,
			Code:      code,
			Weight:    weightInt,
			Status:    status,
			Remark:    remark,
			CreatedAt: now,
			CreatedBy: adminID,
			UpdatedAt: now,
			UpdatedBy: adminID,
		})
	}

	if len(insertData) == 0 {
		return nil
	}

	return uc.repo.BatchCreatePosition(ctx, insertData)
}
