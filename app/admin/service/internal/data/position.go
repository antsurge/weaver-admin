package data

import (
	"context"
	"time"

	"github.com/go-kratos/kratos/v2/log"
	"github.com/hypercoze/kratos-admin/app/admin/service/internal/biz"
	"github.com/hypercoze/kratos-admin/app/admin/service/internal/data/ent"
	"github.com/hypercoze/kratos-admin/app/admin/service/internal/data/ent/position"
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

func (r *positionRepo) CreatePosition(ctx context.Context, req *biz.Position) error {
	_, err := r.data.db.Position.Create().
		SetID(req.ID).
		SetName(req.Name).
		SetCode(req.Code).
		SetWeight(req.Weight).
		SetStatus(position.Status(req.Status)).
		SetDescription(req.Description).
		SetCreatedAt(req.CreatedAt).
		SetUpdatedAt(req.UpdatedAt).
		Save(ctx)
	return err
}

func (r *positionRepo) UpdatePosition(ctx context.Context, req *biz.Position) error {
	_, err := r.data.db.Position.
		UpdateOneID(req.ID).
		SetName(req.Name).
		SetCode(req.Code).
		SetWeight(req.Weight).
		SetStatus(position.Status(req.Status)).
		SetDescription(req.Description).
		SetUpdatedAt(req.UpdatedAt).
		Save(ctx)
	return err
}

func (r *positionRepo) DeletePosition(ctx context.Context, ids []string) error {
	_, err := r.data.db.Position.
		Delete().
		Where(position.IDIn(ids...)).
		Exec(ctx)

	return err
}

func (repo *positionRepo) GetPosition(ctx context.Context, id string) (*biz.Position, error) {
	return nil, nil
}

func (repo *positionRepo) ListPosition(ctx context.Context, req *biz.ListPositionRequest) (*biz.ListPositionResponse, error) {
	query := repo.data.db.Position.Query()

	// 查询总数
	total, err := query.Count(ctx)
	if err != nil {
		return nil, err
	}

	// 分页
	list, err := query.
		Order(ent.Desc(position.FieldCreatedAt)).
		Offset((req.Page - 1) * req.PageSize).
		Limit(req.PageSize).
		All(ctx)
	if err != nil {
		return nil, err
	}

	// 转换为biz结构
	data := make([]*biz.Position, 0, len(list))
	for _, v := range list {
		data = append(data, &biz.Position{
			ID:          v.ID,
			Name:        v.Name,
			Code:        v.Code,
			Weight:      v.Weight,
			Status:      string(v.Status),
			Description: v.Description,
			CreatedAt:   v.CreatedAt,
			UpdatedAt:   v.UpdatedAt,
		})
	}

	return &biz.ListPositionResponse{
		Data:  data,
		Total: total,
	}, nil

	return nil, nil
}

func (r *positionRepo) UpdatePositionStatus(ctx context.Context, id, status string) error {
	_, err := r.data.db.Position.UpdateOneID(id).
		SetStatus(position.Status(status)).
		SetUpdatedAt(time.Now()).
		Save(ctx)
	return err
}
