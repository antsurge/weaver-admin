package data

import (
	"context"
	"database/sql"
	"fmt"
	"os"
	"path/filepath"
	"time"

	entsql "entgo.io/ent/dialect/sql"
	"entgo.io/ent/dialect/sql/schema"
	"github.com/antsurge/weaver-admin/app/admin/service/internal/conf"
	"github.com/antsurge/weaver-admin/app/admin/service/internal/data/ent"
	_ "github.com/antsurge/weaver-admin/app/admin/service/internal/data/ent/runtime"
	enttenant "github.com/antsurge/weaver-admin/app/admin/service/internal/data/ent/tenant"
	"github.com/antsurge/weaver-admin/app/admin/service/internal/data/openapi_scanner"
	"github.com/antsurge/weaver-admin/pkg/tenant"
	"github.com/go-kratos/kratos/v2/log"
	_ "github.com/go-sql-driver/mysql"
	"github.com/google/wire"
	"github.com/redis/go-redis/v9"
)

// ProviderSet is data providers.
var ProviderSet = wire.NewSet(
	NewData,
	NewEntDriver,
	NewEntClient,
	NewSQLDB,
	NewRedisClient,
	NewOpenAPIScanner,
	NewObjectStorage,
	NewObjectStorageProvider,
	NewMessageBroker,

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

	NewApiInterfaceRepo,

	NewNotificationRepo,
	NewLogRepo,

	NewAuthzRepo,
	NewAuthzCache,

	NewConfigRepo,

	NewSecurityRepo,

	NewDataPermissionRepo,

	NewCronJobRepo,
	NewRedisJobLocker,

	NewMonitorRepo,

	NewTenantRepo,

	NewFormSchemaRepo,
	NewFormSubmissionRepo,
	NewGenTableRepo,
	NewDbMetadataRepo,
)

/**
 * EntDriver 持有底层 ent 数据库驱动。
 * 作为中间 provider：NewEntClient 与 NewSQLDB 都从它取值，
 * 确保 *ent.Client 与 *sql.DB 共享同一连接池，且连接只在一个位置创建与清理。
 */
type EntDriver struct {
	drv *entsql.Driver
}

/**
 * NewEntDriver 打开数据库连接并持有底层驱动。
 * 返回 (driver, cleanup, error)，cleanup 负责关闭底层连接。
 */
func NewEntDriver(c *conf.Data, logger log.Logger) (*EntDriver, func(), error) {
	l := log.NewHelper(logger)

	drv, err := entsql.Open(c.Database.Driver, c.Database.Source)
	if err != nil {
		return nil, nil, err
	}

	return &EntDriver{drv: drv}, func() {
		l.Info("closing database connection")
		if err := drv.DB().Close(); err != nil {
			l.Error(err)
		}
	}, nil
}

/**
 * NewEntClient 基于共享驱动初始化 Ent 客户端。
 */
func NewEntClient(d *EntDriver, logger log.Logger) (*ent.Client, error) {
	// 开发环境开启 SQL 日志
	client := ent.NewClient(ent.Driver(d.drv)).Debug()

	// 多租户隔离：查询自动追加 tenant_id 过滤；写入自动填充/校验租户
	client.Intercept(TenantQueryInterceptor())
	client.Use(TenantMutationHook())

	// 开发环境下开启自动迁移 (Auto Migration)
	// 生产环境建议通过 deploy/sql 下的脚本手动管理
	if err := client.Schema.Create(
		context.Background(),
		schema.WithForeignKeys(true),
		schema.WithDropColumn(true),
		schema.WithDropIndex(true),
	); err != nil {
		return nil, err
	}

	return client, nil
}

/**
 * NewSQLDB 提供底层 *sql.DB。
 * 供数据库监控读取连接池状态、代码生成器读取表/列元数据。
 * 与 NewEntClient 共享同一个连接池，不会新建连接。
 */
func NewSQLDB(d *EntDriver) *sql.DB {
	return d.drv.DB()
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
	sqlDB *sql.DB
	redis *redis.Client
}

