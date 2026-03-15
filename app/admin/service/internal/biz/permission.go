package biz

import (
	"context"
	"time"

	"github.com/go-kratos/kratos/v2/log"
	permissionV1 "github.com/hypercoze/kratos-admin/api/gen/go/permission/service/v1"
	"github.com/hypercoze/kratos-admin/pkg/utils/uuid"
)

type Permission struct {
	ID          string    `json:"id"`
	ParentId    string    `json:"parentId"`
	Name        string    `json:"name"`
	Code        string    `json:"code"`
	Description string    `json:"description"`
	Path        string    `json:"path"`
	Icon        string    `json:"icon"`
	Type        string    `json:"type"`
	Url         string    `json:"url"`
	Component   string    `json:"component"`
	Weight      int       `json:"weight"`
	Status      string    `json:"status"`
	CreatedAt   time.Time `json:"createdAt"`
	UpdatedAt   time.Time `json:"updatedAt"`

	Children []*Permission `json:"children"`
}

type PermissionRepo interface {
	CreatePermission(ctx context.Context, p *Permission) error
	UpdatePermission(ctx context.Context, p *Permission) error
	UpdatePermissionStatus(ctx context.Context, id, status string) error
	DeletePermission(ctx context.Context, ids []string) error
	ListPermission(ctx context.Context, req *permissionV1.ListPermissionRequest) ([]*Permission, error)
}

type PermissionUsecase struct {
	repo PermissionRepo
	log  *log.Helper
}

func NewPermissionUsecase(repo PermissionRepo, logger log.Logger) *PermissionUsecase {
	return &PermissionUsecase{
		repo: repo,
		log:  log.NewHelper(logger),
	}
}

// 列表权限的tree
func (uc *PermissionUsecase) PermissionTree(ctx context.Context, req *permissionV1.PermissionTreeRequest) ([]*Permission, error) {
	list, err := uc.repo.ListPermission(ctx, &permissionV1.ListPermissionRequest{
		Page:     0,
		PageSize: 0,
		Keyword:  "",
		ParentId: "",
		Type:     "",
	})
	if err != nil {
		return nil, err
	}
	tree := buildPermissionTree(list)

	return tree, nil
}

// 创建权限
func (uc *PermissionUsecase) CreatePermission(ctx context.Context, req *Permission) (*Permission, error) {
	permission := req

	now := time.Now()
	permission.ID = uuid.GenerateXID()
	permission.CreatedAt = now
	permission.UpdatedAt = now

	err := uc.repo.CreatePermission(ctx, req)
	return permission, err
}

// 更新权限
func (uc *PermissionUsecase) UpdatePermission(ctx context.Context, req *Permission) (*Permission, error) {
	permission := req

	now := time.Now()
	permission.UpdatedAt = now

	err := uc.repo.UpdatePermission(ctx, permission)

	return permission, err
}

// 删除权限
func (uc *PermissionUsecase) DeletePermission(ctx context.Context, ids []string) error {
	return uc.repo.DeletePermission(ctx, ids)
}

// 更新权限状态
func (uc *PermissionUsecase) UpdatePermissionStatus(ctx context.Context, id, status string) error {
	return uc.repo.UpdatePermissionStatus(ctx, id, status)
}
