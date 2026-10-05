package data

import (
	"context"

	"github.com/antsurge/weaver-admin/app/admin/service/internal/biz"
	"github.com/antsurge/weaver-admin/app/admin/service/internal/conf"
	"github.com/antsurge/weaver-admin/pkg/storage"
	"github.com/go-kratos/kratos/v2/log"
)

// NewObjectStorage 初始化通用对象存储客户端（S3 协议，兼容阿里云 OSS / 腾讯云 COS / MinIO）。
// 未启用时返回 nil，由上层按需降级。
func NewObjectStorage(c *conf.OSS, logger log.Logger) *storage.Storage {
	l := log.NewHelper(logger)

	if c == nil || !c.Enabled {
		l.Warn("object storage is disabled (oss.enabled=false)")
		return nil
	}

	st, err := storage.New(&storage.Config{
		Endpoint:        c.Endpoint,
		AccessKeyID:     c.AccessKeyId,
		AccessKeySecret: c.AccessKeySecret,
		Region:          c.Region,
		Bucket:          c.Bucket,
		UseSSL:          c.UseSsl,
		PublicURL:       c.PublicUrl,
		UsePathStyle:    c.UsePathStyle,
	})
	if err != nil {
		l.Errorf("failed to init object storage: %v", err)
		return nil
	}

	// 尝试确保 bucket 存在（失败仅告警，不阻塞启动）
	if err := st.EnsureBucket(context.Background()); err != nil {
		l.Warnf("failed to ensure bucket %q: %v", c.Bucket, err)
	} else {
		l.Infof("object storage initialized, endpoint=%s bucket=%s", c.Endpoint, c.Bucket)
	}

	return st
}

// NewObjectStorageProvider 将 *storage.Storage 适配为 biz.ObjectStorage 接口，
// 未启用或初始化失败时返回 nil（业务层会做降级处理）。
func NewObjectStorageProvider(st *storage.Storage) biz.ObjectStorage {
	if st == nil {
		return nil
	}
	return st
}
