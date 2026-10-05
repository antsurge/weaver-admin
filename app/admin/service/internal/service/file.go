package service

import (
	"context"

	adminV1 "github.com/antsurge/weaver-admin/api/gen/go/admin/service/v1"
	systemV1 "github.com/antsurge/weaver-admin/api/gen/go/system/service/v1"
	"github.com/antsurge/weaver-admin/app/admin/service/internal/biz"
	"github.com/go-kratos/kratos/v2/errors"
	"github.com/go-kratos/kratos/v2/log"
	"google.golang.org/protobuf/types/known/emptypb"
)

// FileService 文件管理服务。
// 注意：上传（UploadFile）与访问（GetFile）由自定义 handler 实现
// （handler.FileHandler，multipart 表单解析 / 私有桶预签名 302 重定向），
// 此处实现 gRPC 契约；删除（DeleteFile）走生成的 HTTP 路由。
type FileService struct {
	adminV1.UnimplementedFileServer

	fileUc *biz.FileUsecase
	log    *log.Helper
}

func NewFileService(
	fileUc *biz.FileUsecase,
	logger log.Logger,
) *FileService {
	return &FileService{
		fileUc: fileUc,
		log:    log.NewHelper(logger),
	}
}

// UploadFile 上传文件（HTTP 层由自定义 handler 处理，此处仅兜底 gRPC 场景）
func (s *FileService) UploadFile(_ context.Context, _ *emptypb.Empty) (*systemV1.UploadFileResponse, error) {
	return nil, errors.BadRequest("FILE_UPLOAD_NOT_SUPPORTED", "文件上传请使用 HTTP 接口")
}

// GetFile 获取文件访问地址（HTTP 层由自定义 handler 处理，此处仅兜底 gRPC 场景）
func (s *FileService) GetFile(_ context.Context, _ *adminV1.GetFileRequest) (*emptypb.Empty, error) {
	return nil, errors.BadRequest("FILE_ACCESS_NOT_SUPPORTED", "文件访问请使用 HTTP 接口")
}

// DeleteFile 删除对象存储中的文件
func (s *FileService) DeleteFile(ctx context.Context, req *systemV1.DeleteFileRequest) (*emptypb.Empty, error) {
	if err := s.fileUc.Delete(ctx, req.ObjectKey); err != nil {
		return nil, err
	}
	return &emptypb.Empty{}, nil
}
