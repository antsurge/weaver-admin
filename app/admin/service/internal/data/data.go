package data

import (
	"context"
	"fmt"
	"time"

	"entgo.io/ent/dialect/sql/schema"
	"github.com/antsurge/weaver-admin/app/admin/service/internal/conf"
	"github.com/antsurge/weaver-admin/app/admin/service/internal/data/ent"
	_ "github.com/antsurge/weaver-admin/app/admin/service/internal/data/ent/runtime"
	"github.com/antsurge/weaver-admin/app/admin/service/internal/data/mq/rabbitmq"
	"github.com/go-kratos/kratos/v2/log"
	_ "github.com/go-sql-driver/mysql"
	"github.com/google/wire"
	"github.com/redis/go-redis/v9"
)

// ProviderSet is data providers.
var ProviderSet = wire.NewSet(
	NewData,
	NewEntClient,
	NewRedisClient,

	NewAdminRepo,
	NewCaptchaRepo,
	NewTokenRepo,

	NewMenuRepo,
	NewRoleRepo,
	NewAdminRoleRepo,
	NewRoleMenuRepo,

	NewDepartmentRepo,
	NewPositionRepo,

	NewDictTypeRepo,
	NewDictDataRepo,

	rabbitmq.ProviderSet,
)

// NewEntClient 初始化 Ent 客户端
func NewEntClient(c *conf.Data, logger log.Logger) *ent.Client {
	l := log.NewHelper(logger)

	client, err := ent.Open(c.Database.Driver, c.Database.Source)
	if err != nil {
		l.Fatalf("failed opening connection to database: %v", err)
	}

	// 开发环境开启 SQL 日志
	client = client.Debug()

	// 开发环境下开启自动迁移 (Auto Migration)
	// 生产环境建议通过 deploy/sql 下的脚本手动管理
	if err := client.Schema.Create(
		context.Background(),
		schema.WithForeignKeys(true),
		schema.WithDropColumn(true),
		schema.WithDropIndex(true),
	); err != nil {
		l.Fatalf("failed creating schema resources: %v", err)
	}

	return client
}

// NewRedisClient 初始化 Redis 客户端
func NewRedisClient(c *conf.Data, logger log.Logger) *redis.Client {
	l := log.NewHelper(logger)

	// 解析超时时间
	readTimeout := 2 * time.Second
	writeTimeout := 2 * time.Second
	// 解析网络类型和地址
	network := "tcp"
	addr := "127.0.0.1:6379"
	// 密码
	password := ""
	if c.Redis != nil {
		if c.Redis.ReadTimeout != nil && c.Redis.ReadTimeout.AsDuration() > 0 {
			readTimeout = c.Redis.ReadTimeout.AsDuration()
		}
		if c.Redis.WriteTimeout != nil && c.Redis.WriteTimeout.AsDuration() > 0 {
			writeTimeout = c.Redis.WriteTimeout.AsDuration()
		}
		if c.Redis.Network != "" {
			network = c.Redis.Network
		}
		if c.Redis.Addr != "" {
			addr = c.Redis.Addr
		}
		if c.Redis.Password != "" {
			password = c.Redis.Password
		}
	}

	rdb := redis.NewClient(&redis.Options{
		Network:      network,
		Addr:         addr,
		ReadTimeout:  readTimeout,
		WriteTimeout: writeTimeout,
		DialTimeout:  2 * time.Second, // 新增
		Password:     password,
	})

	// 测试连接
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := rdb.Ping(ctx).Err(); err != nil {
		l.Fatalf("failed connecting to redis: %v", err)
	}

	l.Infof("redis connected successfully: %s", addr)
	return rdb
}

// Data .
type Data struct {
	db    *ent.Client
	redis *redis.Client
}

// NewData .
func NewData(
	c *conf.Data,
	logger log.Logger,
	entClient *ent.Client,
	redisClient *redis.Client,
	// rabbitmq *rabbitmq.RabbitMQ,
) (*Data, func(), error) {
	l := log.NewHelper(logger)
	d := &Data{
		db:    entClient,
		redis: redisClient,
		//rabbitmq: rabbitmq,
	}
	return d, func() {
		l.Info("message", "closing the data resources")
		if err := d.db.Close(); err != nil {
			l.Error(err)
		}
		if err := d.redis.Close(); err != nil {
			l.Error(err)
		}
	}, nil
}

// GetRedis 获取 Redis 客户端
func (d *Data) GetRedis() *redis.Client {
	return d.redis
}

// InTx 事务包装函数
func (d *Data) InTx(ctx context.Context, fn func(ctx context.Context) error) error {
	// 1. 开启事务
	tx, err := d.db.Tx(ctx)
	if err != nil {
		return err
	}

	// 2. 执行业务逻辑
	// 注意：这里需要考虑逻辑执行中的 panic 恢复
	defer func() {
		if v := recover(); v != nil {
			tx.Rollback()
			panic(v)
		}
	}()

	if err := fn(ctx); err != nil {
		// 3. 发生错误时回滚
		if rerr := tx.Rollback(); rerr != nil {
			return fmt.Errorf("%w: rolling back transaction: %v", err, rerr)
		}
		return err
	}

	// 4. 成功后提交
	return tx.Commit()
}
