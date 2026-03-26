package biz

import (
	"context"
	"time"

	"github.com/go-kratos/kratos/v2/log"
	"github.com/hypercoze/kratos-admin/pkg/enthelper"
	"github.com/hypercoze/kratos-admin/pkg/utils/uuid"
)

type Role struct {
	ID        string    `json:"id"`
	Name      string    `json:"name"`
	Code      string    `json:"code"`
	Remark    string    `json:"remark"`
	Weight    int       `json:"weight"`
	Status    string    `json:"status"`
	DataScope string    `json:"dataScope"`
	IsSystem  bool      `json:"isSystem"`
	CreatedAt time.Time `json:"createdAt"`
	UpdatedAt time.Time `json:"updatedAt"`
	MenuIDs   []string  `json:"menuIds"`
}

type ListRoleRequest struct {
	enthelper.PaginationParams
	Name   string `form:"name" query:"name"`
	Code   string `form:"code" query:"code"`
	Status string `form:"status" query:"status"`
}

type ListRoleResponse struct {
	Items []*Role `json:"items"`
	Total int     `json:"total"`
}

type ListRoleOption struct {
	enthelper.QueryOption
	Name   string `form:"name" query:"name"`
	Code   string `form:"code" query:"code"`
	Status string `form:"status" query:"status"`
}

type RoleRepo interface {
	ListRole(ctx context.Context, params *ListRoleRequest, opts ...*ListRoleOption) (*ListRoleResponse, error)
	GetRole(ctx context.Context, id string) (*Role, error)
	CreateRole(ctx context.Context, role *Role) error
	UpdateRole(ctx context.Context, role *Role) error
	UpdateRoleStatus(ctx context.Context, id string, status string) error
	DeleteRole(ctx context.Context, ids []string) error
}

type RoleUsecase struct {
	repo RoleRepo
	log  *log.Helper
}

func NewRoleUsecase(repo RoleRepo, logger log.Logger) *RoleUsecase {
	return &RoleUsecase{repo: repo, log: log.NewHelper(logger)}
}

func (uc *RoleUsecase) ListRole(ctx context.Context, params *ListRoleRequest) (*ListRoleResponse, error) {
	return uc.repo.ListRole(ctx, params)
}

func (uc *RoleUsecase) GetRole(ctx context.Context, id string) (*Role, error) {
	return uc.repo.GetRole(ctx, id)
}

func (uc *RoleUsecase) CreateRole(ctx context.Context, role *Role) (*Role, error) {
	now := time.Now()
	role.ID = uuid.GenerateXID()
	role.CreatedAt = now
	role.UpdatedAt = now

	err := uc.repo.CreateRole(ctx, role)
	return role, err
}

func (uc *RoleUsecase) UpdateRole(ctx context.Context, role *Role) (*Role, error) {
	now := time.Now()
	role.UpdatedAt = now
	err := uc.repo.UpdateRole(ctx, role)

	return role, err
}

func (uc *RoleUsecase) UpdateRoleStatus(ctx context.Context, id string, status string) error {
	return uc.repo.UpdateRoleStatus(ctx, id, status)
}

func (uc *RoleUsecase) DeleteRole(ctx context.Context, ids []string) error {
	return uc.repo.DeleteRole(ctx, ids)
}
