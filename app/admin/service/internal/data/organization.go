package data

import (
	"context"
	"time"

	"github.com/go-kratos/kratos/v2/log"
	"github.com/hypercoze/kratos-admin/app/admin/service/internal/biz"
	"github.com/hypercoze/kratos-admin/app/admin/service/internal/data/ent"
	"github.com/hypercoze/kratos-admin/app/admin/service/internal/data/ent/organization"
	"github.com/hypercoze/kratos-admin/app/admin/service/internal/data/ent/permission"
)

type organizationRepo struct {
	data *Data
	log  *log.Helper
}

func NewOrganizationRepo(data *Data, logger log.Logger) biz.OrganizationRepo {
	return &organizationRepo{
		data: data,
		log:  log.NewHelper(logger),
	}
}

func (r *organizationRepo) ListOrganization(ctx context.Context, params *biz.OrganizationListResult) ([]*biz.Organization, error) {
	query := r.data.db.Organization.Query().
		Order(ent.Desc(organization.FieldWeight))

	// 名称
	if v := params.Name; len(v) > 0 {
		query = query.Where(organization.NameContains(v))
	}

	// code
	if v := params.Code; len(v) > 0 {
		query = query.Where(organization.CodeContains(v))
	}

	// 状态
	if v := params.Status; len(v) > 0 {
		query = query.Where(organization.StatusEQ(organization.Status(v)))
	}

	list, err := query.All(ctx)

	res := make([]*biz.Organization, 0, len(list))
	if err != nil {
		return nil, err
	}

	for _, v := range list {
		res = append(res, &biz.Organization{
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

func (r *organizationRepo) CreateOrganization(ctx context.Context, req *biz.Organization) error {
	_, err := r.data.db.Organization.Create().
		SetID(req.ID).
		SetParentID(req.ParentID).
		SetName(req.Name).
		SetCode(req.Code).
		SetLeaderName(req.LeaderName).
		SetLeaderPhone(req.LeaderPhone).
		SetLeaderEmail(req.LeaderEmail).
		SetWeight(req.Weight).
		SetStatus(organization.Status(req.Status)).
		SetCreatedAt(req.CreatedAt).
		SetUpdatedAt(req.UpdatedAt).
		Save(ctx)
	return err
}

func (r *organizationRepo) UpdateOrganization(ctx context.Context, req *biz.Organization) error {
	_, err := r.data.db.Organization.UpdateOneID(req.ID).
		SetParentID(req.ParentID).
		SetName(req.Name).
		SetCode(req.Code).
		SetLeaderName(req.LeaderName).
		SetLeaderPhone(req.LeaderPhone).
		SetLeaderEmail(req.LeaderEmail).
		SetWeight(req.Weight).
		SetStatus(organization.Status(req.Status)).
		SetUpdatedAt(req.UpdatedAt).
		Save(ctx)
	return err
}

func (r *organizationRepo) DeleteOrganization(ctx context.Context, ids []string) error {
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

		children, err := r.data.db.Permission.
			Query().
			Where(permission.ParentIDIn(ids...)).
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
	_, err := r.data.db.Organization.
		Delete().
		Where(organization.IDIn(allIDs...)).
		Exec(ctx)
	return err
}

func (r *organizationRepo) UpdateOrganizationStatus(ctx context.Context, id, status string) error {
	_, err := r.data.db.Organization.UpdateOneID(id).
		SetStatus(organization.Status(status)).
		SetUpdatedAt(time.Now()).
		Save(ctx)
	return err
}
