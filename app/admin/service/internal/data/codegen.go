package data

import (
	"context"
	"errors"
	"regexp"
	"strings"
	"time"

	"github.com/antsurge/weaver-admin/app/admin/service/internal/biz"
	"github.com/antsurge/weaver-admin/app/admin/service/internal/data/ent"
	"github.com/antsurge/weaver-admin/app/admin/service/internal/data/ent/gentable"
	"github.com/antsurge/weaver-admin/pkg/enthelper"
	"github.com/go-kratos/kratos/v2/log"
)

// validTableName 校验表名仅包含字母、数字、下划线，防止 SQL 注入
var validTableName = regexp.MustCompile(`^[a-zA-Z0-9_]+$`)

var _ biz.GenTableRepo = (*genTableRepo)(nil)
var _ biz.DbMetadataRepo = (*dbMetadataRepo)(nil)

type genTableRepo struct {
	data *Data
	log  *log.Helper
}

func NewGenTableRepo(data *Data, logger log.Logger) biz.GenTableRepo {
	return &genTableRepo{
		data: data,
		log:  log.NewHelper(logger),
	}
}

func (r *genTableRepo) ListGenTable(ctx context.Context, params *biz.ListGenTableRequest, opts ...*biz.ListGenTableOption) (*biz.ListGenTableResponse, error) {
	query := r.data.db.GenTable.Query().
		Order(ent.Desc(gentable.FieldCreatedAt))

	opt := &biz.ListGenTableOption{}
	if len(opts) > 0 {
		opt = opts[0]
	}

	if opt.OnlyDeleted {
		query = query.Where(gentable.DeletedAtNotNil())
	} else if !opt.IncludeDeleted {
		query = query.Where(gentable.DeletedAtIsNil())
	}

	if v := params.TableName; len(v) > 0 {
		query = query.Where(gentable.TableNameContains(v))
	}
	if v := params.BizName; len(v) > 0 {
		query = query.Where(gentable.BizNameContains(v))
	}
	if v := params.Status; len(v) > 0 {
		query = query.Where(gentable.StatusEQ(gentable.Status(v)))
	}

	res, err := enthelper.Pagination[*ent.GenTable, *ent.GenTableQuery](ctx, query, params.PaginationParams)
	if err != nil {
		return nil, err
	}

	data := make([]*biz.GenTable, 0, res.Total)
	for _, v := range res.Data {
		data = append(data, &biz.GenTable{
			ID:           v.ID,
			TableName:    v.TableName,
			TableComment: v.TableComment,
			ModuleName:   v.ModuleName,
			BizName:      v.BizName,
			FieldsJSON:   v.FieldsJSON,
			GenType:      v.GenType,
			MenuEnabled:  v.MenuEnabled,
			MenuModule:   v.MenuModule,
			Buttons:      v.ButtonsJSON,
			Status:       string(v.Status),
			CreatedAt:    v.CreatedAt,
			UpdatedAt:    v.UpdatedAt,
			DeletedAt:    v.DeletedAt,
		})
	}

	return &biz.ListGenTableResponse{
		Data:  data,
		Total: res.Total,
	}, nil
}

func (r *genTableRepo) GetGenTable(ctx context.Context, id string) (*biz.GenTable, error) {
	v, err := r.data.db.GenTable.Query().
		Where(
			gentable.IDEQ(id),
			gentable.DeletedAtIsNil(),
		).
		Only(ctx)
	if err != nil {
		if ent.IsNotFound(err) {
			return nil, nil
		}
		return nil, err
	}

	return &biz.GenTable{
		ID:           v.ID,
		TableName:    v.TableName,
		TableComment: v.TableComment,
		ModuleName:   v.ModuleName,
		BizName:      v.BizName,
		FieldsJSON:   v.FieldsJSON,
		GenType:      v.GenType,
		MenuEnabled:  v.MenuEnabled,
		MenuModule:   v.MenuModule,
		Buttons:      v.ButtonsJSON,
		Status:       string(v.Status),
		CreatedAt:    v.CreatedAt,
		UpdatedAt:    v.UpdatedAt,
		DeletedAt:    v.DeletedAt,
	}, nil
}

