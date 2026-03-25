package biz

import (
	"context"
	"time"

	"github.com/go-kratos/kratos/v2/log"
	"github.com/hypercoze/kratos-admin/pkg/utils/uuid"
)

type Menu struct {
	ID        string    `json:"id"`
	ParentID  string    `json:"parentID"`
	Name      string    `json:"name"`
	Code      string    `json:"code"`
	Remark    string    `json:"remark"`
	Path      string    `json:"path"`
	Icon      string    `json:"icon"`
	Type      string    `json:"type"`
	Url       string    `json:"url"`
	Component string    `json:"component"`
	Weight    int       `json:"weight"`
	Status    string    `json:"status"`
	CreatedAt time.Time `json:"createdAt"`
	UpdatedAt time.Time `json:"updatedAt"`

	Children []*Menu `json:"children"`
}

type MenuRepo interface {
	CreateMenu(ctx context.Context, p *Menu) error
	UpdateMenu(ctx context.Context, p *Menu) error
	UpdateMenuStatus(ctx context.Context, id, status string) error
	DeleteMenu(ctx context.Context, ids []string) error
	ListMenu(ctx context.Context, req *ListMenuRequest) ([]*Menu, error)
}

type ListMenuRequest struct {
	Name   string
	Code   string
	Status string
}

type MenuUsecase struct {
	repo MenuRepo
	log  *log.Helper
}

func NewMenuUsecase(repo MenuRepo, logger log.Logger) *MenuUsecase {
	return &MenuUsecase{
		repo: repo,
		log:  log.NewHelper(logger),
	}
}

// 列表权限的tree
func (uc *MenuUsecase) MenuTree(ctx context.Context, req *ListMenuRequest) ([]*Menu, error) {
	list, err := uc.repo.ListMenu(ctx, req)
	if err != nil {
		return nil, err
	}
	tree := buildMenuTree(list)

	return tree, nil
}

// 创建权限
func (uc *MenuUsecase) CreateMenu(ctx context.Context, req *Menu) (*Menu, error) {
	permission := req

	now := time.Now()
	permission.ID = uuid.GenerateXID()
	permission.CreatedAt = now
	permission.UpdatedAt = now

	err := uc.repo.CreateMenu(ctx, req)
	return permission, err
}

// 更新权限
func (uc *MenuUsecase) UpdateMenu(ctx context.Context, req *Menu) (*Menu, error) {
	permission := req

	now := time.Now()
	permission.UpdatedAt = now

	err := uc.repo.UpdateMenu(ctx, permission)

	return permission, err
}

// 删除权限
func (uc *MenuUsecase) DeleteMenu(ctx context.Context, ids []string) error {
	return uc.repo.DeleteMenu(ctx, ids)
}

// 更新权限状态
func (uc *MenuUsecase) UpdateMenuStatus(ctx context.Context, id, status string) error {
	return uc.repo.UpdateMenuStatus(ctx, id, status)
}
