package data

import (
	"context"
	"time"

	"entgo.io/ent/dialect/sql"
	"github.com/antsurge/weaver-admin/app/admin/service/internal/biz"
	"github.com/antsurge/weaver-admin/app/admin/service/internal/data/ent"
	"github.com/antsurge/weaver-admin/app/admin/service/internal/data/ent/position"
	"github.com/antsurge/weaver-admin/pkg/enthelper"
	"github.com/go-kratos/kratos/v2/errors"
	"github.com/go-kratos/kratos/v2/log"
)

type positionRepo struct {
	data *Data
	log  *log.Helper
}

func NewPositionRepo(data *Data, logger log.Logger) biz.PositionRepo {
	return &positionRepo{
		data: data,
		log:  log.NewHelper(logger),
	}
}

func (r *positionRepo) ListPosition(ctx context.Context, params *biz.ListPositionRequest, opts ...*biz.ListPositionOption) (*biz.ListPositionResponse, error) {
	query := r.data.db.Position.Query().
		Order(ent.Desc(position.FieldCreatedAt))

	// 排序
	sorts := make([]enthelper.Sort, 0)
	sorts = append(sorts, enthelper.Sort{
		Field: "created_at",
		Order: "desc",
	})

	opt := &biz.ListPositionOption{}
	if len(opts) > 0 {
		opt = opts[0]
		sorts = opt.Sorts
	}

	if opt.OnlyDeleted {
		query = query.Where(position.DeletedAtNotNil())
	} else if !opt.IncludeDeleted {
		query = query.Where(position.DeletedAtIsNil())
	}

	// 名称
	if v := params.Name; len(v) > 0 {
		query = query.Where(position.NameContains(v))
	}

	// code
	if v := params.Code; len(v) > 0 {
		query = query.Where(position.CodeContains(v))
	}

	if v := params.Names; len(v) > 0 {
		query = query.Where(position.NameIn(v...))
	}
	if v := params.Codes; len(v) > 0 {
		query = query.Where(position.CodeIn(v...))
	}

	// 状态
	if v := params.Status; len(v) > 0 {
		query = query.Where(position.StatusEQ(position.Status(v)))
	}

	// 排序
	query = enthelper.ApplySorts(query, sorts, PositionSortMap())

	res, err := enthelper.Pagination[
		*ent.Position,
		*ent.PositionQuery,
	](ctx, query, params.PaginationParams)
	if err != nil {
		return nil, err
	}

	// 转换为biz结构
	data := make([]*biz.Position, 0, res.Total)
	for _, v := range res.Data {
		data = append(data, &biz.Position{
			ID:        v.ID,
			Name:      v.Name,
			Code:      v.Code,
			Weight:    v.Weight,
			Status:    string(v.Status),
			Remark:    v.Remark,
			CreatedAt: v.CreatedAt,
			UpdatedAt: v.UpdatedAt,
		})
	}

	return &biz.ListPositionResponse{
		Items: data,
		Total: res.Total,
	}, nil
}

func (r *positionRepo) GetPosition(ctx context.Context, id string) (*biz.Position, error) {
	p, err := r.data.db.Position.Get(ctx, id)
	if err != nil {
		if ent.IsNotFound(err) {
			return nil, errors.NotFound("POSITION_NOT_FOUND", "岗位不存在")
		}
		return nil, err
	}
	position := &biz.Position{
		ID:        p.ID,
		Name:      p.Name,
		Code:      p.Code,
		Weight:    p.Weight,
		Status:    string(p.Status),
		Remark:    p.Remark,
		CreatedAt: p.CreatedAt,
		UpdatedAt: p.UpdatedAt,
	}

	return position, nil
}

func (r *positionRepo) CreatePosition(ctx context.Context, req *biz.Position) error {
	_, err := r.data.db.Position.Create().
		SetID(req.ID).
		SetName(req.Name).
		SetCode(req.Code).
		SetWeight(req.Weight).
		SetStatus(position.Status(req.Status)).
		SetRemark(req.Remark).
		SetCreatedAt(req.CreatedAt).
		SetUpdatedAt(req.UpdatedAt).
		Save(ctx)
	return err
}

func (r *positionRepo) BatchCreatePosition(ctx context.Context, list []*biz.Position) error {
	if len(list) == 0 {
		return nil
	}

	builders := make([]*ent.PositionCreate, 0, len(list))

	for _, item := range list {
		builder := r.data.db.Position.Create().
			SetID(item.ID).
			SetName(item.Name).
			SetCode(item.Code).
			SetWeight(item.Weight).
			SetStatus(position.Status(item.Status)).
			SetRemark(item.Remark).
			SetCreatedAt(item.CreatedAt).
			SetUpdatedAt(item.UpdatedAt)

		builders = append(builders, builder)
	}

	return r.data.db.Position.CreateBulk(builders...).Exec(ctx)
}

func (r *positionRepo) UpdatePosition(ctx context.Context, req *biz.Position) error {
	_, err := r.data.db.Position.
		UpdateOneID(req.ID).
		SetName(req.Name).
		SetCode(req.Code).
		SetWeight(req.Weight).
		SetStatus(position.Status(req.Status)).
		SetRemark(req.Remark).
		SetUpdatedAt(req.UpdatedAt).
		Save(ctx)
	return err
}

func (r *positionRepo) DeletePosition(ctx context.Context, ids []string) error {
	err := r.data.db.Position.
		Update().
		Where(position.IDIn(ids...)).
		SetDeletedAt(time.Now()).
		Exec(ctx)

	return err
}

func (r *positionRepo) UpdatePositionStatus(ctx context.Context, id, status string) error {
	_, err := r.data.db.Position.UpdateOneID(id).
		SetStatus(position.Status(status)).
		SetUpdatedAt(time.Now()).
		Save(ctx)
	return err
}

func (r *positionRepo) IsPositionCodeExists(ctx context.Context, code, id string) (bool, error) {
	query := r.data.db.Position.
		Query().
		Where(position.CodeEQ(code))

	// 如果是编辑，排除当前记录
	if id != "" {
		query = query.Where(position.IDNEQ(id))
	}

	// 判断是否存在
	return query.Exist(ctx)
}

func (r *positionRepo) IsPositionNameExists(ctx context.Context, name, id string) (bool, error) {
	query := r.data.db.Position.
		Query().
		Where(position.Name(name))

	// 如果是编辑，排除当前记录
	if id != "" {
		query = query.Where(position.IDNEQ(id))
	}

	// 判断是否存在
	return query.Exist(ctx)
}

func PositionSortMap() map[string]enthelper.OrderFunc[*ent.PositionQuery] {
	return map[string]enthelper.OrderFunc[*ent.PositionQuery]{

		"created_at": func(q *ent.PositionQuery, order string) *ent.PositionQuery {
			if order == "asc" {
				return q.Order(position.ByCreatedAt())
			}
			return q.Order(position.ByCreatedAt(sql.OrderDesc()))
		},

		"name": func(q *ent.PositionQuery, order string) *ent.PositionQuery {
			if order == "asc" {
				return q.Order(position.ByName())
			}
			return q.Order(position.ByName(sql.OrderDesc()))
		},

		"weight": func(q *ent.PositionQuery, order string) *ent.PositionQuery {
			if order == "asc" {
				return q.Order(position.ByWeight())
			}
			return q.Order(position.ByWeight(sql.OrderDesc()))
		},
	}
}
