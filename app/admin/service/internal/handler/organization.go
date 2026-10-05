package handler

import (
	"io"
	"net/url"
	"path/filepath"
	"strings"

	organizationV1 "github.com/antsurge/weaver-admin/api/gen/go/organization/service/v1"
	"github.com/antsurge/weaver-admin/app/admin/service/internal/biz"
	"github.com/go-kratos/kratos/v2/errors"
	"github.com/go-kratos/kratos/v2/transport/http"
	"github.com/jinzhu/copier"
	"github.com/xuri/excelize/v2"

	nethttp "net/http"
)

type OrganizationHandler struct {
	positionUc   *biz.PositionUsecase
	departmentUc *biz.DepartmentUsecase
}

func NewOrganizationHandler(positionUc *biz.PositionUsecase, departmentUc *biz.DepartmentUsecase) *OrganizationHandler {
	return &OrganizationHandler{
		positionUc:   positionUc,
		departmentUc: departmentUc,
	}
}

func (h *OrganizationHandler) ExportPosition(ctx http.Context) error {
	var req organizationV1.ListPositionRequest

	// 绑定参数（GET/POST 自动解析）
	if err := ctx.Bind(&req); err != nil {
		return err
	}

	input := biz.ListPositionRequest{}
	if err := copier.Copy(&input, &req); err != nil {
		return err
	}

	data, err := h.positionUc.ExportPosition(ctx, &input)
	if err != nil {
		return err
	}

	fileName := url.QueryEscape("职务列表.xlsx")

	ctx.Response().Header().Set("Content-Type", "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet")
	ctx.Response().Header().Set("Content-Disposition", "attachment; filename*=UTF-8''"+fileName)

	_, err = ctx.Response().Write(data)
	if err != nil {
		return err
	}

	return nil
}

func (h *OrganizationHandler) DownloadPositionTemplate(ctx http.Context) error {
	data, err := h.positionUc.PositionTemplate()
	if err != nil {
		return err
	}

	fileName := url.QueryEscape("职务导入模板.xlsx")

	ctx.Response().Header().Set("Content-Type", "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet")
	ctx.Response().Header().Set("Content-Disposition", "attachment; filename*=UTF-8''"+fileName)

	_, err = ctx.Response().Write(data)
	if err != nil {
		return err
	}

	return nil
}

func (h *OrganizationHandler) ImportPosition(ctx http.Context) error {
	req := ctx.Request()

	// 解析 multipart 表单
	if err := req.ParseMultipartForm(32 << 20); err != nil {
		return err
	}

	// 获取文件
	file, fileHeader, err := ctx.Request().FormFile("file")
	if err != nil {
		return err
	}
	defer file.Close()

	// 校验文件大小
	const maxFileSize = 5 << 20 // 5MB
	if fileHeader.Size > maxFileSize {
		return errors.BadRequest("FILE_TOO_LARGE", "FILE_TOO_LARGE")
	}

	// 校验文件类型（扩展名）
	ext := strings.ToLower(filepath.Ext(fileHeader.Filename))
	allowExt := map[string]bool{
		".xlsx": true,
		".xls":  true,
		".csv":  true,
	}
	if !allowExt[ext] {
		return errors.BadRequest("INVALID_FILE_TYPE", "仅支持 xlsx/xls/csv 文件")
	}

	// ✅ 6. 校验 MIME 类型（更安全）
	buffer := make([]byte, 512)
	if _, err := file.Read(buffer); err != nil {
		return errors.InternalServer("READ_FILE_ERROR", err.Error())
	}

	contentType := nethttp.DetectContentType(buffer)

	allowMime := map[string]bool{
		"application/zip":          true, // ✅ xlsx
		"application/vnd.ms-excel": true,
		"text/csv":                 true,
		"text/plain":               true,
	}

	if !allowMime[contentType] {
		return errors.BadRequest("INVALID_FILE_TYPE", "非法文件类型")
	}

	// ⚠️ 重要：重置文件指针
	if _, err := file.Seek(0, io.SeekStart); err != nil {
		return err
	}

	// ✅ 7. 解析 Excel（示例）
	f, err := excelize.OpenReader(file)
	if err != nil {
		return errors.BadRequest("PARSE_ERROR", "文件解析失败，请检查格式")
	}

	sheet := f.GetSheetName(0)
	rows, err := f.GetRows(sheet)
	if err != nil {
		return errors.BadRequest("PARSE_ERROR", "文件解析失败，请检查格式")
	}

	// 模板结构：第1行标题、第2行说明、第3行表头，数据从第4行开始
	if len(rows) < 4 {
		return errors.BadRequest("EMPTY_FILE", "文件内容为空，请使用导入模板填写数据后上传")
	}

	// 处理数据
	data := rows[3:]

	return h.positionUc.ImportPosition(ctx, data)
}
