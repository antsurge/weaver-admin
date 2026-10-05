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
		Order(ent.Desc(department.FieldWeight)).
		// 排除软删除数据
		Where(department.DeletedAtIsNil())

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
			Type:        string(v.Type),
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
		SetType(department.Type(req.Type)).
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
	// 类型：仅当请求携带非空值时更新，空值跳过（避免置空触发 Ent 枚举校验失败）
	var typeVal *department.Type
	if req.Type != "" {
		t := department.Type(req.Type)
		typeVal = &t
	}

	_, err := r.data.db.Department.UpdateOneID(req.ID).
		SetParentID(req.ParentID).
		SetName(req.Name).
		SetCode(req.Code).
		SetNillableType(typeVal).
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

	// 已访问集合防环：若部门数据存在循环 parent 引用，直接递归会无限递归导致栈溢出
	visited := make(map[string]bool, len(ids))

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

		// 收集子节点 ID（去除已访问节点，防止环导致无限递归）
		childIDs := make([]string, 0, len(children))
		for _, c := range children {
			if visited[c.ID] {
				continue
			}
			visited[c.ID] = true
			childIDs = append(childIDs, c.ID)
		}
		allIDs = append(allIDs, childIDs...)

		// 递归查子节点
		return collect(childIDs)
	}

	// 初始化 allIDs 与 visited
	allIDs = append(allIDs, ids...)
	for _, id := range ids {
		visited[id] = true
	}
	if err := collect(ids); err != nil {
		return err
	}

	// 软删除所有节点（schema 定义了 deleted_at，与 admin/role 策略保持一致）
	// 注意：department.code 唯一索引下，软删除后重建相同 code 的部门会冲突（已知权衡）
	err := r.data.db.Department.
		Update().
		Where(department.IDIn(allIDs...)).
		SetDeletedAt(time.Now()).
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

func (r *departmentRepo) IsDepartmentCodeExists(ctx context.Context, code, id string) (bool, error) {
	query := r.data.db.Department.
		Query().
		Where(
			department.CodeEQ(code),
			department.DeletedAtIsNil(),
		)

	// 如果是编辑，排除当前记录
	if id != "" {
		query = query.Where(department.IDNEQ(id))
	}

	// 判断是否存在
	return query.Exist(ctx)
}

func (r *departmentRepo) GetDepartment(ctx context.Context, id string) (*biz.Department, error) {
	// 参数校验
	if id == "" {
		return nil, errors.BadRequest("INVALID_ID", "id不能为空")
	}

	v, err := r.data.db.Department.
		Query().
		Where(
			department.IDEQ(id),
			department.DeletedAtIsNil(),
		).
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
		Type:        string(v.Type),
		Weight:      v.Weight,
		Status:      string(v.Status),
		LeaderName:  v.LeaderName,
		LeaderPhone: v.LeaderPhone,
		LeaderEmail: v.LeaderEmail,
		CreatedAt:   v.CreatedAt,
		UpdatedAt:   v.UpdatedAt,
	}
}
