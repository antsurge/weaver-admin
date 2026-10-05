package biz

import (
	"context"
	"time"

	"github.com/antsurge/weaver-admin/pkg/enthelper"
	"github.com/antsurge/weaver-admin/pkg/utils/uuid"
	"github.com/go-kratos/kratos/v2/log"
)

type SysConfig struct {
	ID         string     `json:"id"`
	Name       string     `json:"name"`
	Key        string     `json:"key"`
	Value      string     `json:"value"`
	ConfigType string     `json:"configType"`
	Status     string     `json:"status"`
	Remark     string     `json:"remark"`
	Group      string     `json:"group,omitempty"`
	CreatedAt  time.Time  `json:"createdAt"`
	UpdatedAt  time.Time  `json:"updatedAt"`
	DeletedAt  *time.Time `json:"deletedAt,omitempty"`
}

type ListConfigRequest struct {
	enthelper.PaginationParams
	Name   string `form:"name" query:"name"`
	Key    string `form:"key" query:"key"`
	Status string `form:"status" query:"status"`
	Group  string `form:"group" query:"group"`
}

type ListConfigOption struct {
	enthelper.QueryOption
}

type ListConfigResponse struct {
	Data  []*SysConfig
	Total int
}

type ConfigRepo interface {
	ListConfig(context.Context, *ListConfigRequest, ...*ListConfigOption) (*ListConfigResponse, error)
	GetConfigByKey(context.Context, string) (*SysConfig, error)
	CreateConfig(context.Context, *SysConfig) error
	UpdateConfig(context.Context, *SysConfig) error
	DeleteConfig(context.Context, []string) error
	UpdateConfigStatus(context.Context, string, string) error
}

type ConfigUsecase struct {
	repo ConfigRepo
	log  *log.Helper
}

func NewConfigUsecase(repo ConfigRepo, logger log.Logger) *ConfigUsecase {
	return &ConfigUsecase{
		repo: repo,
		log:  log.NewHelper(logger),
	}
}

func (uc *ConfigUsecase) List(ctx context.Context, req *ListConfigRequest) (*ListConfigResponse, error) {
	return uc.repo.ListConfig(ctx, req)
}

func (uc *ConfigUsecase) CreateConfig(ctx context.Context, req *SysConfig) (*SysConfig, error) {
	config := req
	config.ID = uuid.GenerateXID()
	config.CreatedAt = time.Now()
	config.UpdatedAt = time.Now()
	if config.Status == "" {
		config.Status = "enabled"
	}
	if config.ConfigType == "" {
		config.ConfigType = "N"
	}

	err := uc.repo.CreateConfig(ctx, config)
	return config, err
}

func (uc *ConfigUsecase) UpdateConfig(ctx context.Context, req *SysConfig) (*SysConfig, error) {
	config := req
	config.UpdatedAt = time.Now()

	err := uc.repo.UpdateConfig(ctx, config)
	return config, err
}

func (uc *ConfigUsecase) DeleteConfig(ctx context.Context, ids []string) error {
	return uc.repo.DeleteConfig(ctx, ids)
}

func (uc *ConfigUsecase) UpdateConfigStatus(ctx context.Context, id, status string) error {
	return uc.repo.UpdateConfigStatus(ctx, id, status)
}

// 按键名查询参数（供业务页面消费，禁用或不存在返回 nil）
func (uc *ConfigUsecase) GetConfigByKey(ctx context.Context, key string) (*SysConfig, error) {
	return uc.repo.GetConfigByKey(ctx, key)
}
