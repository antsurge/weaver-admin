package data

import (
	"context"
	"time"

	"github.com/antsurge/weaver-admin/app/admin/service/internal/biz"
	"github.com/antsurge/weaver-admin/app/admin/service/internal/data/ent"
	"github.com/antsurge/weaver-admin/app/admin/service/internal/data/ent/department"
	"github.com/go-kratos/kratos/v2/errors"
	"github.com/go-kratos/kratos/v2/log"
)

type departmentRepo struct {
	data *Data
	log  *log.Helper
}

func NewDepartmentRepo(data *Data, logger log.Logger) biz.DepartmentRepo {
	return &departmentRepo{
		data: data,
		log:  log.NewHelper(logger),
	}
}

func (r *departmentRepo) ListDepartment(ctx context.Context, params *biz.DepartmentListResult) ([]*biz.Department, error) {
	query := r.data.db.Department.Query().
		Order(ent.Desc(department.FieldWeight))

	// 名称
	if v := params.Name; len(v) > 0 {
		query = query.Where(department.NameContains(v))
	}

	// code
	if v := params.Code; len(v) > 0 {
		query = query.Where(department.CodeContains(v))
	}

	// 状态
	if v := params.Status; len(v) > 0 {
		query = query.Where(department.StatusEQ(department.Status(v)))
	}

	list, err := query.All(ctx)

	res := make([]*biz.Department, 0, len(list))
	if err != nil {
		return nil, err
	}

	for _, v := range list {
		res = append(res, &biz.Department{
			ID:          v.ID,
			ParentID:    v.ParentID,
			Name:        v.Name,
			Code:        v.Code,
			Weight:      v.Weight,
			Status:      string(v.Status),
			LeaderName:  v.LeaderName,
			LeaderPhone: v.LeaderPhone,
			LeaderEmail: v.LeaderEmail,
			CreatedAt:   v.CreatedAt,
			UpdatedAt:   v.UpdatedAt,
		})
	}

	return res, nil
}

func (r *departmentRepo) CreateDepartment(ctx context.Context, req *biz.Department) error {
	_, err := r.data.db.Department.Create().
		SetID(req.ID).
		SetParentID(req.ParentID).
		SetName(req.Name).
		SetCode(req.Code).
		SetLeaderName(req.LeaderName).
		SetLeaderPhone(req.LeaderPhone).
		SetLeaderEmail(req.LeaderEmail).
		SetWeight(req.Weight).
		SetStatus(department.Status(req.Status)).
		SetCreatedAt(req.CreatedAt).
		SetUpdatedAt(req.UpdatedAt).
		Save(ctx)
	return err
}

func (r *departmentRepo) UpdateDepartment(ctx context.Context, req *biz.Department) error {
	_, err := r.data.db.Department.UpdateOneID(req.ID).
		SetParentID(req.ParentID).
		SetName(req.Name).
		SetCode(req.Code).
		SetLeaderName(req.LeaderName).
		SetLeaderPhone(req.LeaderPhone).
		SetLeaderEmail(req.LeaderEmail).
		SetWeight(req.Weight).
		SetStatus(department.Status(req.Status)).
		SetUpdatedAt(req.UpdatedAt).
		Save(ctx)
	return err
}

func (r *departmentRepo) DeleteDepartment(ctx context.Context, ids []string) error {
	if len(ids) == 0 {
		return nil
	}

	var allIDs []string

	// 递归收集子孙节点
	var collect func(ids []string) error
	collect = func(ids []string) error {
		if len(ids) == 0 {
			return nil
		}

		children, err := r.data.db.Department.
			Query().
			Where(department.ParentIDIn(ids...)).
			All(ctx)
		if err != nil {
			return err
		}

		// 收集子节点 ID
		childIDs := make([]string, len(children))
		for i, c := range children {
			childIDs[i] = c.ID
		}
		allIDs = append(allIDs, childIDs...)

		// 递归查子节点
		return collect(childIDs)
	}

	// 初始化 allIDs
	allIDs = append(allIDs, ids...)
	if err := collect(ids); err != nil {
		return err
	}

	// 删除所有节点
	_, err := r.data.db.Department.
		Delete().
		Where(department.IDIn(allIDs...)).
		Exec(ctx)
	return err
}

func (r *departmentRepo) UpdateDepartmentStatus(ctx context.Context, id, status string) error {
	_, err := r.data.db.Department.UpdateOneID(id).
		SetStatus(department.Status(status)).
		SetUpdatedAt(time.Now()).
		Save(ctx)
	return err
}

func (r *departmentRepo) GetDepartment(ctx context.Context, id string) (*biz.Department, error) {
	// 参数校验
	if id == "" {
		return nil, errors.BadRequest("INVALID_ID", "id不能为空")
	}

	v, err := r.data.db.Department.
		Query().
		Where(department.IDEQ(id)).
		Only(ctx)

	if err != nil {
		if ent.IsNotFound(err) {
			return nil, errors.NotFound("DEPARTMENT_NOT_FOUND", "部门不存在")
		}
		return nil, err
	}

	return r.toBiz(v), nil
}

func (r *departmentRepo) toBiz(v *ent.Department) *biz.Department {
	return &biz.Department{
		ID:          v.ID,
		ParentID:    v.ParentID,
		Name:        v.Name,
		Code:        v.Code,
		Weight:      v.Weight,
		Status:      string(v.Status),
		LeaderName:  v.LeaderName,
		LeaderPhone: v.LeaderPhone,
		LeaderEmail: v.LeaderEmail,
		CreatedAt:   v.CreatedAt,
		UpdatedAt:   v.UpdatedAt,
	}
}
