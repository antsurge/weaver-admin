// Package storage 提供基于 S3 协议的通用对象存储封装。
//
// 通过自定义 endpoint 即可兼容主流厂商：
//   - 阿里云 OSS:  https://oss-cn-hangzhou.aliyuncs.com
//   - 腾讯云 COS:  https://cos.ap-guangzhou.myqcloud.com
//   - 七牛云 Kodo: https://s3-cn-east-1.qiniucs.com
//   - MinIO 自建:  http://127.0.0.1:9000
package storage

import (
	"context"
	"io"
	"net/url"
	"path"
	"strings"
	"time"

	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
)

// Config 对象存储配置。
type Config struct {
	// Endpoint 服务端点。阿里云 OSS / 腾讯云 COS 等请带上协议（https://）。
	// 若不带协议，将根据 UseSSL 自动补齐。
	Endpoint string
	// AccessKeyID / AccessKeySecret 访问密钥。
	AccessKeyID     string
	AccessKeySecret string
	// Region 区域，部分厂商必填（如腾讯云 COS）。
	Region string
	// Bucket 存储桶名称。
	Bucket string
	// UseSSL 是否使用 HTTPS。当 Endpoint 未显式带协议时生效。
	UseSSL bool
	// PublicURL 自定义公开访问域名（CDN 或绑定域名）。
	// 为空时使用 Endpoint + Bucket 拼装默认访问地址。
	PublicURL string
	// UsePathStyle 是否使用 path-style 访问。
	// MinIO / 七牛云等通常需要 true；阿里云 OSS / 腾讯云 COS 使用 virtual-hosted-style。
	UsePathStyle bool
}

// Storage 通用对象存储客户端。
type Storage struct {
	client       *minio.Client
	bucket       string
	publicURL    string
	endpoint     string
	useSSL       bool
	usePathStyle bool
}

// PutObjectOptions 上传选项。
type PutObjectOptions struct {
	// ContentType 文件内容类型，如 image/png。
	ContentType string
	// ContentDisposition 内容处置方式，如 attachment; filename="a.png"。
	ContentDisposition string
}

// UploadResult 上传结果。
type UploadResult struct {
	// ObjectKey 对象存储中的对象键。
	ObjectKey string `json:"objectKey"`
	// URL 文件访问地址。
	URL string `json:"url"`
	// FileName 原始文件名。
	FileName string `json:"fileName"`
	// Size 文件大小（字节）。
	Size int64 `json:"size"`
}

// New 创建通用对象存储客户端。
func New(cfg *Config) (*Storage, error) {
	endpoint := cfg.Endpoint
	useSSL := cfg.UseSSL

	// 解析 Endpoint 中可能携带的协议，minio 客户端要求传入不带协议的 host。
	if strings.Contains(endpoint, "://") {
		u, err := url.Parse(endpoint)
		if err != nil {
			return nil, err
		}
		useSSL = u.Scheme == "https"
		endpoint = u.Host
	} else {
		// 无协议时按配置补齐
		scheme := "http://"
		if useSSL {
			scheme = "https://"
		}
		if u, err := url.Parse(scheme + endpoint); err == nil {
			endpoint = u.Host
		}
	}

	client, err := minio.New(endpoint, &minio.Options{
		Creds:        credentials.NewStaticV4(cfg.AccessKeyID, cfg.AccessKeySecret, ""),
		Secure:       useSSL,
		Region:       cfg.Region,
		BucketLookup: bucketLookupType(cfg.UsePathStyle),
	})
	if err != nil {
		return nil, err
	}

	publicURL := strings.TrimRight(cfg.PublicURL, "/")

	return &Storage{
		client:       client,
		bucket:       cfg.Bucket,
		publicURL:    publicURL,
		endpoint:     endpoint,
		useSSL:       useSSL,
		usePathStyle: cfg.UsePathStyle,
	}, nil
}

func bucketLookupType(usePathStyle bool) minio.BucketLookupType {
	if usePathStyle {
		return minio.BucketLookupPath
	}
	return minio.BucketLookupAuto
}

// EnsureBucket 确保存储桶存在，不存在则创建。
func (s *Storage) EnsureBucket(ctx context.Context) error {
	exists, err := s.client.BucketExists(ctx, s.bucket)
	if err != nil {
		return err
	}
	if exists {
		return nil
	}
	return s.client.MakeBucket(ctx, s.bucket, minio.MakeBucketOptions{})
}

// Upload 上传对象并返回其可访问 URL。
func (s *Storage) Upload(ctx context.Context, objectKey string, reader io.Reader, size int64, opts *PutObjectOptions) (string, error) {
	key := normalizeKey(objectKey)

	putOpts := minio.PutObjectOptions{}
	if opts != nil {
		putOpts.ContentType = opts.ContentType
		putOpts.ContentDisposition = opts.ContentDisposition
	}

	if _, err := s.client.PutObject(ctx, s.bucket, key, reader, size, putOpts); err != nil {
		return "", err
	}

	return s.URL(key), nil
}

// Delete 删除对象。
func (s *Storage) Delete(ctx context.Context, objectKey string) error {
	return s.client.RemoveObject(ctx, s.bucket, normalizeKey(objectKey), minio.RemoveObjectOptions{})
}

// Exists 判断对象是否存在。
func (s *Storage) Exists(ctx context.Context, objectKey string) (bool, error) {
	_, err := s.client.StatObject(ctx, s.bucket, normalizeKey(objectKey), minio.StatObjectOptions{})
	if err != nil {
		resp := minio.ToErrorResponse(err)
		if resp.Code == "NoSuchKey" {
			return false, nil
		}
		return false, err
	}
	return true, nil
}

// PresignedPutURL 生成直传预签名 URL，供前端直传使用（可选能力）。
func (s *Storage) PresignedPutURL(ctx context.Context, objectKey string, expiry time.Duration) (string, error) {
	u, err := s.client.PresignedPutObject(ctx, s.bucket, normalizeKey(objectKey), expiry)
	if err != nil {
		return "", err
	}
	return u.String(), nil
}

// PresignedGetURL 生成临时可访问的预签名 URL，适用于私有桶。
// 生成的 URL 在 expiry 之前有效，可直接作为文件访问地址返回给前端。
func (s *Storage) PresignedGetURL(ctx context.Context, objectKey string, expiry time.Duration) (string, error) {
	u, err := s.client.PresignedGetObject(ctx, s.bucket, normalizeKey(objectKey), expiry, nil)
	if err != nil {
		return "", err
	}
	return u.String(), nil
}

// URL 返回对象的公开访问地址。
func (s *Storage) URL(objectKey string) string {
	key := normalizeKey(objectKey)

	// 1. 使用自定义域名（CDN / 绑定域名）
	if s.publicURL != "" {
		return s.publicURL + "/" + key
	}

	// 2. 默认拼装地址
	scheme := "http"
	if s.useSSL {
		scheme = "https"
	}
	if s.usePathStyle {
		return scheme + "://" + s.endpoint + "/" + s.bucket + "/" + key
	}
	return scheme + "://" + s.bucket + "." + s.endpoint + "/" + key
}

// Bucket 返回存储桶名称。
func (s *Storage) Bucket() string {
	return s.bucket
}

// normalizeKey 规范化对象键：去除首部斜杠与路径穿越。
func normalizeKey(key string) string {
	key = strings.TrimSpace(key)
	key = strings.TrimPrefix(key, "/")
	return path.Clean(key)
}