func (r *genTableRepo) CreateGenTable(ctx context.Context, d *biz.GenTable) error {
	_, err := r.data.db.GenTable.Create().
		SetID(d.ID).
		SetTableName(d.TableName).
		SetNillableTableComment(&d.TableComment).
		SetModuleName(d.ModuleName).
		SetBizName(d.BizName).
		SetFieldsJSON(d.FieldsJSON).
		SetNillableGenType(&d.GenType).
		SetMenuEnabled(d.MenuEnabled).
		SetNillableMenuModule(&d.MenuModule).
		SetButtonsJSON(d.Buttons).
		SetStatus(gentable.Status(d.Status)).
		SetCreatedAt(d.CreatedAt).
		SetUpdatedAt(d.UpdatedAt).
		Save(ctx)
	return err
}

func (r *genTableRepo) UpdateGenTable(ctx context.Context, d *biz.GenTable) error {
	_, err := r.data.db.GenTable.UpdateOneID(d.ID).
		SetTableName(d.TableName).
		SetNillableTableComment(&d.TableComment).
		SetModuleName(d.ModuleName).
		SetBizName(d.BizName).
		SetFieldsJSON(d.FieldsJSON).
		SetNillableGenType(&d.GenType).
		SetMenuEnabled(d.MenuEnabled).
		SetNillableMenuModule(&d.MenuModule).
		SetButtonsJSON(d.Buttons).
		SetStatus(gentable.Status(d.Status)).
		SetUpdatedAt(d.UpdatedAt).
		Save(ctx)
	return err
}

func (r *genTableRepo) DeleteGenTable(ctx context.Context, ids []string) error {
	if len(ids) == 0 {
		return nil
	}

	now := time.Now()
	return r.data.db.GenTable.
		Update().
		Where(gentable.IDIn(ids...)).
		SetDeletedAt(now).
		Exec(ctx)
}

// GetGenTables 批量查询生成配置（仅未软删数据）
func (r *genTableRepo) GetGenTables(ctx context.Context, ids []string) ([]*biz.GenTable, error) {
	if len(ids) == 0 {
		return nil, nil
	}

	list, err := r.data.db.GenTable.Query().
		Where(
			gentable.IDIn(ids...),
			gentable.DeletedAtIsNil(),
		).
		All(ctx)
	if err != nil {
		return nil, err
	}

	res := make([]*biz.GenTable, 0, len(list))
	for _, v := range list {
		res = append(res, &biz.GenTable{
			ID:           v.ID,
			TableName:    v.TableName,
			TableComment: v.TableComment,
			ModuleName:   v.ModuleName,
			BizName:      v.BizName,
			FieldsJSON:   v.FieldsJSON,
			GenType:      v.GenType,
			MenuEnabled:  v.MenuEnabled,
			MenuModule:   v.MenuModule,
			Buttons:      v.ButtonsJSON,
			Status:       string(v.Status),
			CreatedAt:    v.CreatedAt,
			UpdatedAt:    v.UpdatedAt,
			DeletedAt:    v.DeletedAt,
		})
	}
	return res, nil
}

// DeleteGenTablePermanent 物理删除生成配置（彻底删除时使用）
func (r *genTableRepo) DeleteGenTablePermanent(ctx context.Context, ids []string) error {
	if len(ids) == 0 {
		return nil
	}
	if _, err := r.data.db.GenTable.Delete().Where(gentable.IDIn(ids...)).Exec(ctx); err != nil {
		return err
	}
	return nil
}

