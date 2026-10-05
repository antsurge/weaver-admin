package handler

import (
	"path/filepath"
	"strings"

	"github.com/antsurge/weaver-admin/app/admin/service/internal/biz"
	"github.com/go-kratos/kratos/v2/errors"
	"github.com/go-kratos/kratos/v2/transport/http"
)

// accessURLRedirect 访问地址重定向的默认状态码提示。
const accessURLRedirectCode = 302

// FileHandler 文件上传相关的自定义 HTTP 处理（multipart）。
type FileHandler struct {
	fileUc *biz.FileUsecase
}

func NewFileHandler(fileUc *biz.FileUsecase) *FileHandler {
	return &FileHandler{fileUc: fileUc}
}

// UploadFile 上传文件到对象存储。
// 支持表单字段：
//   - file: 文件（必填）
//   - dir:  自定义存储目录（可选）
func (h *FileHandler) UploadFile(ctx http.Context) error {
	req := ctx.Request()

	// 解析 multipart 表单（最大 32MB 内存缓冲）
	if err := req.ParseMultipartForm(32 << 20); err != nil {
		return errors.BadRequest("PARSE_FORM_ERROR", "解析上传表单失败")
	}

	file, fileHeader, err := req.FormFile("file")
	if err != nil {
		return errors.BadRequest("FILE_REQUIRED", "请选择要上传的文件")
	}
	defer file.Close()

	contentType := fileHeader.Header.Get("Content-Type")
	if contentType == "" {
		contentType = guessContentType(fileHeader.Filename)
	}

	result, err := h.fileUc.Upload(ctx, &biz.FileUploadOptions{
		FileName:    fileHeader.Filename,
		Size:        fileHeader.Size,
		ContentType: contentType,
		Dir:         req.FormValue("dir"),
		Reader:      file,
	})
	if err != nil {
		return err
	}

	return ctx.Result(200, result)
}

// GetFile 返回文件访问地址：私有桶下生成带签名的临时地址并 302 重定向。
// 传参：GET /admin/v1/file?objectKey=uploads/avatar/xxx.png
func (h *FileHandler) GetFile(ctx http.Context) error {
	objectKey := ctx.Query().Get("objectKey")
	objectKey = strings.TrimSpace(objectKey)
	if objectKey == "" {
		return errors.BadRequest("INVALID_OBJECT_KEY", "对象键不能为空")
	}

	accessURL, err := h.fileUc.GetAccessURL(ctx, objectKey)
	if err != nil {
		return err
	}

	ctx.Response().Header().Set("Cache-Control", "no-store")
	ctx.Response().Header().Set("Location", accessURL)
	ctx.Response().WriteHeader(accessURLRedirectCode)
	return nil
}

// guessContentType 根据扩展名粗略推断 Content-Type，避免上传后浏览器按二进制下载。
func guessContentType(filename string) string {
	switch strings.ToLower(filepath.Ext(filename)) {
	case ".jpg", ".jpeg":
		return "image/jpeg"
	case ".png":
		return "image/png"
	case ".gif":
		return "image/gif"
	case ".webp":
		return "image/webp"
	case ".svg":
		return "image/svg+xml"
	case ".pdf":
		return "application/pdf"
	case ".txt":
		return "text/plain; charset=utf-8"
	case ".json":
		return "application/json"
	case ".xlsx":
		return "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet"
	case ".xls":
		return "application/vnd.ms-excel"
	case ".doc":
		return "application/msword"
	case ".docx":
		return "application/vnd.openxmlformats-officedocument.wordprocessingml.document"
	default:
		return "application/octet-stream"
	}
}
