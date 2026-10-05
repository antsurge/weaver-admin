package handler

import (
	"errors"
	"io"
	nethttp "net/http"
	"strings"

	"github.com/antsurge/weaver-admin/app/admin/service/internal/biz"
	"github.com/go-kratos/kratos/v2/log"
	khttp "github.com/go-kratos/kratos/v2/transport/http"
	"github.com/xuri/excelize/v2"
)

// OrderHandler 订单导入导出处理器
type OrderHandler struct {
	uc  *biz.OrderUsecase
	log *log.Helper
}

func NewOrderHandler(uc *biz.OrderUsecase, logger log.Logger) *OrderHandler {
	return &OrderHandler{uc: uc, log: log.NewHelper(logger)}
}

// ExportOrder 导出订单（xlsx 二进制流）
func (h *OrderHandler) ExportOrder(ctx khttp.Context) error {
	req := ctx.Request()
	if err := req.ParseForm(); err != nil {
		return err
	}
	pageSize := int64(0)
	list, err := h.uc.List(ctx, &biz.ListOrderRequest{})
	_ = pageSize
	if err != nil {
		return err
	}

	f := excelize.NewFile()
	defer f.Close()
	sheet := "Sheet1"
	f.SetSheetName("Sheet1", sheet)
	var cell string
	cell, _ = excelize.CoordinatesToCellName(1, 1)
	_ = f.SetCellValue(sheet, cell, "name")
	cell, _ = excelize.CoordinatesToCellName(2, 1)
	_ = f.SetCellValue(sheet, cell, "sn")

	for i, v := range list.Data {
		row := i + 2
		cell, _ = excelize.CoordinatesToCellName(1, row)
		_ = f.SetCellValue(sheet, cell, v.Name)
		cell, _ = excelize.CoordinatesToCellName(2, row)
		_ = f.SetCellValue(sheet, cell, v.Sn)
	}

	ctx.Response().Header().Set("Content-Type", "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet")
	ctx.Response().Header().Set("Content-Disposition", "attachment; filename="+"order.xlsx")
	ctx.Response().WriteHeader(nethttp.StatusOK)
	_, err = f.WriteTo(ctx.Response())
	return err
}

// ImportOrder 导入订单（multipart/form-data 文件上传）
func (h *OrderHandler) ImportOrder(ctx khttp.Context) error {
	req := ctx.Request()
	if err := req.ParseMultipartForm(32 << 20); err != nil {
		return errors.New("请上传文件")
	}
	file, _, err := req.FormFile("file")
	if err != nil {
		return errors.New("请上传文件")
	}
	defer file.Close()

	// 限制文件大小 10MB
	data, err := io.ReadAll(io.LimitReader(file, 10<<20))
	if err != nil {
		return errors.New("文件读取失败")
	}
	if len(data) == 0 {
		return errors.New("文件内容为空")
	}

	f, err := excelize.OpenReader(strings.NewReader(string(data)))
	if err != nil {
		return errors.New("文件解析失败，请上传正确的 Excel 文件")
	}
	defer f.Close()

	rows, err := f.GetRows("Sheet1")
	if err != nil || len(rows) < 2 {
		return errors.New("文件无有效数据")
	}

	count := 0
	for _, row := range rows[1:] {
		item := &biz.Order{}
		if len(row) > 1 {
			item.Name = row[1]
		}
		if len(row) > 2 {
			item.Sn = row[2]
		}
		if _, err := h.uc.Create(ctx, item); err == nil {
			count++
		}
	}
	h.log.Infof("导入订单成功 %d 条", count)
	return nil
}