// NewData .
func NewData(
	c *conf.Data,
	logger log.Logger,
	entClient *ent.Client,
	sqlDB *sql.DB,
	redisClient *redis.Client,
) (*Data, func(), error) {
	l := log.NewHelper(logger)
	d := &Data{
		db:    entClient,
		sqlDB: sqlDB,
		redis: redisClient,
	}

	// 幂等写入默认租户种子数据（default），保证 single 模式下 tenant 表
	// 就绪、multi 模式可按 code=default 登录。失败仅记日志，不阻塞启动。
	if err := ensureDefaultTenant(context.Background(), entClient, l); err != nil {
		l.Errorf("seed default tenant failed: %v", err)
	}

	// 注意：菜单种子（ensureDefaultMenus/ensureSecurityMenu/ensureLogMenu/ensureLowcodeMenu）
	// 已移除：它们会在每次启动时按 code 重建菜单，导致用户在菜单管理页删除的
	// "系统管理/安全中心/日志中心/低代码"等目录在服务重启后被重新插入。
	// 如需新环境初始化菜单，请手动执行 docs/ 下的 SQL 脚本。

	// 注意：数据库连接由 NewEntClient 的 cleanup 负责关闭（资源创建处清理），
	// 这里仅关闭 Redis，避免与 ent client 关闭同一连接池造成双重关闭。
	return d, func() {
		l.Info("message", "closing the data resources")
		if err := d.redis.Close(); err != nil {
			l.Error(err)
		}
	}, nil
}

// ensureDefaultTenant 幂等确保默认租户（id=default, code=default）存在。
// tenant 表为平台级表（无租户字段），不受租户隔离钩子约束。
func ensureDefaultTenant(ctx context.Context, client *ent.Client, l *log.Helper) error {
	exists, err := client.Tenant.Query().
		Where(
			enttenant.IDEQ(tenant.DefaultTenantID),
			enttenant.DeletedAtIsNil(),
		).
		Exist(ctx)
	if err != nil {
		return err
	}
	if exists {
		return nil
	}
	now := time.Now()
	_, err = client.Tenant.Create().
		SetID(tenant.DefaultTenantID).
		SetCode(tenant.DefaultTenantID).
		SetName("默认租户").
		SetStatus(enttenant.StatusEnabled).
		SetMaxUsers(100).
		SetMaxRoles(100).
		SetCreatedAt(now).
		SetUpdatedAt(now).
		Save(ctx)
	if err != nil {
		return err
	}
	l.Infof("seeded default tenant (id=%s code=%s)", tenant.DefaultTenantID, tenant.DefaultTenantID)
	return nil
}

// defaultOpenAPIPath 默认 openapi.yaml 位置（项目根相对路径）。
const defaultOpenAPIPath = "api/gen/openapi/openapi.yaml"

// resolveOpenAPIPath 解析 openapi.yaml 路径，按优先级依次尝试。
// 依次尝试：配置指定路径 → cwd 逐级向上 → 可执行文件目录逐级向上 → Docker 镜像路径。
// 返回第一个存在的文件路径；全部失败则返回空字符串。
func resolveOpenAPIPath(cfgPath string) string {
	// 1. 优先使用配置指定的路径
	if cfgPath != "" {
		if _, err := os.Stat(cfgPath); err == nil {
			return cfgPath
		}
	}
	// 2. 从工作目录与可执行文件目录出发，逐级向上查找（最多 6 级）
	starts := []string{}
	if cwd, err := os.Getwd(); err == nil {
		starts = append(starts, cwd)
	}
	if exe, err := os.Executable(); err == nil {
		starts = append(starts, filepath.Dir(exe))
	}
	for _, start := range starts {
		dir := start
		for i := 0; i < 6; i++ {
			candidate := filepath.Join(dir, defaultOpenAPIPath)
			if _, err := os.Stat(candidate); err == nil {
				return candidate
			}
			parent := filepath.Dir(dir)
			if parent == dir {
				break
			}
			dir = parent
		}
	}
	// 3. Docker 镜像路径
	if _, err := os.Stat("/app/" + defaultOpenAPIPath); err == nil {
		return "/app/" + defaultOpenAPIPath
	}
	return ""
}

// NewOpenAPIScanner 创建 openapi 扫描器并立即扫描一次。
// 失败仅记日志不阻塞启动，便于开发期缺 openapi.yaml 的情况。
func NewOpenAPIScanner(c *conf.OpenAPI, logger log.Logger) *openapi_scanner.Service {
	l := log.NewHelper(logger)
	scanner := openapi_scanner.New()

	// 从配置或 fallback 路径解析
	cfgPath := ""
	if c != nil {
		cfgPath = c.Path
	}
	path := resolveOpenAPIPath(cfgPath)
	if path == "" {
		l.Warnf("openapi scanner: no openapi.yaml found (config=%q), api-metadata will be empty", cfgPath)
		return scanner
	}

	if err := scanner.Scan(path); err != nil {
		// 启动阶段失败应当被关注但不应阻塞进程，
		// 这样开发期缺少 openapi.yaml 也能把服务起来。
		l.Errorf("openapi scanner init failed, path=%s err=%v", path, err)
		return scanner
	}
	l.Infof("openapi scanner initialized, path=%s", path)
	return scanner
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
