package handler

import (
	"io"
	"path/filepath"
	"strings"

	"github.com/antsurge/weaver-admin/app/admin/service/internal/biz"
	"github.com/go-kratos/kratos/v2/errors"
	"github.com/go-kratos/kratos/v2/transport/http"
)

type SystemHandler struct {
	apiInterfaceUc *biz.ApiInterfaceUsecase
}

func NewSystemHandler(apiInterfaceUc *biz.ApiInterfaceUsecase) *SystemHandler {
	return &SystemHandler{
		apiInterfaceUc: apiInterfaceUc,
	}
}

// ImportApiInterface 导入 openapi.yaml 文件
func (h *SystemHandler) ImportApiInterface(ctx http.Context) error {
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

	// 校验文件大小（最大 10MB）
	const maxFileSize = 10 << 20
	if fileHeader.Size > maxFileSize {
		return errors.BadRequest("FILE_TOO_LARGE", "FILE_TOO_LARGE")
	}

	// 校验文件类型（扩展名）
	ext := strings.ToLower(filepath.Ext(fileHeader.Filename))
	allowExt := map[string]bool{
		".yaml": true,
		".yml":  true,
	}
	if !allowExt[ext] {
		return errors.BadRequest("INVALID_FILE_TYPE", "仅支持 yaml/yml 文件")
	}

	// 读取文件内容
	dataBytes, err := io.ReadAll(file)
	if err != nil {
		return errors.InternalServer("READ_FILE_ERROR", "读取文件失败: "+err.Error())
	}

	// 解析 openapi 文件
	items, err := biz.ParseOpenAPIFile(dataBytes)
	if err != nil {
		return errors.BadRequest("PARSE_OPENAPI_ERROR", "解析 openapi 文件失败: "+err.Error())
	}

	// 调用 biz 层导入
	result, err := h.apiInterfaceUc.Import(ctx, items)
	if err != nil {
		return err
	}

	// 返回 JSON 结果
	return ctx.Result(200, result)
}