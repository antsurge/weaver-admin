package data

import (
	"context"
	"fmt"
	"time"

	"entgo.io/ent/dialect/sql/schema"
	"github.com/go-kratos/kratos/v2/log"
	_ "github.com/go-sql-driver/mysql"
	"github.com/google/wire"
	"github.com/hypercoze/kratos-admin/app/admin/service/internal/conf"
	"github.com/hypercoze/kratos-admin/app/admin/service/internal/data/ent"
	_ "github.com/hypercoze/kratos-admin/app/admin/service/internal/data/ent/runtime"
	"github.com/hypercoze/kratos-admin/app/admin/service/internal/data/mq/rabbitmq"
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

	// 开发环境下开启自动迁移 (Auto Migration)
	// 生产环境建议通过 deploy/sql 下的脚本手动管理
	if err := client.Schema.Create(
		context.Background(),
		//schema.WithAtlas(true),       // 使用 Atlas 引擎（功能更强）
		schema.WithForeignKeys(true), // 生成外键
		schema.WithDropColumn(true),  // 允许删除字段 (慎用!)
		schema.WithDropIndex(true),   // 允许删除索引
	); err != nil {
		l.Fatalf("failed creating schema resources: %v", err)
	}

	return client
}

// NewRedisClient 初始化 Redis 客户端
func NewRedisClient(c *conf.Data, logger log.Logger) *redis.Client {
	l := log.NewHelper(logger)

	// 解析超时时间
	readTimeout := time.Duration(0)
	writeTimeout := time.Duration(0)
	if c.Redis != nil {
		if c.Redis.ReadTimeout != nil {
			readTimeout = c.Redis.ReadTimeout.AsDuration()
		}
		if c.Redis.WriteTimeout != nil {
			writeTimeout = c.Redis.WriteTimeout.AsDuration()
		}
	}

	// 设置默认值
	if readTimeout == 0 {
		readTimeout = 2 * time.Second
	}
	if writeTimeout == 0 {
		writeTimeout = 2 * time.Second
	}

	// 解析网络类型和地址
	network := "tcp"
	addr := "127.0.0.1:6379"
	if c.Redis != nil {
		if c.Redis.Network != "" {
			network = c.Redis.Network
		}
		if c.Redis.Addr != "" {
			addr = c.Redis.Addr
		}
	}

	rdb := redis.NewClient(&redis.Options{
		Network:      network,
		Addr:         addr,
		ReadTimeout:  readTimeout,
		WriteTimeout: writeTimeout,
		DialTimeout:  2 * time.Second, // 新增
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
