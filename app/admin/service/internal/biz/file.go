package biz

import (
	"context"
	"io"
	"path/filepath"
	"strings"
	"time"

	"github.com/antsurge/weaver-admin/app/admin/service/internal/conf"
	"github.com/antsurge/weaver-admin/pkg/storage"
	"github.com/antsurge/weaver-admin/pkg/utils/uuid"
	"github.com/go-kratos/kratos/v2/errors"
	"github.com/go-kratos/kratos/v2/log"
)

// uploadTimeout 上传对象存储的超时时间，独立于服务端请求超时。
const uploadTimeout = 60 * time.Second

// accessURLTimeout 生成访问地址时的超时时间。
const accessURLTimeout = 10 * time.Second

// defaultPresignExpiry 默认预签名有效期。
const defaultPresignExpiry = 15 * time.Minute

// ObjectStorage 定义对象存储能力，便于业务层与具体实现解耦。
type ObjectStorage interface {
	Upload(ctx context.Context, objectKey string, reader io.Reader, size int64, opts *storage.PutObjectOptions) (string, error)
	Delete(ctx context.Context, objectKey string) error
	Exists(ctx context.Context, objectKey string) (bool, error)
	URL(objectKey string) string
	PresignedGetURL(ctx context.Context, objectKey string, expiry time.Duration) (string, error)
}

// FileUploadOptions 文件上传参数。
type FileUploadOptions struct {
	// FileName 原始文件名（含扩展名）。
	FileName string
	// Size 文件大小（字节）。
	Size int64
	// ContentType 文件类型。
	ContentType string
	// Dir 自定义存储目录（相对 upload_prefix），为空时使用默认值。
	Dir string
	// Reader 文件内容。
	Reader io.Reader
}

// FileUsecase 文件上传业务用例。
type FileUsecase struct {
	storage         ObjectStorage
	allowExtensions []string
	maxFileSize     int64
	uploadPrefix    string
	privateBucket   bool
	presignExpiry   time.Duration
	log             *log.Helper
}

// NewFileUsecase 创建文件上传用例。
func NewFileUsecase(
	storageClient ObjectStorage,
	ossCfg *conf.OSS,
	logger log.Logger,
) *FileUsecase {
	uc := &FileUsecase{
		storage:       storageClient,
		presignExpiry: defaultPresignExpiry,
		log:           log.NewHelper(logger),
	}
	if ossCfg != nil {
		for _, ext := range ossCfg.AllowExtensions {
			uc.allowExtensions = append(uc.allowExtensions, strings.ToLower(ext))
		}
		uc.maxFileSize = ossCfg.MaxFileSize
		uc.uploadPrefix = strings.Trim(ossCfg.UploadPrefix, "/")
		uc.privateBucket = ossCfg.Private
		if ossCfg.PresignExpiry != nil && ossCfg.PresignExpiry.AsDuration() > 0 {
			uc.presignExpiry = ossCfg.PresignExpiry.AsDuration()
		}
	}
	return uc
}

// Upload 上传文件并返回访问地址。
func (uc *FileUsecase) Upload(ctx context.Context, opts *FileUploadOptions) (*storage.UploadResult, error) {
	if uc.storage == nil {
		return nil, errors.ServiceUnavailable("STORAGE_DISABLED", "对象存储未启用")
	}
	if opts == nil || opts.FileName == "" {
		return nil, errors.BadRequest("INVALID_FILE", "文件不能为空")
	}

	// 1. 大小校验
	if uc.maxFileSize > 0 && opts.Size > uc.maxFileSize {
		return nil, errors.BadRequest("FILE_TOO_LARGE", "文件大小超出限制")
	}

	// 2. 扩展名校验
	ext := strings.ToLower(filepath.Ext(opts.FileName))
	if len(uc.allowExtensions) > 0 && !containsString(uc.allowExtensions, ext) {
		return nil, errors.BadRequest("INVALID_FILE_TYPE", "不支持的文件类型")
	}

	// 3. 生成对象键：prefix/yyyy/mm/dd/uuid.ext
	dir := opts.Dir
	if dir == "" {
		dir = time.Now().Format("2006/01/02")
	}
	segments := make([]string, 0, 3)
	if uc.uploadPrefix != "" {
		segments = append(segments, uc.uploadPrefix)
	}
	segments = append(segments, strings.Trim(dir, "/"), uuid.GenerateXID()+ext)
	objectKey := strings.Join(segments, "/")

	// 4. 上传（剥离服务端 ctx 的 deadline，单独设置较长的上传超时）
	uploadCtx, cancel := context.WithTimeout(context.WithoutCancel(ctx), uploadTimeout)
	defer cancel()
	url, err := uc.storage.Upload(uploadCtx, objectKey, opts.Reader, opts.Size, &storage.PutObjectOptions{
		ContentType: opts.ContentType,
	})
	if err != nil {
		uc.log.Errorf("upload file failed, key=%s err=%v", objectKey, err)
		return nil, errors.InternalServer("UPLOAD_FAILED", "文件上传失败")
	}

	return &storage.UploadResult{
		ObjectKey: objectKey,
		URL:       url,
		FileName:  opts.FileName,
		Size:      opts.Size,
	}, nil
}

// GetAccessURL 根据对象键生成可访问的文件地址。
// 公有桶直接返回拼接的直链；私有桶返回带签名的临时地址（有效期 presignExpiry）。
func (uc *FileUsecase) GetAccessURL(ctx context.Context, objectKey string) (string, error) {
	if uc.storage == nil {
		return "", errors.ServiceUnavailable("STORAGE_DISABLED", "对象存储未启用")
	}
	key := strings.TrimSpace(objectKey)
	if key == "" {
		return "", errors.BadRequest("INVALID_OBJECT_KEY", "对象键不能为空")
	}

	// 公有桶：直接返回直链
	if !uc.privateBucket {
		return uc.storage.URL(key), nil
	}

	// 私有桶：生成预签名 URL
	signCtx, cancel := context.WithTimeout(context.WithoutCancel(ctx), accessURLTimeout)
	defer cancel()
	url, err := uc.storage.PresignedGetURL(signCtx, key, uc.presignExpiry)
	if err != nil {
		uc.log.Errorf("generate presigned url failed, key=%s err=%v", key, err)
		return "", errors.InternalServer("SIGN_URL_FAILED", "生成文件访问地址失败")
	}
	return url, nil
}

// Delete 删除文件。
func (uc *FileUsecase) Delete(ctx context.Context, objectKey string) error {
	if uc.storage == nil {
		return errors.ServiceUnavailable("STORAGE_DISABLED", "对象存储未启用")
	}
	if strings.TrimSpace(objectKey) == "" {
		return errors.BadRequest("INVALID_OBJECT_KEY", "对象键不能为空")
	}
	if err := uc.storage.Delete(ctx, objectKey); err != nil {
		uc.log.Errorf("delete file failed, key=%s err=%v", objectKey, err)
		return errors.InternalServer("DELETE_FAILED", "文件删除失败")
	}
	return nil
}

// Enabled 是否已启用对象存储。
func (uc *FileUsecase) Enabled() bool {
	return uc.storage != nil
}

func containsString(list []string, s string) bool {
	for _, item := range list {
		if item == s {
			return true
		}
	}
	return false
}
