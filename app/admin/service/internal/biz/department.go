package biz

import (
	"context"
	"time"

	"github.com/go-kratos/kratos/v2/log"
	"github.com/hypercoze/kratos-admin/pkg/utils/uuid"
)

type Department struct {
	ID          string    `json:"id"`          // 部门ID
	ParentID    string    `json:"parentID"`    // 父部门ID，空表示根节点
	Name        string    `json:"name"`        // 部门名称
	Code        string    `json:"code"`        // 部门code
	Type        string    `json:"type"`        // 类型（company=公司 subsidiary=子公司 department=部门 position=岗位）
	Weight      int       `json:"weight"`      // 权重
	Status      string    `json:"status"`      // 状态：enabled=启用 disabled=禁用
	LeaderName  string    `json:"leaderName"`  // 负责人姓名
	LeaderPhone string    `json:"leaderPhone"` // 联系电话
	LeaderEmail string    `json:"leaderEmail"` // 邮箱
	CreatedAt   time.Time `json:"createdAt"`   // 创建时间
	UpdatedAt   time.Time `json:"updatedAt"`   // 更新时间

	Children []*Department `json:"children"`
}

type DepartmentListResult struct {
	Name   string `form:"name" query:"name"`
	Code   string `form:"code" query:"code"`
	Status string `form:"status" query:"status"`
}

type DepartmentRepo interface {
	ListDepartment(ctx context.Context, req *DepartmentListResult) ([]*Department, error)
	CreateDepartment(ctx context.Context, o *Department) error
	UpdateDepartment(ctx context.Context, o *Department) error
	DeleteDepartment(ctx context.Context, ids []string) error
	UpdateDepartmentStatus(ctx context.Context, id, status string) error
	GetDepartment(ctx context.Context, id string) (*Department, error)
}

type DepartmentUsecase struct {
	repo DepartmentRepo
	log  *log.Helper
}

func NewDepartmentUsecase(repo DepartmentRepo, logger log.Logger) *DepartmentUsecase {
	return &DepartmentUsecase{
		repo: repo,
		log:  log.NewHelper(logger),
	}
}

// 列表权限的tree
func (uc *DepartmentUsecase) DepartmentTree(ctx context.Context, req *DepartmentListResult) ([]*Department, error) {
	list, err := uc.repo.ListDepartment(ctx, req)
	if err != nil {
		return nil, err
	}
	tree := buildDepartmentTree(list)

	return tree, nil
}

func (uc *DepartmentUsecase) GetDepartment(ctx context.Context, id string) (*Department, error) {
	return uc.repo.GetDepartment(ctx, id)
}

// 创建
func (uc *DepartmentUsecase) CreateDepartment(ctx context.Context, req *Department) (*Department, error) {
	department := req

	now := time.Now()
	department.ID = uuid.GenerateXID()
	department.CreatedAt = now
	department.UpdatedAt = now

	err := uc.repo.CreateDepartment(ctx, department)
	return department, err
}

// 更新
func (uc *DepartmentUsecase) UpdateDepartment(ctx context.Context, req *Department) (*Department, error) {
	department := req

	now := time.Now()
	department.UpdatedAt = now

	err := uc.repo.UpdateDepartment(ctx, department)
	return department, err
}

// 删除
func (uc *DepartmentUsecase) DeleteDepartment(ctx context.Context, ids []string) error {
	return uc.repo.DeleteDepartment(ctx, ids)
}

// 更新状态
func (uc *DepartmentUsecase) UpdateDepartmentStatus(ctx context.Context, id, status string) error {
	return uc.repo.UpdateDepartmentStatus(ctx, id, status)
}
