package biz

import (
	"context"
	"time"

	"github.com/go-kratos/kratos/v2/log"
	"github.com/hypercoze/kratos-admin/pkg/utils/uuid"
)

type Organization struct {
	ID          string    `json:"id"`          // 部门ID
	ParentID    string    `json:"parentID"`    // 父部门ID，空表示根节点
	Name        string    `json:"name"`        // 部门名称
	Code        string    `json:"code"`        // 部门code
	Weight      int       `json:"weight"`      // 权重
	Status      string    `json:"status"`      // 状态：enabled=启用 disabled=禁用
	LeaderName  string    `json:"leaderName"`  // 负责人姓名
	LeaderPhone string    `json:"leaderPhone"` // 联系电话
	LeaderEmail string    `json:"leaderEmail"` // 邮箱
	CreatedAt   time.Time `json:"createdAt"`   // 创建时间
	UpdatedAt   time.Time `json:"updatedAt"`   // 更新时间

	Children []*Organization `json:"children"`
}

type OrganizationListResult struct {
	Name   string `form:"name" query:"name"`
	Code   string `form:"code" query:"code"`
	Status string `form:"status" query:"status"`
}

type OrganizationRepo interface {
	ListOrganization(ctx context.Context, req *OrganizationListResult) ([]*Organization, error)
	CreateOrganization(ctx context.Context, o *Organization) error
	UpdateOrganization(ctx context.Context, o *Organization) error
	DeleteOrganization(ctx context.Context, ids []string) error
	UpdateOrganizationStatus(ctx context.Context, id, status string) error
}

type OrganizationUsecase struct {
	repo OrganizationRepo
	log  *log.Helper
}

func NewOrganizationUsecase(repo OrganizationRepo, logger log.Logger) *OrganizationUsecase {
	return &OrganizationUsecase{
		repo: repo,
		log:  log.NewHelper(logger),
	}
}

// 列表权限的tree
func (uc *OrganizationUsecase) OrganizationTree(ctx context.Context, req *OrganizationListResult) ([]*Organization, error) {
	list, err := uc.repo.ListOrganization(ctx, req)
	if err != nil {
		return nil, err
	}
	tree := buildOrganizationTree(list)

	return tree, nil
}

// 创建
func (uc *OrganizationUsecase) CreateOrganization(ctx context.Context, req *Organization) (*Organization, error) {
	organization := req

	now := time.Now()
	organization.ID = uuid.GenerateXID()
	organization.CreatedAt = now
	organization.UpdatedAt = now

	err := uc.repo.CreateOrganization(ctx, organization)
	return organization, err
}

// 更新
func (uc *OrganizationUsecase) UpdateOrganization(ctx context.Context, req *Organization) (*Organization, error) {
	organization := req

	now := time.Now()
	organization.UpdatedAt = now

	err := uc.repo.UpdateOrganization(ctx, organization)
	return organization, err
}

// 删除
func (uc *OrganizationUsecase) DeleteOrganization(ctx context.Context, ids []string) error {
	return uc.repo.DeleteOrganization(ctx, ids)
}

// 更新状态
func (uc *OrganizationUsecase) UpdateOrganizationStatus(ctx context.Context, id, status string) error {
	return uc.repo.UpdateOrganizationStatus(ctx, id, status)
}