// DropTables 删除真实数据表（DROP TABLE IF EXISTS），表名已做白名单校验
func (r *dbMetadataRepo) DropTables(ctx context.Context, tableNames []string) error {
	sqlDB := r.data.sqlDB
	if sqlDB == nil || len(tableNames) == 0 {
		return nil
	}

	for _, name := range tableNames {
		name = strings.TrimSpace(name)
		if name == "" {
			continue
		}
		if !validTableName.MatchString(name) {
			return errors.New("非法的数据表名: " + name)
		}
		if _, err := sqlDB.ExecContext(ctx, "DROP TABLE IF EXISTS `"+name+"`"); err != nil {
			return err
		}
		r.log.Infof("codegen drop table: %s", name)
	}
	return nil
}

// ---------- 数据库元数据读取（information_schema） ----------

type dbMetadataRepo struct {
	data *Data
	log  *log.Helper
}

func NewDbMetadataRepo(data *Data, logger log.Logger) biz.DbMetadataRepo {
	return &dbMetadataRepo{
		data: data,
		log:  log.NewHelper(logger),
	}
}

// ListDbTables 读取当前业务数据库全部表（排除系统库）
func (r *dbMetadataRepo) ListDbTables(ctx context.Context, keyword string) ([]*biz.DbTable, error) {
	sqlDB := r.data.sqlDB
	if sqlDB == nil {
		return nil, nil
	}

	// 取当前数据库名
	var dbName string
	if err := sqlDB.QueryRowContext(ctx, "SELECT DATABASE()").Scan(&dbName); err != nil {
		return nil, err
	}
	if dbName == "" {
		return nil, nil
	}

	query := `SELECT TABLE_NAME, IFNULL(TABLE_COMMENT, '') FROM information_schema.TABLES
		WHERE TABLE_SCHEMA = ? AND TABLE_TYPE = 'BASE TABLE' AND TABLE_NAME <> 'ent_schema_migrations'`
	args := []any{dbName}
	if keyword != "" {
		query += " AND TABLE_NAME LIKE ?"
		args = append(args, "%"+keyword+"%")
	}
	query += " ORDER BY TABLE_NAME"

	rows, err := sqlDB.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	items := make([]*biz.DbTable, 0, 16)
	for rows.Next() {
		var t biz.DbTable
		if err := rows.Scan(&t.TableName, &t.TableComment); err != nil {
			return nil, err
		}
		items = append(items, &t)
	}
	return items, rows.Err()
}

// ListDbColumns 读取指定表字段元数据
func (r *dbMetadataRepo) ListDbColumns(ctx context.Context, tableName string) ([]*biz.DbColumn, error) {
	sqlDB := r.data.sqlDB
	if sqlDB == nil {
		return nil, nil
	}

	var dbName string
	if err := sqlDB.QueryRowContext(ctx, "SELECT DATABASE()").Scan(&dbName); err != nil {
		return nil, err
	}
	if dbName == "" {
		return nil, nil
	}

	query := `SELECT COLUMN_NAME, IFNULL(COLUMN_COMMENT, ''), DATA_TYPE,
		IF(COLUMN_KEY = 'PRI', 1, 0), IF(IS_NULLABLE = 'YES', 1, 0), IF(EXTRA LIKE '%auto_increment%', 1, 0)
		FROM information_schema.COLUMNS
		WHERE TABLE_SCHEMA = ? AND TABLE_NAME = ?
		ORDER BY ORDINAL_POSITION`

	rows, err := sqlDB.QueryContext(ctx, query, dbName, tableName)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	items := make([]*biz.DbColumn, 0, 16)
	for rows.Next() {
		var c biz.DbColumn
		var pk, nullable, auto int
		if err := rows.Scan(&c.ColumnName, &c.ColumnComment, &c.DataType, &pk, &nullable, &auto); err != nil {
			return nil, err
		}
		c.IsPrimaryKey = pk == 1
		c.IsNullable = nullable == 1
		c.AutoIncrement = auto == 1
		// 统一小写，便于模板映射
		c.DataType = strings.ToLower(c.DataType)
		items = append(items, &c)
	}
	return items, rows.Err()
}
