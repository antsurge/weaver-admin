package biz

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"text/template"
	"time"
	"unicode"

	"github.com/antsurge/weaver-admin/app/admin/service/internal/conf"
	"github.com/antsurge/weaver-admin/pkg/enthelper"
	"github.com/antsurge/weaver-admin/pkg/utils/uuid"
	"github.com/go-kratos/kratos/v2/log"
)

// GenTable 代码生成器-生成表配置
type GenTable struct {
	ID           string         `json:"id"`
	TableName    string         `json:"tableName"`
	TableComment string         `json:"tableComment"`
	ModuleName   string         `json:"moduleName"`
	BizName      string         `json:"bizName"`
	FieldsJSON   map[string]any `json:"fieldsJson,omitempty"`
	GenType      string         `json:"genType"`
	MenuEnabled  bool           `json:"menuEnabled"`
	MenuModule   string         `json:"menuModule"`
	Buttons      []string       `json:"buttons"`
	Status       string         `json:"status"`
	CreatedAt    time.Time      `json:"createdAt"`
	UpdatedAt    time.Time      `json:"updatedAt"`
	DeletedAt    *time.Time     `json:"deletedAt,omitempty"`
}

// 生成按钮常量（与前端高级配置 buttons 多选对应）
const (
	GenBtnList        = "list"
	GenBtnDetail      = "detail"
	GenBtnCreate      = "create"
	GenBtnEdit        = "edit"
	GenBtnDelete      = "delete"
	GenBtnBatchDelete = "batchDelete"
	GenBtnStatus      = "status"
	GenBtnImport      = "import"
	GenBtnExport      = "export"
)

// defaultButtons 默认生成全部按钮（兼容未配置 buttons 的旧配置）
var defaultButtons = []string{
	GenBtnList, GenBtnDetail, GenBtnCreate, GenBtnEdit,
	GenBtnDelete, GenBtnBatchDelete, GenBtnStatus, GenBtnImport, GenBtnExport,
}

// hasButton 判断按钮列表是否包含指定按钮；未配置时默认全部生成
func hasButton(buttons []string, btn string) bool {
	if len(buttons) == 0 {
		return true
	}
	for _, b := range buttons {
		if b == btn {
			return true
		}
	}
	return false
}

// DbTable 数据库表元数据
type DbTable struct {
	TableName    string `json:"tableName"`
	TableComment string `json:"tableComment"`
}

// DbColumn 数据库字段元数据
type DbColumn struct {
	ColumnName    string `json:"columnName"`
	ColumnComment string `json:"columnComment"`
	DataType      string `json:"dataType"`
	IsPrimaryKey  bool   `json:"isPrimaryKey"`
	IsNullable    bool   `json:"isNullable"`
	AutoIncrement bool   `json:"autoIncrement"`
}

type ListGenTableRequest struct {
	enthelper.PaginationParams
	TableName string `form:"tableName" query:"tableName"`
	BizName   string `form:"bizName" query:"bizName"`
	Status    string `form:"status" query:"status"`
}

type ListGenTableOption struct {
	enthelper.QueryOption
}

type ListGenTableResponse struct {
	Data  []*GenTable
	Total int
}

// 生成的代码文件
type GenCodeFile struct {
	FileName string `json:"fileName"`
	Content  string `json:"content"`
}

type GenTableRepo interface {
	ListGenTable(context.Context, *ListGenTableRequest, ...*ListGenTableOption) (*ListGenTableResponse, error)
	GetGenTable(context.Context, string) (*GenTable, error)
	GetGenTables(context.Context, []string) ([]*GenTable, error)
	CreateGenTable(context.Context, *GenTable) error
	UpdateGenTable(context.Context, *GenTable) error
	DeleteGenTable(context.Context, []string) error
	DeleteGenTablePermanent(context.Context, []string) error
}

type DbMetadataRepo interface {
	ListDbTables(context.Context, string) ([]*DbTable, error)
	ListDbColumns(context.Context, string) ([]*DbColumn, error)
	DropTables(context.Context, []string) error
}

type CodegenUsecase struct {
	repo        GenTableRepo
	dbMeta      DbMetadataRepo
	log         *log.Helper
	gen         *CodeGenerator
	outputRoot  string // 写盘根目录（项目根路径），空则不写盘
	writeToDisk bool   // 是否启用写盘
}

func NewCodegenUsecase(repo GenTableRepo, dbMeta DbMetadataRepo, codegenConf *conf.Codegen, logger log.Logger) *CodegenUsecase {
	outputRoot := ""
	writeToDisk := false
	if codegenConf != nil {
		outputRoot = strings.TrimSpace(codegenConf.GetOutputRoot())
		writeToDisk = codegenConf.GetWriteToDisk()
	}
	return &CodegenUsecase{
		repo:        repo,
		dbMeta:      dbMeta,
		log:         log.NewHelper(logger),
		gen:         NewCodeGenerator(),
		outputRoot:  outputRoot,
		writeToDisk: writeToDisk && outputRoot != "",
	}
}

func (uc *CodegenUsecase) List(ctx context.Context, req *ListGenTableRequest) (*ListGenTableResponse, error) {
	return uc.repo.ListGenTable(ctx, req)
}

func (uc *CodegenUsecase) ListDbTables(ctx context.Context, keyword string) ([]*DbTable, error) {
	return uc.dbMeta.ListDbTables(ctx, keyword)
}

func (uc *CodegenUsecase) ListDbColumns(ctx context.Context, tableName string) ([]*DbColumn, error) {
	return uc.dbMeta.ListDbColumns(ctx, tableName)
}

func (uc *CodegenUsecase) CreateGenTable(ctx context.Context, req *GenTable) (*GenTable, error) {
	table := req
	table.ID = uuid.GenerateXID()
	table.CreatedAt = time.Now()
	table.UpdatedAt = time.Now()
	if table.Status == "" {
		table.Status = "enabled"
	}
	if table.GenType == "" {
		table.GenType = "single"
	}

	err := uc.repo.CreateGenTable(ctx, table)
	return table, err
}

func (uc *CodegenUsecase) UpdateGenTable(ctx context.Context, req *GenTable) (*GenTable, error) {
	table := req
	table.UpdatedAt = time.Now()

	err := uc.repo.UpdateGenTable(ctx, table)
	return table, err
}

func (uc *CodegenUsecase) DeleteGenTable(ctx context.Context, ids []string) error {
	return uc.repo.DeleteGenTable(ctx, ids)
}

// DeleteGeneratedResult 深度删除结果
type DeleteGeneratedResult struct {
	RemovedFiles   int      `json:"removedFiles"`
	DroppedTables  []string `json:"droppedTables"`
	DeletedConfigs []*GenTable
}

// DeleteGenerated 彻底删除生成配置：删除写盘的代码文件、删除真实数据表、物理删除配置记录。
// 返回结果包含待清理的菜单引用，由上层（handler）负责删除已落库的菜单树。
func (uc *CodegenUsecase) DeleteGenerated(ctx context.Context, ids []string) (*DeleteGeneratedResult, error) {
	if len(ids) == 0 {
		return nil, errors.New("请至少选择一条记录")
	}

	gens, err := uc.repo.GetGenTables(ctx, ids)
	if err != nil {
		return nil, err
	}
	if len(gens) == 0 {
		return nil, errors.New("未找到有效的生成配置")
	}

	result := &DeleteGeneratedResult{DeletedConfigs: gens}

	// 1. 删除写盘的代码文件
	for _, g := range gens {
		for _, f := range uc.generatedFilePaths(g) {
			if removed, _ := uc.removeGeneratedFile(f); removed {
				result.RemovedFiles++
			}
		}
		// 1.1 同步注销 wire 依赖（三层 ProviderSet + grpc/http server 注册）
		if err := uc.syncProviderRegistrations(g, false); err != nil {
			uc.log.Errorf("codegen unregister wire providers for %s: %v", g.BizName, err)
		}
		// 1.2 同步移除 i18n 语言包 key（删除场景无列元数据，字段 key 由 gen_table 配置兜底）
		if err := uc.syncI18nKeys(g, nil, false); err != nil {
			uc.log.Errorf("codegen remove i18n keys for %s: %v", g.BizName, err)
		}
	}

	// 2. 删除真实数据表
	tables := make([]string, 0, len(gens))
	for _, g := range gens {
		if strings.TrimSpace(g.TableName) != "" {
			tables = append(tables, strings.TrimSpace(g.TableName))
		}
	}
	if err := uc.dbMeta.DropTables(ctx, tables); err != nil {
		return result, errors.New("数据表删除失败: " + err.Error())
	}
	result.DroppedTables = tables

	// 3. 物理删除配置记录
	if err := uc.repo.DeleteGenTablePermanent(ctx, ids); err != nil {
		return result, err
	}

	return result, nil
}

// generatedFilePaths 根据生成配置推导写盘的文件相对路径列表（与 Generate 输出保持一致）
func (uc *CodegenUsecase) generatedFilePaths(t *GenTable) []string {
	if !uc.writeToDisk || uc.outputRoot == "" {
		return nil
	}
	module := strings.TrimSpace(t.ModuleName)
	if module == "" {
		module = "system"
	}
	snake := ToSnake(strings.TrimSpace(t.BizName))
	if snake == "" {
		snake = ToSnake(strings.TrimSpace(t.TableName))
	}

	files := []string{
		"proto/" + module + "/service/v1/" + snake + ".proto",
		"proto/admin/service/v1/i_" + snake + ".proto",
		"ent/schema/" + snake + ".go",
		"internal/biz/" + snake + ".go",
		"internal/data/" + snake + ".go",
		"internal/service/" + snake + ".go",
		"frontend/api/" + module + "/" + snake + ".ts",
	}

	// 导入/导出场景生成了自定义 handler 与 pb 注册文件
	if hasButton(t.Buttons, GenBtnImport) || hasButton(t.Buttons, GenBtnExport) {
		files = append(files,
			"internal/handler/"+snake+".go",
			"internal/handler/pb/i_"+snake+"_handler_http.pb.go",
		)
	}

	// 前端：form=仅表单页；single=列表页+schema
	if t.GenType == "form" {
		files = append(files, "frontend/views/"+module+"/"+snake+"/index.vue")
	} else {
		files = append(files,
			"frontend/views/"+module+"/"+snake+"/index.vue",
			"frontend/schemas/"+module+"/"+snake+".json",
		)
	}

	// 菜单：打开菜单生成时存在 SQL/JSON 文档
	if t.MenuEnabled {
		files = append(files,
			"docs/menu_"+snake+".sql",
			"docs/menu_"+snake+".json",
		)
	}
	return files
}

// removeGeneratedFile 删除指定相对路径对应磁盘文件；删除后尝试清理空父目录
func (uc *CodegenUsecase) removeGeneratedFile(fileName string) (bool, error) {
	target := resolveProjectPath(uc.outputRoot, fileName)
	if target == "" {
		return false, nil
	}
	if err := os.Remove(target); err != nil {
		if os.IsNotExist(err) {
			return false, nil
		}
		return false, err
	}
	// 清理空目录（仅索引.vue 被删后残留的空 views 目录等）
	if dir := filepath.Dir(target); dir != "" && dir != uc.outputRoot {
		_ = os.Remove(dir)
	}
	uc.log.Infof("codegen removed file: %s", target)
	return true, nil
}

// wireRegistrations 目标宿主文件 -> 需登记的模块构造函数条目列表。
// 每条含插入锚点（Anchor 行的下一行插入 Insert）与精确行（Exact，用于幂等判断/删除）。
type wireRegistration struct {
	Target string // 相对 outputRoot 的宿主文件路径（internal/... 或直接相对）
	Anchor string // 在该行之后插入
	Insert string // 要插入的内容（含行首缩进）
	Exact  string // 精确行（TrimSpace 后用于幂等与删除匹配）
}

// providerRegistrationEntries 返回某业务模块的完整 wire 登记清单：
// 三层 ProviderSet（biz/data/service）+ gRPC/HTTP server 注册（参数与调用）。
func providerRegistrationEntries(bizName string, withHandler bool) []wireRegistration {
	biz := ToCamel(bizName)
	lc := LowerFirst(biz)
	param := lc + "Service *service." + biz + "Service"
	grpcCall := "\tadminV1.Register" + biz + "ServiceServer(srv, " + lc + "Service)"
	httpCall := "\tadminV1.Register" + biz + "ServiceHTTPServer(srv, " + lc + "Service)"

	entries := []wireRegistration{
		// 1. biz 层 ProviderSet
		{
			Target: "internal/biz/biz.go",
			Anchor: "\tNewCodegenUsecase,",
			Insert: "\tNew" + biz + "Usecase,",
			Exact:  "\tNew" + biz + "Usecase,",
		},
		// 2. data 层 ProviderSet
		{
			Target: "internal/data/data.go",
			Anchor: "\tNewDbMetadataRepo,",
			Insert: "\tNew" + biz + "Repo,",
			Exact:  "\tNew" + biz + "Repo,",
		},
		// 3. service 层 ProviderSet
		{
			Target: "internal/service/service.go",
			Anchor: "\tNewLowcodeService,",
			Insert: "\tNew" + biz + "Service,",
			Exact:  "\tNew" + biz + "Service,",
		},
		// 4. gRPC server：构造函数参数 + 注册调用
		{
			Target: "internal/server/grpc.go",
			Anchor: "\tlowcodeService *service.LowcodeService,",
			Insert: "\t" + param + ",",
			Exact:  "\t" + param + ",",
		},
		{
			Target: "internal/server/grpc.go",
			Anchor: "\tadminV1.RegisterLowcodeServer(srv, lowcodeService)",
			Insert: grpcCall,
			Exact:  grpcCall,
		},
		// 5. HTTP server：构造函数参数 + 注册调用
		{
			Target: "internal/server/http.go",
			Anchor: "\tlowcodeService *service.LowcodeService,",
			Insert: "\t" + param + ",",
			Exact:  "\t" + param + ",",
		},
		{
			Target: "internal/server/http.go",
			Anchor: "\tadminV1.RegisterLowcodeHTTPServer(srv, lowcodeService)",
			Insert: httpCall,
			Exact:  httpCall,
		},
	}

	// 6. handler 层：仅当生成导入/导出 handler 时登记（ProviderSet + HTTP 参数与注册）
	if withHandler {
		entries = append(entries,
			wireRegistration{
				Target: "internal/handler/handler.go",
				Anchor: "\tNewProfileHandler,",
				Insert: "\tNew" + biz + "Handler,",
				Exact:  "\tNew" + biz + "Handler,",
			},
			wireRegistration{
				Target: "internal/server/http.go",
				Anchor: "\tprofileHandler *handler.ProfileHandler,",
				Insert: "\t" + lc + "Handler *handler." + biz + "Handler,",
				Exact:  "\t" + lc + "Handler *handler." + biz + "Handler,",
			},
			wireRegistration{
				Target: "internal/server/http.go",
				Anchor: "\tpb.RegisterProfileHandlerServer(srv, profileHandler)",
				Insert: "\tpb.Register" + biz + "HandlerServer(srv, " + lc + "Handler)",
				Exact:  "\tpb.Register" + biz + "HandlerServer(srv, " + lc + "Handler)",
			},
		)
	}
	return entries
}

// syncProviderRegistrations 登记(add=true)/注销(add=false)模块的 wire 依赖。
// 幂等：注册时已存在则跳过；注销时不存在则忽略。
func (uc *CodegenUsecase) syncProviderRegistrations(t *GenTable, add bool) error {
	if !uc.writeToDisk || uc.outputRoot == "" {
		return nil
	}
	bizName := strings.TrimSpace(t.BizName)
	if bizName == "" {
		bizName = ToCamel(t.TableName)
	}
	withHandler := hasButton(t.Buttons, GenBtnImport) || hasButton(t.Buttons, GenBtnExport)
	for _, reg := range providerRegistrationEntries(bizName, withHandler) {
		target := resolveProjectPath(uc.outputRoot, reg.Target)
		if target == "" {
			continue
		}
		if err := uc.applyWireRegistration(target, reg, add); err != nil {
			return errors.New("更新 " + reg.Target + " 失败: " + err.Error())
		}
	}
	return nil
}

// i18nLangFiles 生成模块字段 i18n 语言包文件（相对 outputRoot）。
// 结构：frontend/apps/web-antd/src/locales/langs/{lang}/lowcode.json（与 codegenPathPrefixes 映射一致）。
func (uc *CodegenUsecase) i18nLangFile(lang string) string {
	return filepath.Join(uc.outputRoot,
		"frontend/apps/web-antd/src/locales/langs/", lang, "lowcode.json")
}

// syncI18nKeys 登记(add=true)/注销(add=false)模块字段的 i18n 语言包 key。
// 幂等：登记时按 module.snake 覆盖式合并节点；注销时删除该节点。
// 字段列表复用 buildGenColumns（与页面生成完全一致），保证 key 不遗漏、不漂移。
// zh-CN 用列注释、en-US 用列名（生成器无翻译能力，便于人工后续翻译）。
func (uc *CodegenUsecase) syncI18nKeys(t *GenTable, cols []*DbColumn, add bool) error {
	if !uc.writeToDisk || uc.outputRoot == "" {
		return nil
	}
	module := strings.TrimSpace(t.ModuleName)
	if module == "" {
		module = "system"
	}
	bizName := strings.TrimSpace(t.BizName)
	if bizName == "" {
		bizName = ToCamel(t.TableName)
	}
	modKey := ToSnake(module)
	snake := ToSnake(bizName)
	if snake == "" {
		return nil
	}

	// 与页面字段保持一致（fields_json 为空时按 cols 兜底），确保语言包 key 与页面 $t 引用一一对应
	gc := buildGenColumns(t, cols, parseFieldConfigs(t.FieldsJSON))

	buildNode := func(lang string) map[string]any {
		comment := strings.TrimSpace(t.TableComment)
		if comment == "" {
			comment = bizName
		}
		node := map[string]any{
			"_title":  comment,
			"_create": "新增" + comment,
			"_edit":   "编辑" + comment,
			"_fill":   "填写" + comment,
		}
		if lang == "en-US" {
			node = map[string]any{
				"_title":  comment,
				"_create": "Create " + comment,
				"_edit":   "Edit " + comment,
				"_fill":   "Fill " + comment,
			}
		}
		// 列字段：zh-CN 取注释（空则回退列名），en-US 直接取列名
		for _, c := range gc {
			if c.JSONName == "" {
				continue
			}
			if lang == "zh-CN" {
				if c.Comment == "" {
					c.Comment = c.Name
				}
				node[c.JSONName] = c.Comment
			} else {
				node[c.JSONName] = c.Name
			}
		}
		return node
	}

	for _, lang := range []string{"zh-CN", "en-US"} {
		path := uc.i18nLangFile(lang)
		root := map[string]any{}
		b, err := os.ReadFile(path)
		if err != nil {
			if !os.IsNotExist(err) {
				return errors.New("读取语言包失败: " + path + ": " + err.Error())
			}
			b = nil
		} else if len(b) > 0 {
			if err := json.Unmarshal(b, &root); err != nil {
				return errors.New("解析语言包失败: " + path + ": " + err.Error())
			}
		}
		codegen := ensureMap(root["codegen"])
		gen := ensureMap(codegen["gen"])
		modNode := ensureMap(gen[modKey])
		if add {
			// 覆盖式合并模块下的业务节点（modNode 可能是新建 map，必须写回 gen）
			modNode[snake] = buildNode(lang)
			gen[modKey] = modNode
		} else {
			delete(modNode, snake)
			// 模块节点清空后整体删除，保持语言包整洁
			if len(modNode) == 0 {
				delete(gen, modKey)
			} else {
				gen[modKey] = modNode
			}
		}
		codegen["gen"] = gen
		root["codegen"] = codegen
		if len(gen) == 0 {
			delete(codegen, "gen")
		}
		out, err := json.MarshalIndent(root, "", "  ")
		if err != nil {
			return err
		}
		if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
			return err
		}
		if err := os.WriteFile(path, append(out, '\n'), 0o644); err != nil {
			return errors.New("写入语言包失败: " + path + ": " + err.Error())
		}
		uc.log.Infof("codegen %si18n keys %s.%s -> %s", map[bool]string{true: "", false: "remove "}[add], modKey, snake, path)
	}
	return nil
}

// ensureMap 取出 map[string]any 值，不存在时新建空 map
func ensureMap(v any) map[string]any {
	if m, ok := v.(map[string]any); ok {
		return m
	}
	return map[string]any{}
}

// applyWireRegistration 对单个宿主文件执行插入/移除：
// add=true 时在 Anchor 行后插入 Insert（Insert 已存在则跳过）；
// add=false 时移除与 Exact 精确匹配（TrimSpace 相等）的行。
func (uc *CodegenUsecase) applyWireRegistration(target string, reg wireRegistration, add bool) error {
	b, err := os.ReadFile(target)
	if err != nil {
		if os.IsNotExist(err) {
			return nil
		}
		return err
	}
	lines := strings.Split(string(b), "\n")
	found := false
	var out []string
	for i, ln := range lines {
		if strings.TrimSpace(ln) == "" {
			out = append(out, ln)
			continue
		}
		if !add && strings.TrimSpace(ln) == strings.TrimSpace(reg.Exact) {
			found = true // 注销：跳过该行
			uc.log.Infof("codegen unregister: %s <= %s", target, strings.TrimSpace(reg.Exact))
			continue
		}
		out = append(out, ln)
		if add && strings.TrimSpace(ln) == strings.TrimSpace(reg.Anchor) {
			// 检查下一行是否已是目标行（幂等）
			if i+1 < len(lines) && strings.TrimSpace(lines[i+1]) == strings.TrimSpace(reg.Exact) {
				found = true
			} else {
				out = append(out, reg.Insert)
				found = true
				uc.log.Infof("codegen register: %s += %s", target, strings.TrimSpace(reg.Insert))
			}
		}
	}
	if !found {
		// 注册但锚点未命中：跳过（锚点可能被用户调整过），不视为错误
		return nil
	}
	return os.WriteFile(target, []byte(strings.Join(out, "\n")), 0o644)
}

// GenerateCode 生成前后端代码
func (uc *CodegenUsecase) GenerateCode(ctx context.Context, req *GenerateCodeInput) ([]*GenCodeFile, error) {
	genReq := &GenTable{
		TableName:    req.TableName,
		ModuleName:   req.ModuleName,
		BizName:      req.BizName,
		TableComment: req.TableComment,
		MenuEnabled:  req.GenMenu,
		MenuModule:   req.MenuModule,
		Buttons:      req.Buttons,
	}

	// 支持基于已保存配置生成
	if req.ID != "" {
		// 从配置表加载（简化：走 repo 查询单条）
		if loaded, err := uc.loadGenTable(ctx, req.ID); err == nil && loaded != nil {
			genReq = loaded
		}
	}

	// 读取真实字段元数据（从零新建时表可能不存在，读取失败不阻断，
	// 交由 Generate 内部按 fields_json 字段配置兜底生成）
	cols, _ := uc.dbMeta.ListDbColumns(ctx, genReq.TableName)

	return uc.gen.Generate(genReq, cols)
}

// GenerateAll 一次调用完成"保存配置 + 生成代码"：
// 内部按功能拆分为多个小方法（构建配置、读列元数据、构建字段模型、渲染各类代码、写盘、落库），
// 任一步失败即中断返回。
//
// 回滚语义：构建配置仅为内存操作，不落库；代码渲染全部成功后，先写盘、最后才保存配置。
// 因此中间任一步（如模板渲染 / 写盘）报错，都不会向 gen_table 写入任何记录，无需额外回滚。
// 写盘内部采用"失败即清理已写文件"策略，避免项目目录残留半成品。
func (uc *CodegenUsecase) GenerateAll(ctx context.Context, req *GenerateAllInput) (*GenerateAllResult, error) {
	// step1: 构建生成配置（纯内存，不落库）
	table, isNew, err := uc.buildGenTableFromInput(ctx, req)
	if err != nil {
		return nil, err
	}

	// step2: 读取真实字段元数据（表可能不存在，读取失败不阻断，
	// 交由 Generate 内部按 fields_json 字段配置兜底生成）
	cols, _ := uc.dbMeta.ListDbColumns(ctx, table.TableName)

	// step3: 构建字段模型 + 渲染各文件（小方法调用）
	// 任一步渲染报错都会在此返回，此时尚未落库，无脏数据残留
	files, err := uc.gen.Generate(table, cols)
	if err != nil {
		return nil, err
	}

	// step4: 渲染成功后，若启用了写盘则写入项目磁盘；失败即返回（已写文件被清理，配置不落库）
	if uc.writeToDisk {
		if err := uc.WriteToProject(files); err != nil {
			return nil, err
		}
		// step4.1: 自动登记 wire 依赖（biz/data/service 三层 ProviderSet + grpc/http server 注册）。
		// 失败时回滚已写文件并返回，避免生成文件与宿主文件状态不一致。
		if err := uc.syncProviderRegistrations(table, true); err != nil {
			for _, f := range files {
				_, _ = uc.removeGeneratedFile(f.FileName)
			}
			return nil, err
		}
		// step4.2: 登记字段 i18n 语言包 key（合并进 lowcode.json），失败回滚已写文件
		if err := uc.syncI18nKeys(table, cols, true); err != nil {
			for _, f := range files {
				_, _ = uc.removeGeneratedFile(f.FileName)
			}
			return nil, err
		}
	}

	// step5: 渲染/写盘全部成功后，才保存/更新生成配置（新建写入 / 已有更新）
	if err := uc.saveGenTable(ctx, table, isNew); err != nil {
		return nil, err
	}

	return &GenerateAllResult{ID: table.ID, Table: table, Files: files}, nil
}

// codegenPathPrefixes 生成文件相对路径前缀 → 项目内实际目录（相对 outputRoot）的映射。
// 按最长前缀优先匹配。
var codegenPathPrefixes = []struct {
	prefix string
	target string
}{
	{"frontend/api/", "frontend/apps/web-antd/src/api/"},
	{"frontend/views/", "frontend/apps/web-antd/src/views/"},
	{"frontend/schemas/", "frontend/apps/web-antd/src/schemas/"},
	{"frontend/", "frontend/"},
	{"proto/", "api/proto/"},
	{"ent/schema/", "app/admin/service/internal/data/ent/schema/"},
	{"internal/", "app/admin/service/internal/"},
	{"docs/", "docs/"},
}

// resolveProjectPath 将生成的相对文件路径映射到项目根目录下的实际路径；
// 无法识别的前缀返回空字符串（调用方跳过写盘）。
func resolveProjectPath(root, fileName string) string {
	name := strings.TrimLeft(fileName, "/")
	for _, m := range codegenPathPrefixes {
		if strings.HasPrefix(name, m.prefix) {
			rel := strings.TrimPrefix(name, m.prefix)
			if rel == "" {
				return ""
			}
			return filepath.Join(root, m.target, rel)
		}
	}
	return ""
}

// WriteToProject 将生成的代码文件写入项目磁盘。
// 任一文件写入失败时，清理本次已写入的文件并返回错误，保证不残留半成品。
func (uc *CodegenUsecase) WriteToProject(files []*GenCodeFile) error {
	if !uc.writeToDisk || uc.outputRoot == "" {
		return nil
	}
	written := make([]string, 0, len(files))
	cleanup := func() {
		for _, p := range written {
			if err := os.Remove(p); err != nil && !os.IsNotExist(err) {
				uc.log.Errorf("codegen rollback remove %s: %v", p, err)
			}
		}
	}
	for _, f := range files {
		target := resolveProjectPath(uc.outputRoot, f.FileName)
		if target == "" {
			uc.log.Warnf("codegen skip unknown path: %s", f.FileName)
			continue
		}
		if err := os.MkdirAll(filepath.Dir(target), 0o755); err != nil {
			cleanup()
			return errors.New("创建目录失败: " + filepath.Dir(target) + ": " + err.Error())
		}
		if err := os.WriteFile(target, []byte(f.Content), 0o644); err != nil {
			cleanup()
			return errors.New("写入文件失败: " + target + ": " + err.Error())
		}
		written = append(written, target)
		uc.log.Infof("codegen wrote %s", target)
	}
	return nil
}

// buildGenTableFromInput 构建生成配置（GenerateAll 的小方法，仅内存操作不落库）。
// 返回值 isNew 标记是否为新建（true=新建，false=更新已有配置）。
func (uc *CodegenUsecase) buildGenTableFromInput(ctx context.Context, req *GenerateAllInput) (*GenTable, bool, error) {
	table := &GenTable{
		TableName:    req.TableName,
		TableComment: req.TableComment,
		ModuleName:   req.ModuleName,
		BizName:      req.BizName,
		FieldsJSON:   req.FieldsJSON,
		GenType:      req.GenType,
		Status:       req.Status,
		MenuEnabled:  req.MenuEnabled,
		MenuModule:   req.MenuModule,
		Buttons:      req.Buttons,
	}

	if req.ID == "" {
		// 新建
		table.ID = uuid.GenerateXID()
		table.CreatedAt = time.Now()
		table.UpdatedAt = time.Now()
		if table.Status == "" {
			table.Status = "enabled"
		}
		if table.GenType == "" {
			table.GenType = "single"
		}
		return table, true, nil
	}

	// 编辑：先加载已保存配置，未显式传入的字段配置保留旧值，避免覆盖丢失
	loaded, err := uc.loadGenTable(ctx, req.ID)
	if err != nil {
		return nil, false, err
	}
	if loaded == nil {
		return nil, false, errors.New("生成配置不存在")
	}
	if len(req.FieldsJSON) == 0 {
		table.FieldsJSON = loaded.FieldsJSON
	}
	if table.TableName == "" {
		table.TableName = loaded.TableName
	}
	if table.TableComment == "" {
		table.TableComment = loaded.TableComment
	}
	if table.ModuleName == "" {
		table.ModuleName = loaded.ModuleName
	}
	if table.BizName == "" {
		table.BizName = loaded.BizName
	}
	if table.GenType == "" {
		table.GenType = loaded.GenType
	}
	if table.Status == "" {
		table.Status = loaded.Status
	}
	if len(table.Buttons) == 0 {
		table.Buttons = loaded.Buttons
	}
	table.ID = req.ID
	table.UpdatedAt = time.Now()
	return table, false, nil
}

// saveGenTable 保存/更新生成配置（GenerateAll 的小方法，仅在所有步骤成功后调用）
func (uc *CodegenUsecase) saveGenTable(ctx context.Context, table *GenTable, isNew bool) error {
	if isNew {
		return uc.repo.CreateGenTable(ctx, table)
	}
	return uc.repo.UpdateGenTable(ctx, table)
}

func (uc *CodegenUsecase) loadGenTable(ctx context.Context, id string) (*GenTable, error) {
	return uc.repo.GetGenTable(ctx, id)
}

// GetGenTable 根据 ID 查询生成配置（供 handler 等外部调用）
func (uc *CodegenUsecase) GetGenTable(ctx context.Context, id string) (*GenTable, error) {
	return uc.repo.GetGenTable(ctx, id)
}

// ---------- 代码生成引擎 ----------

// GenerateAllInput 一次生成调用所需的完整输入（合并"保存配置 + 生成代码"两步）
type GenerateAllInput struct {
	ID           string         // 生成表配置 ID（为空则新建配置并落库）
	TableName    string         // 数据库表名
	TableComment string         // 表备注
	ModuleName   string         // 模块名
	BizName      string         // 业务名（驼峰）
	FieldsJSON   map[string]any // 字段配置（画布字段，前端传入）
	GenType      string         // 生成类型：single=单表 CRUD，form=仅表单
	Status       string         // 状态
	MenuEnabled  bool           // 是否生成菜单
	MenuModule   string         // 菜单所属模块
	Buttons      []string       // 生成的按钮列表
}

// GenerateAllResult 一次生成调用的返回结果
type GenerateAllResult struct {
	ID    string         // 生成表配置 ID
	Table *GenTable      // 保存后的配置（含默认值回填）
	Files []*GenCodeFile // 生成的代码文件
}

// GenerateCodeInput 基于已保存配置生成代码的输入（供下载等场景复用）
type GenerateCodeInput struct {
	ID           string
	TableName    string
	ModuleName   string
	BizName      string
	TableComment string
	GenMenu      bool
	MenuModule   string
	Buttons      []string
}

// 列元数据（带 Go/Proto/前端映射）
type genColumn struct {
	Name         string // 列名（snake_case）
	FieldName    string // 配置的字段名（驼峰，覆盖列名推断）
	GoName       string // Go 字段名（CamelCase）
	JSONName     string // json tag
	Comment      string
	ProtoType    string // proto 标量类型
	GoType       string // Go 类型
	EntType      string // ent 字段类型（String/Int64/Float/Bool/Time/JSON）
	TsType       string // TypeScript 类型（number/string/boolean）
	IsPrimaryKey bool
	// 字段级生成配置（来自 gen_table.fields_json）
	ComponentType string // 前端组件：Input/Select/DatePicker/...
	IsList        bool   // 是否列表展示
	IsForm        bool   // 是否表单展示
	IsQuery       bool   // 是否查询条件
	IsRequired    bool   // 是否必填
	Length        int    // 字段长度（字符串/整型宽度，0 表示不限制）
	// i18n：字段标题 key（如 lowcode.codegen.gen.order.name），模板经 $t() 引用
	I18nKey string
	// 预渲染的 Vue 片段（模板直接输出）
	VueFormControl string // 表单控件片段
	VueCol         string // 列表列定义片段
	VueQueryItem   string // 查询条件片段
}

// 字段级配置（gen_table.fields_json -> fields 数组项）
type genFieldConfig struct {
	ColumnName    string `json:"columnName"`
	ColumnComment string `json:"columnComment"`
	FieldName     string `json:"fieldName"`
	GoType        string `json:"goType"`
	ComponentType string `json:"componentType"`
	List          bool   `json:"list"`
	Form          bool   `json:"form"`
	Query         bool   `json:"query"`
	Required      bool   `json:"required"`
	PrimaryKey    bool   `json:"primaryKey"`
	Length        int    `json:"length"`
}

// parseFieldConfigs 解析 fields_json 中保存的字段配置，返回 列名 -> 配置
func parseFieldConfigs(fieldsJSON map[string]any) map[string]genFieldConfig {
	out := make(map[string]genFieldConfig)
	if fieldsJSON == nil {
		return out
	}
	raw, ok := fieldsJSON["fields"].([]any)
	if !ok {
		return out
	}
	for _, item := range raw {
		data, err := json.Marshal(item)
		if err != nil {
			continue
		}
		var cfg genFieldConfig
		if err := json.Unmarshal(data, &cfg); err != nil {
			continue
		}
		if cfg.ColumnName != "" {
			out[cfg.ColumnName] = cfg
		}
	}
	return out
}

// buildGenColumns 构建生成字段列。
// 字段来源优先级：
//  1. gen_table.fields_json 中手工配置的字段（从零新建时数据库表可能不存在）
//  2. 数据库元数据列（选择数据表方式），字段级配置仅做覆盖
func buildGenColumns(t *GenTable, cols []*DbColumn, configs map[string]genFieldConfig) []genColumn {
	// i18n key 前缀（如 lowcode.codegen.gen.order），按业务名兜底表名
	module := strings.TrimSpace(t.ModuleName)
	if module == "" {
		module = "system"
	}
	bizName := strings.TrimSpace(t.BizName)
	if bizName == "" {
		bizName = ToCamel(t.TableName)
	}
	i18nPrefix := i18nModulePrefix(module, ToSnake(bizName))

	// 方式1：fields_json 里显式配置了字段（含 columnName/goType/componentType），直接按配置构建
	if cfgList, ok := t.FieldsJSON["fields"].([]any); ok && len(cfgList) > 0 {
		gc := make([]genColumn, 0, len(cfgList))
		for _, item := range cfgList {
			data, err := json.Marshal(item)
			if err != nil {
				continue
			}
			var cfg genFieldConfig
			if err := json.Unmarshal(data, &cfg); err != nil {
				continue
			}
			if cfg.ColumnName == "" {
				continue
			}
			// 跳过系统字段（id/created_at/updated_at/deleted_at/tenant_id/tenant_code）。
			// 与方式2一致：proto 模板已写死 string id = 1、createdAt = 100、updatedAt = 101，
			// 若不跳过会导致 message 内重复定义同名字段，buf 编译报
			// `symbol "...id" already defined`。
			if skipColumn(cfg.ColumnName) {
				continue
			}
			col := genColumn{
				Name:          cfg.ColumnName,
				GoName:        ToCamel(cfg.ColumnName),
				JSONName:      ToLowerCamel(cfg.ColumnName),
				Comment:       cfg.ColumnComment,
				IsPrimaryKey:  cfg.PrimaryKey,
				IsList:        cfg.List,
				IsForm:        cfg.Form,
				IsQuery:       cfg.Query,
				IsRequired:    cfg.Required,
				ComponentType: cfg.ComponentType,
				Length:        cfg.Length,
			}
			if cfg.ColumnComment == "" {
				col.Comment = cfg.ColumnName
			}
			if cfg.FieldName != "" {
				col.FieldName = cfg.FieldName
				col.GoName = ToCamel(cfg.FieldName)
				col.JSONName = ToLowerCamel(cfg.FieldName)
			}
			// 有明确 GoType 时以其为准；否则按列名/常见类型兜底为 string
			if cfg.GoType != "" {
				applyGoType(&col, cfg.GoType)
			} else {
				col.GoType = "string"
				col.TsType = "string"
				col.ProtoType = "string"
				col.EntType = "String"
			}
			// 预渲染 Vue 片段
			col.I18nKey = i18nPrefix + "." + col.JSONName
			col.VueCol = buildVueCol(col)
			col.VueQueryItem = buildVueQueryItem(col)
			col.VueFormControl = buildVueFormControl(col)
			gc = append(gc, col)
		}
		return gc
	}

	// 方式2：从数据库元数据构建，字段级配置仅覆盖展示/校验属性
	gc := make([]genColumn, 0, len(cols))
	for _, c := range cols {
		if skipColumn(c.ColumnName) {
			continue
		}
		col := genColumn{
			Name:          c.ColumnName,
			GoName:        ToCamel(c.ColumnName),
			JSONName:      ToLowerCamel(c.ColumnName),
			Comment:       c.ColumnComment,
			ProtoType:     toProtoType(c.DataType),
			GoType:        toGoType(c.DataType),
			EntType:       toEntType(c.DataType),
			TsType:        toTsType(c.DataType),
			IsPrimaryKey:  c.IsPrimaryKey,
			IsList:        !c.IsPrimaryKey,
			IsForm:        !c.IsPrimaryKey,
			ComponentType: mapComponentType(c.DataType),
		}
		if cfg, ok := configs[c.ColumnName]; ok {
			if cfg.ColumnComment != "" {
				col.Comment = cfg.ColumnComment
			}
			if cfg.FieldName != "" {
				col.FieldName = cfg.FieldName
				col.GoName = ToCamel(cfg.FieldName)
				col.JSONName = ToLowerCamel(cfg.FieldName)
			}
			if cfg.GoType != "" {
				applyGoType(&col, cfg.GoType)
			}
			col.ComponentType = cfg.ComponentType
			col.IsList = cfg.List
			col.IsForm = cfg.Form
			col.IsQuery = cfg.Query
			col.IsRequired = cfg.Required
			col.IsPrimaryKey = cfg.PrimaryKey
		}
		// 预渲染 Vue 片段
		col.I18nKey = i18nPrefix + "." + col.JSONName
		col.VueCol = buildVueCol(col)
		col.VueQueryItem = buildVueQueryItem(col)
		col.VueFormControl = buildVueFormControl(col)
		gc = append(gc, col)
	}
	return gc
}

type CodeGenerator struct {
	tpls map[string]*template.Template
}

func NewCodeGenerator() *CodeGenerator {
	funcs := template.FuncMap{
		"inc":      func(i int) int { return i + 1 },
		"add":      func(a, b int) int { return a + b },
		"lower":    func(s string) string { return strings.ToLower(s) },
		"isString": func(s string) bool { return s == "string" },
	}
	tpls := make(map[string]*template.Template, 13)
	for name, src := range map[string]string{
		"proto":      protoTpl,
		"adminProto": adminProtoTpl,
		"ent":        entTpl,
		"biz":        bizTpl,
		"data":       dataTpl,
		"service":    serviceTpl,
		"handler":    handlerTpl,
		"handlerPB":  handlerPBTpl,
		"api":        apiTpl,
		"schema":     schemaTpl,
		"vue":        vueTpl,
		"vueForm":    vueFormTpl,
		"menuSQL":    menuSQLTpl,
		"menuJSON":   menuJSONTpl,
	} {
		tpls[name] = template.Must(template.New(name).Funcs(funcs).Parse(src))
	}
	return &CodeGenerator{
		tpls: tpls,
	}
}

// Generate 根据表 + 列生成一组代码文件（大方法：仅串行调用下方各小方法）
func (g *CodeGenerator) Generate(t *GenTable, cols []*DbColumn) ([]*GenCodeFile, error) {
	model, err := g.BuildModel(t, cols)
	if err != nil {
		return nil, err
	}

	files := make([]*GenCodeFile, 0, 13)
	// 后端代码：proto（消息 + admin 聚合路由）/ ent / biz / data / service
	files = append(files,
		g.RenderProto(model),
		g.RenderAdminProto(model),
		g.RenderEnt(model),
		g.RenderBiz(model),
		g.RenderData(model),
		g.RenderService(model),
		g.RenderAPI(model),
	)
	// 导入/导出：自定义 handler（一个功能一个方法）
	files = append(files, g.RenderHandler(model)...)
	// 前端页面：single=列表页+schema，form=表单页
	files = append(files, g.RenderFrontend(model)...)
	// 菜单：SQL + JSON（一个功能一个方法）
	files = append(files, g.RenderMenu(model)...)

	return files, nil
}

// BuildModel 构建模板渲染数据（小方法）：解析字段配置 -> 生成列 -> 命名规约 -> 按展示位拆分
func (g *CodeGenerator) BuildModel(t *GenTable, cols []*DbColumn) (map[string]any, error) {
	configs := parseFieldConfigs(t.FieldsJSON)
	gc := buildGenColumns(t, cols, configs)
	if len(gc) == 0 {
		return nil, errors.New("暂无可生成的字段，请先在字段配置中维护字段")
	}

	// 业务名 & 命名规约
	bizName := t.BizName
	if bizName == "" {
		bizName = ToCamel(t.TableName)
	}
	// 强制规范化为首字母大写（导出标识符）：兼容用户输入小写（如 test5）或复制产生的小写 bizName。
	// 否则生成的 ent schema 类型为未导出小写（type test5 struct），ent generate 会静默跳过，
	// 导致生成的 data 层代码 import ent/test5 报 "no required module provides package"。
	bizName = ToCamel(bizName)
	lcFirst := LowerFirst(bizName)
	lowerSnake := ToSnake(bizName)
	module := t.ModuleName
	if module == "" {
		module = "system"
	}
	comment := t.TableComment
	if comment == "" {
		comment = bizName
	}

	// 按展示位拆分列
	var listCols, formCols, queryCols []genColumn
	for _, c := range gc {
		if c.IsList {
			listCols = append(listCols, c)
		}
		if c.IsForm {
			formCols = append(formCols, c)
		}
		if c.IsQuery {
			queryCols = append(queryCols, c)
		}
	}

	// 表单列是否包含 string 类型（导入 Excel 循环里 row 变量是否会被引用）
	hasFormString := false
	for _, c := range formCols {
		if c.GoType == "string" {
			hasFormString = true
			break
		}
	}

	// 前后端是否生成表单相关（genType 差异化：form 仅生成表单+提交接口）
	isFormGen := t.GenType == "form"

	// 菜单所属模块（父目录归属，未配置则用模块名）
	menuModule := t.MenuModule
	if menuModule == "" {
		menuModule = module
	}
	// 权限码前缀（模块:业务，如 System:Job / Order:OrderCenter）。
	// 模块名与业务名均走 ToAuthCode 智能分词转首字母大写驼峰，
	// 兼容小写/下划线/驼峰输入：order -> Order，ordercenter -> OrderCenter。
	authPrefix := ToAuthCode(menuModule) + ":" + ToAuthCode(bizName)
	// 按钮列表（未配置默认全选）
	buttons := t.Buttons
	if len(buttons) == 0 {
		buttons = defaultButtons
	}

	return map[string]any{
		"ModuleName":     module,
		"BizName":        bizName,
		"LcFirst":        lcFirst,
		"SnakeName":      lowerSnake,
		"TableName":      t.TableName,
		"Comment":        comment,
		"I18nPrefix":     i18nModulePrefix(module, lowerSnake),
		"Columns":        gc,
		"ListColumns":    listCols,
		"FormColumns":    formCols,
		"QueryColumns":   queryCols,
		"HasFormString":  hasFormString,
		"IsFormGen":      isFormGen,
		"GenMenu":        t.MenuEnabled,
		"MenuModule":     menuModule,
		"AuthPrefix":     authPrefix,
		"Buttons":        buttons,
		"BtnList":        hasButton(buttons, GenBtnList),
		"BtnDetail":      hasButton(buttons, GenBtnDetail),
		"BtnCreate":      hasButton(buttons, GenBtnCreate),
		"BtnEdit":        hasButton(buttons, GenBtnEdit),
		"BtnDelete":      hasButton(buttons, GenBtnDelete),
		"BtnBatchDelete": hasButton(buttons, GenBtnBatchDelete),
		"BtnStatus":      hasButton(buttons, GenBtnStatus),
		"BtnImport":      hasButton(buttons, GenBtnImport),
		"BtnExport":      hasButton(buttons, GenBtnExport),
	}, nil
}

// RenderProto 生成模块 proto 文件（仅消息定义，小方法）
func (g *CodeGenerator) RenderProto(model map[string]any) *GenCodeFile {
	return &GenCodeFile{
		FileName: "proto/" + model["ModuleName"].(string) + "/service/v1/" + strings.ToLower(model["SnakeName"].(string)) + ".proto",
		Content:  g.render("proto", model),
	}
}

// RenderAdminProto 生成 admin 聚合 proto 文件（service + HTTP 路由，小方法）
// 遵循项目架构：service/路由统一声明在 admin/service/v1/i_*.proto，
// 模块 proto 只保留消息定义（参见 i_system.proto、i_identity.proto 等）。
func (g *CodeGenerator) RenderAdminProto(model map[string]any) *GenCodeFile {
	return &GenCodeFile{
		FileName: "proto/admin/service/v1/i_" + strings.ToLower(model["SnakeName"].(string)) + ".proto",
		Content:  g.render("adminProto", model),
	}
}

// RenderEnt 生成 ent schema 文件（小方法）
func (g *CodeGenerator) RenderEnt(model map[string]any) *GenCodeFile {
	return &GenCodeFile{
		FileName: "ent/schema/" + strings.ToLower(model["SnakeName"].(string)) + ".go",
		Content:  g.render("ent", model),
	}
}

// RenderBiz 生成 biz 层文件（小方法）
func (g *CodeGenerator) RenderBiz(model map[string]any) *GenCodeFile {
	return &GenCodeFile{
		FileName: "internal/biz/" + strings.ToLower(model["SnakeName"].(string)) + ".go",
		Content:  g.render("biz", model),
	}
}

// RenderData 生成 data 层文件（小方法）
func (g *CodeGenerator) RenderData(model map[string]any) *GenCodeFile {
	return &GenCodeFile{
		FileName: "internal/data/" + strings.ToLower(model["SnakeName"].(string)) + ".go",
		Content:  g.render("data", model),
	}
}

// RenderService 生成 service 层文件（小方法）
func (g *CodeGenerator) RenderService(model map[string]any) *GenCodeFile {
	return &GenCodeFile{
		FileName: "internal/service/" + strings.ToLower(model["SnakeName"].(string)) + ".go",
		Content:  g.render("service", model),
	}
}

// RenderAPI 生成前端 api 文件（小方法）
func (g *CodeGenerator) RenderAPI(model map[string]any) *GenCodeFile {
	return &GenCodeFile{
		FileName: "frontend/api/" + model["ModuleName"].(string) + "/" + strings.ToLower(model["SnakeName"].(string)) + ".ts",
		Content:  g.render("api", model),
	}
}

// RenderHandler 生成导入/导出自定义 handler（小方法）；未开启导入/导出时返回空
func (g *CodeGenerator) RenderHandler(model map[string]any) []*GenCodeFile {
	if !model["BtnImport"].(bool) && !model["BtnExport"].(bool) {
		return nil
	}
	name := strings.ToLower(model["SnakeName"].(string))
	return []*GenCodeFile{
		{
			FileName: "internal/handler/" + name + ".go",
			Content:  g.render("handler", model),
		},
		{
			FileName: "internal/handler/pb/i_" + name + "_handler_http.pb.go",
			Content:  g.render("handlerPB", model),
		},
	}
}

// RenderFrontend 生成前端页面（小方法）：single=列表页+schema，form=仅表单页
func (g *CodeGenerator) RenderFrontend(model map[string]any) []*GenCodeFile {
	name := strings.ToLower(model["SnakeName"].(string))
	module := model["ModuleName"].(string)
	if model["IsFormGen"].(bool) {
		return []*GenCodeFile{
			{
				FileName: "frontend/views/" + module + "/" + name + "/index.vue",
				Content:  g.render("vueForm", model),
			},
		}
	}
	return []*GenCodeFile{
		{
			FileName: "frontend/views/" + module + "/" + name + "/index.vue",
			Content:  g.render("vue", model),
		},
		{
			FileName: "frontend/schemas/" + module + "/" + name + ".json",
			Content:  g.render("schema", model),
		},
	}
}

// RenderMenu 生成菜单 SQL/JSON 文件（小方法）；未开启菜单时返回空
func (g *CodeGenerator) RenderMenu(model map[string]any) []*GenCodeFile {
	if !model["GenMenu"].(bool) {
		return nil
	}
	name := strings.ToLower(model["SnakeName"].(string))
	return []*GenCodeFile{
		{
			FileName: "docs/menu_" + name + ".sql",
			Content:  g.render("menuSQL", model),
		},
		{
			FileName: "docs/menu_" + name + ".json",
			Content:  g.render("menuJSON", model),
		},
	}
}

func (g *CodeGenerator) render(name string, model map[string]any) string {
	var buf bytes.Buffer
	if err := g.tpls[name].Execute(&buf, model); err != nil {
		return "// 模板渲染失败: " + err.Error()
	}
	return buf.String()
}

func skipColumn(col string) bool {
	switch col {
	case "id", "created_at", "updated_at", "deleted_at", "tenant_id", "tenant_code":
		return true
	}
	return false
}

// i18nModulePrefix 生成模块 i18n key 前缀：lowcode.codegen.gen.<module>.<snake>
// 模块与业务名均走 ToSnake，保证 key 段为合法的小写蛇形（兼容中文模块名降级为原串）。
func i18nModulePrefix(module, snake string) string {
	module = strings.TrimSpace(module)
	if module == "" {
		module = "system"
	}
	snake = strings.TrimSpace(snake)
	if snake == "" {
		snake = "module"
	}
	return "lowcode.codegen.gen." + ToSnake(module) + "." + ToSnake(snake)
}

// toProtoType MySQL 数据类型 -> proto 标量类型
func toProtoType(dt string) string {
	switch dt {
	case "tinyint", "smallint", "int", "mediumint", "bigint":
		return "int64"
	case "float", "double", "decimal":
		return "double"
	case "date", "time", "datetime", "timestamp":
		return "string"
	case "json":
		return "google.protobuf.Struct"
	default:
		return "string"
	}
}

// toGoType MySQL 数据类型 -> Go 类型
func toGoType(dt string) string {
	switch dt {
	case "tinyint", "smallint", "int", "mediumint", "bigint":
		return "int"
	case "float", "double", "decimal":
		return "float64"
	case "boolean":
		return "bool"
	case "date", "time", "datetime", "timestamp":
		return "string"
	default:
		return "string"
	}
}

// toEntType MySQL 数据类型 -> ent 字段类型
func toEntType(dt string) string {
	switch dt {
	case "tinyint", "smallint", "int", "mediumint", "bigint":
		return "Int64"
	case "float", "double", "decimal":
		return "Float"
	case "boolean":
		return "Bool"
	case "date", "time", "datetime", "timestamp":
		return "Time"
	case "json":
		return "JSON"
	default:
		return "String"
	}
}

// toTsType MySQL 数据类型 -> TypeScript 类型
func toTsType(dt string) string {
	switch dt {
	case "tinyint", "smallint", "int", "mediumint", "bigint", "float", "double", "decimal":
		return "number"
	case "boolean":
		return "boolean"
	default:
		return "string"
	}
}

// mapComponentType MySQL 数据类型 -> 默认前端组件
func mapComponentType(dt string) string {
	switch dt {
	case "tinyint", "smallint", "int", "mediumint", "bigint", "float", "double", "decimal":
		return "InputNumber"
	case "date", "time", "datetime", "timestamp":
		return "DatePicker"
	case "text", "mediumtext", "longtext":
		return "Textarea"
	case "json":
		return "Textarea"
	default:
		return "Input"
	}
}

// applyGoType 依据字段配置的 GoType 覆盖列的类型推断（Go/Proto/ent/TS）
func applyGoType(col *genColumn, goType string) {
	col.TsType = "string"
	col.ProtoType = "string"
	col.EntType = "String"
	switch goType {
	case "int", "int64":
		col.GoType = "int64"
		col.TsType = "number"
		col.ProtoType = "int64"
		col.EntType = "Int64"
	case "float", "float64":
		col.GoType = "float64"
		col.TsType = "number"
		col.ProtoType = "double"
		col.EntType = "Float"
	case "bool", "boolean":
		col.GoType = "bool"
		col.TsType = "boolean"
		col.ProtoType = "bool"
		col.EntType = "Bool"
	case "time.Time":
		col.GoType = "time.Time"
		col.ProtoType = "string"
		col.EntType = "Time"
	default:
		col.GoType = "string"
	}
}

// buildVueCol 生成列表页列定义片段（vxe column）。
// 列标题统一走 i18n key（$t），保证与内置页面一致支持中英文切换。
func buildVueCol(c genColumn) string {
	formatter := ""
	if c.EntType == "Time" {
		formatter = "formatDateTime"
	}
	width := "150"
	if c.TsType == "boolean" {
		width = "100"
	}
	title := "$t('" + c.I18nKey + "')"
	if formatter != "" {
		return "      {\n        field: '" + c.JSONName + "',\n        align: 'center',\n        title: " + title + ",\n        minWidth: " + width + ",\n        formatter: '" + formatter + "',\n      },"
	}
	return "      {\n        field: '" + c.JSONName + "',\n        align: 'center',\n        title: " + title + ",\n        minWidth: " + width + ",\n      },"
}

// buildVueQueryItem 生成查询条件片段（vxe formOptions schema 项）。
// label/placeholder 统一走 i18n key。
func buildVueQueryItem(c genColumn) string {
	component := "Input"
	switch c.ComponentType {
	case "Select", "Radio":
		component = "Select"
	case "DatePicker":
		component = "DatePicker"
	case "InputNumber":
		component = "InputNumber"
	}
	label := "$t('" + c.I18nKey + "')"
	return "      {\n        component: '" + component + "',\n        componentProps: {\n          placeholder: " + label + ",\n        },\n        fieldName: '" + c.JSONName + "',\n        label: " + label + ",\n      },"
}

// buildVueFormControl 根据组件类型生成表单控件片段。
// placeholder 统一走 i18n key。
func buildVueFormControl(c genColumn) string {
	jname := c.JSONName
	ph := "$t('" + c.I18nKey + "')"
	switch c.ComponentType {
	case "Textarea":
		return "<Textarea v-model:value=\"form." + jname + "\" :rows=\"3\" :placeholder=\"" + ph + "\" />"
	case "Select":
		return "<Select v-model:value=\"form." + jname + "\" allow-clear :placeholder=\"" + ph + "\" />"
	case "Radio":
		return "<RadioGroup v-model:value=\"form." + jname + "\" />"
	case "Checkbox":
		return "<CheckboxGroup v-model:value=\"form." + jname + "\" />"
	case "DatePicker":
		return "<DatePicker v-model:value=\"form." + jname + "\" value-format=\"YYYY-MM-DD HH:mm:ss\" style=\"width: 100%\" :placeholder=\"" + ph + "\" />"
	case "InputNumber":
		return "<InputNumber v-model:value=\"form." + jname + "\" style=\"width: 100%\" :placeholder=\"" + ph + "\" />"
	case "Switch":
		return "<Switch v-model:checked=\"form." + jname + "\" />"
	default:
		return "<Input v-model:value=\"form." + jname + "\" allow-clear :placeholder=\"" + ph + "\" />"
	}
}

// authWordDict 权限码分词常用词表（小写）。用于把无分隔符的连续小写串拆分为单词，
// 使 ordercenter -> order + center -> OrderCenter。按长度降序贪心匹配，未命中时保留原串。
var authWordDict = buildAuthWordDict()

func buildAuthWordDict() map[string]struct{} {
	words := []string{
		// 常见业务词（按长词优先贪心匹配）
		"permission", "permissions", "organization", "organizations", "department", "departments",
		"transaction", "transactions", "notification", "notifications", "announcement", "announcements",
		"configuration", "dictionary", "dictionaries", "statistic", "statistics", "specification", "specifications",
		"settlement", "settlements", "scheduler", "dashboard", "category", "categories", "attribute", "attributes",
		"management", "supplier", "suppliers", "customer", "customers", "warehouse", "address", "addresses",
		"promotion", "promotions", "activity", "activities", "campaign", "campaigns", "interface", "interfaces",
		"consumer", "consumers", "producer", "producers", "subscribe", "subscribers", "pagination",
		"order", "orders", "center", "centers", "user", "users", "admin", "system", "tenant", "tenants",
		"account", "accounts", "menu", "menus", "role", "roles", "auth", "authz", "login", "logout",
		"register", "verify", "captcha", "token", "refresh", "password", "dept", "depts", "org", "orgs",
		"post", "posts", "position", "positions", "dict", "dicts", "config", "configs", "parameter", "parameters",
		"param", "params", "setting", "settings", "option", "options", "log", "logs", "audit", "audits",
		"operator", "operation", "operations", "task", "tasks", "job", "jobs", "schedule", "schedules",
		"cron", "notice", "notices", "message", "messages", "msg", "file", "files", "storage", "upload",
		"uploads", "download", "downloads", "import", "imports", "export", "exports", "batch", "detail",
		"details", "info", "infos", "create", "created", "update", "updated", "delete", "deleted", "edit",
		"edited", "list", "lists", "page", "pages", "query", "queries", "search", "switch", "enable",
		"disable", "status", "state", "states", "monitor", "monitors", "server", "servers", "cache", "caches",
		"database", "databases", "api", "apis", "report", "reports", "home", "goods", "product", "products",
		"sku", "stock", "payment", "payments", "invoice", "invoices", "freight", "express", "ship", "shipping",
		"store", "stores", "shop", "brand", "brands", "value", "values", "price", "prices", "cost", "money",
		"amount", "amounts", "count", "number", "numbers", "name", "names", "title", "titles", "content",
		"contents", "remark", "remarks", "comment", "comments", "reply", "replies", "level", "levels",
		"grade", "rate", "rates", "ratio", "area", "areas", "region", "regions", "city", "cities",
		"province", "provinces", "country", "countries", "zone", "zones", "group", "groups", "team", "teams",
		"member", "members", "staff", "employee", "employees", "student", "students", "course", "courses",
		"class", "classes", "subject", "subjects", "question", "questions", "answer", "answers", "paper",
		"papers", "exam", "exams", "test", "tests", "score", "scores", "result", "results", "award",
		"coupon", "coupons", "point", "points", "reward", "rewards", "gift", "cards", "ticket", "tickets",
		"event", "events", "banner", "banners", "advert", "discount", "discounts", "finance", "financial",
		"fund", "funds", "wallet", "wallets", "balance", "balances", "trade", "trades", "record", "records",
		"history", "histories", "session", "sessions", "online", "visitor", "visitors", "traffic", "visit",
		"visits", "click", "clicks", "view", "views", "share", "shares", "like", "likes", "collect",
		"collection", "favorite", "favorites", "follow", "follows", "push", "send", "sends", "receive",
		"receives", "mail", "email", "emails", "sms", "phone", "phones", "mobile", "time", "date", "dates",
		"week", "month", "months", "year", "years", "day", "days", "hour", "hours", "minute", "minutes",
		"second", "seconds", "period", "periods", "cycle", "cycles", "interval", "intervals", "start",
		"end", "ends", "begin", "finish", "close", "open", "init", "initialize", "sync", "merge", "move",
		"copy", "paste", "save", "saves", "cancel", "confirm", "reject", "approve", "approval", "review",
		"check", "validate", "validation", "submit", "apply", "agree", "refuse", "suspend", "pause", "stop",
		"resume", "continue", "complete", "completed", "done", "pending", "wait", "timeout", "fail", "failed",
		"error", "errors", "warn", "warning", "warnings", "tip", "hints", "note", "dialog", "modal",
		"drawer", "form", "forms", "table", "tables", "grid", "chart", "charts", "graph", "graphs", "tree",
		"trees", "card", "cards", "preview", "previews", "add", "adds", "remove", "removes", "clear",
		"filter", "filters", "sort", "sorts", "total", "totals", "empty", "none", "all", "first", "last",
		"next", "previous", "current", "index", "sum", "avg", "min", "max", "top", "rank", "mode", "modes",
		"type", "types", "kind", "kinds", "tag", "tags", "label", "labels", "key", "keys", "field", "fields",
		"column", "columns", "row", "rows", "cell", "cells", "item", "items", "entry", "entries", "data",
		"dataset", "model", "models", "entity", "entities", "struct", "schema", "schemas", "meta", "json",
		"xml", "html", "css", "js", "ts", "vue", "react", "web", "app", "apps", "client", "clients",
		"gateway", "proxy", "route", "routes", "router", "routers", "link", "links", "url", "domain",
		"host", "port", "ip", "ssl", "cert", "secret", "secrets", "encrypt", "decrypt", "hash", "salt",
		"sign", "cookie", "cookies", "redis", "mq", "kafka", "rabbit", "queue", "queues", "topic", "topics",
		"exchange", "stream", "streams", "sms", "mail", "score", "group", "rule", "rules", "workflow",
		"process", "processes", "flow", "flows", "template", "templates", "demo",
	}
	m := make(map[string]struct{}, len(words))
	for _, w := range words {
		m[w] = struct{}{}
	}
	return m
}

// ToAuthCode 业务名/模块名 -> 权限码规范的 PascalCase（首字母大写驼峰 + 智能分词）：
//
//	order -> Order
//	order_center / orderCenter / OrderCenter -> OrderCenter
//	ordercenter -> OrderCenter（词表分词）
//
// 未命中词表的连续小写串按整词处理（仅首字母大写，如 customxyz -> Customxyz）。
func ToAuthCode(s string) string {
	var words []string
	for _, seg := range strings.FieldsFunc(strings.TrimSpace(s), func(r rune) bool {
		return r == '_' || r == '-' || r == ' ' || r == '.'
	}) {
		if seg == "" {
			continue
		}
		words = append(words, splitAuthSegment(strings.ToLower(seg))...)
	}
	for i, w := range words {
		if w == "" {
			continue
		}
		rs := []rune(w)
		rs[0] = unicode.ToUpper(rs[0])
		words[i] = string(rs)
	}
	return strings.Join(words, "")
}

// splitAuthSegment 把单个小写分段按词表贪心拆分为单词；数字单独成段。
func splitAuthSegment(lower string) []string {
	// 整个串命中词表直接返回
	if _, ok := authWordDict[lower]; ok {
		return []string{lower}
	}
	var words []string
	rest := lower
	for len(rest) > 0 {
		// 数字段单独切出
		if rest[0] >= '0' && rest[0] <= '9' {
			j := 1
			for j < len(rest) && rest[j] >= '0' && rest[j] <= '9' {
				j++
			}
			words = append(words, rest[:j])
			rest = rest[j:]
			continue
		}
		// 词表贪心最长前缀匹配
		matched := false
		for i := len(rest); i >= 1; i-- {
			if _, ok := authWordDict[rest[:i]]; ok {
				words = append(words, rest[:i])
				rest = rest[i:]
				matched = true
				break
			}
		}
		if !matched {
			// 剩余无法分词，整段作为一个词
			words = append(words, rest)
			break
		}
	}
	return words
}

// ToCamel snake_case -> CamelCase (FooBar)
func ToCamel(s string) string {
	parts := strings.Split(s, "_")
	var b strings.Builder
	for _, p := range parts {
		if p == "" {
			continue
		}
		rs := []rune(p)
		rs[0] = unicode.ToUpper(rs[0])
		b.WriteString(string(rs))
	}
	return b.String()
}

// ToLowerCamel snake_case -> lowerCamel (fooBar)
func ToLowerCamel(s string) string {
	c := ToCamel(s)
	return LowerFirst(c)
}

// LowerFirst 首字母小写
func LowerFirst(s string) string {
	if s == "" {
		return s
	}
	rs := []rune(s)
	rs[0] = unicode.ToLower(rs[0])
	return string(rs)
}

// ToSnake CamelCase -> snake_case
func ToSnake(s string) string {
	var b strings.Builder
	for i, r := range s {
		if unicode.IsUpper(r) {
			if i > 0 {
				b.WriteByte('_')
			}
			b.WriteRune(unicode.ToLower(r))
		} else {
			b.WriteRune(r)
		}
	}
	return b.String()
}

// ---------- 代码生成模板 ----------

const protoTpl = `syntax = "proto3";

package {{.ModuleName}}.service.v1;

import "buf/validate/validate.proto";
import "gnostic/openapi/v3/annotations.proto";
import "google/protobuf/struct.proto";
import "google/protobuf/timestamp.proto";

// 注意：该文件仅定义消息结构（Message），HTTP 路由与 Service 定义统一在
// admin/service/v1/i_{{.SnakeName}}.proto 中聚合声明（与 system/service/v1/dict_type.proto 等模块保持一致）。

// {{.Comment}}
message {{.BizName}} {
  string id = 1;
{{- range $i, $c := .Columns}}
  // {{$c.Comment}}
{{- if and $c.IsRequired (ne $c.ProtoType "google.protobuf.Struct") }}
  {{$c.ProtoType}} {{$c.JSONName}} = {{add $i 2}} [(buf.validate.field).required = true];
{{- else }}
  {{$c.ProtoType}} {{$c.JSONName}} = {{add $i 2}};
{{- end }}
{{- end}}
  google.protobuf.Timestamp createdAt = 100;
  google.protobuf.Timestamp updatedAt = 101;
}

// {{.Comment}}列表查询
message List{{.BizName}}Request {
  int64 currentPage = 1 [(gnostic.openapi.v3.property) = {description: "当前页码（从1开始）"}];
  int64 pageSize = 2 [(gnostic.openapi.v3.property) = {description: "每页条数（传0为全部）"}];
{{- range $i, $c := .QueryColumns}}
  {{$c.ProtoType}} {{$c.JSONName}} = {{add $i 3}} [(gnostic.openapi.v3.property) = {description: "{{$c.Comment}}"}];
{{- end}}
}

message List{{.BizName}}Response {
  repeated {{.BizName}} items = 1;
  int64 total = 2;
}

{{- if .BtnDetail }}
// {{.Comment}}详情
message Get{{.BizName}}Request {
  string id = 1 [
    (buf.validate.field).cel = {
      id: "{{.SnakeName}}.id.required"
      message: "{{.SnakeName}}.id.required"
      expression: "this.size() > 0"
    },
    (gnostic.openapi.v3.property) = {description: "ID"}
  ];
}

message Get{{.BizName}}Response {
  {{.BizName}} data = 1;
}
{{- end }}

{{- if .BtnCreate }}
// 创建{{.Comment}}
message Create{{.BizName}}Request {
{{- range $i, $c := .FormColumns}}
{{- if $c.IsRequired }}
  {{$c.ProtoType}} {{$c.JSONName}} = {{inc $i}} [(buf.validate.field).required = true];
{{- else }}
  {{$c.ProtoType}} {{$c.JSONName}} = {{inc $i}};
{{- end }}
{{- end}}
}

message Create{{.BizName}}Response {
  {{.BizName}} data = 1;
}
{{- end }}

{{- if .BtnEdit }}
// 更新{{.Comment}}
message Update{{.BizName}}Request {
  string id = 1 [
    (buf.validate.field).cel = {
      id: "{{.SnakeName}}.id.required"
      message: "{{.SnakeName}}.id.required"
      expression: "this.size() > 0"
    }
  ];
{{- range $i, $c := .FormColumns}}
  {{$c.ProtoType}} {{$c.JSONName}} = {{add $i 2}};
{{- end}}
}

message Update{{.BizName}}Response {
  {{.BizName}} data = 1;
}
{{- end }}

{{- if or .BtnDelete .BtnBatchDelete }}
// 删除{{.Comment}}
message Delete{{.BizName}}Request {
  repeated string ids = 1 [
    (buf.validate.field).repeated = {
      min_items: 1
      max_items: 100
    },
    (gnostic.openapi.v3.property) = {description: "ID列表，至少一个"}
  ];
}

message Delete{{.BizName}}Response {}
{{- end }}

{{- if .BtnStatus }}
// 更新{{.Comment}}状态
message Update{{.BizName}}StatusRequest {
  string id = 1 [
    (buf.validate.field).cel = {
      id: "{{.SnakeName}}.id.required"
      message: "{{.SnakeName}}.id.required"
      expression: "this.size() > 0"
    }
  ];
  string status = 2 [(gnostic.openapi.v3.property) = {description: "状态"}];
}

message Update{{.BizName}}StatusResponse {}
{{- end }}
`

const adminProtoTpl = `syntax = "proto3";

package admin.service.v1;

import "gnostic/openapi/v3/annotations.proto";
import "google/api/annotations.proto";
import "google/protobuf/empty.proto";
import "{{.ModuleName}}/service/v1/{{.SnakeName}}.proto";

// {{.Comment}}服务（HTTP 路由聚合声明，与其他模块 i_*.proto 保持一致）
service {{.BizName}}Service {
{{- if .BtnList }}
  // 列表查询
  rpc List{{.BizName}}({{.ModuleName}}.service.v1.List{{.BizName}}Request) returns ({{.ModuleName}}.service.v1.List{{.BizName}}Response) {
    option (google.api.http) = {get: "/admin/v1/{{.SnakeName}}s"};
    option (gnostic.openapi.v3.operation) = {
      tags: "{{.MenuModule}}"
      summary: "{{.Comment}}-列表"
      description: "{{.Comment}}分页列表"
    };
  }
{{- end }}
{{- if .BtnDetail }}
  // 详情
  rpc Get{{.BizName}}({{.ModuleName}}.service.v1.Get{{.BizName}}Request) returns ({{.ModuleName}}.service.v1.Get{{.BizName}}Response) {
    option (google.api.http) = {get: "/admin/v1/{{.SnakeName}}s/{id}"};
    option (gnostic.openapi.v3.operation) = {
      tags: "{{.MenuModule}}"
      summary: "{{.Comment}}-详情"
      description: "根据ID获取{{.Comment}}详情"
    };
  }
{{- end }}
{{- if .BtnCreate }}
  // 创建
  rpc Create{{.BizName}}({{.ModuleName}}.service.v1.Create{{.BizName}}Request) returns ({{.ModuleName}}.service.v1.Create{{.BizName}}Response) {
    option (google.api.http) = {
      post: "/admin/v1/{{.SnakeName}}s"
      body: "*"
    };
    option (gnostic.openapi.v3.operation) = {
      tags: "{{.MenuModule}}"
      summary: "{{.Comment}}-创建"
      description: "创建{{.Comment}}"
    };
  }
{{- end }}
{{- if .BtnEdit }}
  // 更新
  rpc Update{{.BizName}}({{.ModuleName}}.service.v1.Update{{.BizName}}Request) returns ({{.ModuleName}}.service.v1.Update{{.BizName}}Response) {
    option (google.api.http) = {
      put: "/admin/v1/{{.SnakeName}}s/{id}"
      body: "*"
    };
    option (gnostic.openapi.v3.operation) = {
      tags: "{{.MenuModule}}"
      summary: "{{.Comment}}-修改"
      description: "修改{{.Comment}}"
    };
  }
{{- end }}
{{- if or .BtnDelete .BtnBatchDelete }}
  // 删除（单条/批量）
  rpc Delete{{.BizName}}({{.ModuleName}}.service.v1.Delete{{.BizName}}Request) returns ({{.ModuleName}}.service.v1.Delete{{.BizName}}Response) {
    option (google.api.http) = {delete: "/admin/v1/{{.SnakeName}}s"};
    option (gnostic.openapi.v3.operation) = {
      tags: "{{.MenuModule}}"
      summary: "{{.Comment}}-删除(批量)"
      description: "删除{{.Comment}}(支持批量)"
    };
  }
{{- end }}
{{- if .BtnStatus }}
  // 状态切换
  rpc Update{{.BizName}}Status({{.ModuleName}}.service.v1.Update{{.BizName}}StatusRequest) returns ({{.ModuleName}}.service.v1.Update{{.BizName}}StatusResponse) {
    option (google.api.http) = {
      put: "/admin/v1/{{.SnakeName}}s/{id}/status"
      body: "*"
    };
    option (gnostic.openapi.v3.operation) = {
      tags: "{{.MenuModule}}"
      summary: "{{.Comment}}-更新状态"
      description: "更新{{.Comment}}状态"
    };
  }
{{- end }}
}
`

const entTpl = `package schema

import (
	"time"

	"entgo.io/ent"
	"entgo.io/ent/dialect/entsql"
	"entgo.io/ent/schema"
	"entgo.io/ent/schema/field"
)

// {{.BizName}} {{.Comment}}
type {{.BizName}} struct {
	ent.Schema
}

func ({{.BizName}}) Annotations() []schema.Annotation {
	return []schema.Annotation{
		entsql.Annotation{Table: "{{.TableName}}"},
		entsql.WithComments(true),
		schema.Comment("{{.Comment}}"),
	}
}

// Mixin of the {{.BizName}}.
func ({{.BizName}}) Mixin() []ent.Mixin {
	return []ent.Mixin{TenantMixin{}}
}

func ({{.BizName}}) Fields() []ent.Field {
	return []ent.Field{
		field.String("id").
			MaxLen(36).
			Unique().
			Immutable().
			Comment("ID"),
{{- range .Columns}}
{{- if eq .EntType "JSON"}}
		field.JSON("{{.Name}}", map[string]any{}).
			Optional().
			Comment("{{.Comment}}"),
{{- else if eq .EntType "Time"}}
		field.Time("{{.Name}}").
{{- if not .IsRequired }}
			Optional().
{{- end }}
			Comment("{{.Comment}}"),
{{- else}}
		field.{{.EntType}}("{{.Name}}").
{{- if not .IsRequired }}
			Optional().
{{- end }}
{{- if eq .EntType "String" }}
{{- if gt .Length 0 }}
			MaxLen({{.Length}}).
{{- end }}
{{- end }}
			Comment("{{.Comment}}"),
{{- end }}
{{- end}}
		field.Time("created_at").
			Immutable().
			Default(time.Now).
			Comment("创建时间"),
		field.Time("updated_at").
			Default(time.Now).
			UpdateDefault(time.Now).
			Comment("更新时间"),
		field.Time("deleted_at").
			Optional().
			Nillable().
			Comment("删除时间"),
	}
}
`

const bizTpl = `package biz

import (
	"context"
	"time"

	"github.com/antsurge/weaver-admin/pkg/enthelper"
	"github.com/antsurge/weaver-admin/pkg/utils/uuid"
	"github.com/go-kratos/kratos/v2/log"
)

// {{.BizName}} {{.Comment}}
type {{.BizName}} struct {
	ID string ` + "`json:\"id\"`" + `
{{- range .Columns}}
	{{.GoName}} {{.GoType}} ` + "`json:\"{{.JSONName}}\"`" + `
{{- end}}
	CreatedAt time.Time  ` + "`json:\"createdAt\"`" + `
	UpdatedAt time.Time  ` + "`json:\"updatedAt\"`" + `
	DeletedAt *time.Time ` + "`json:\"deletedAt,omitempty\"`" + `
}

type List{{.BizName}}Request struct {
	enthelper.PaginationParams
}

type List{{.BizName}}Response struct {
	Data  []*{{.BizName}}
	Total int
}

type {{.BizName}}Repo interface {
	List(context.Context, *List{{.BizName}}Request) (*List{{.BizName}}Response, error)
{{- if .BtnDetail }}
	Get(context.Context, string) (*{{.BizName}}, error)
{{- end }}
	Create(context.Context, *{{.BizName}}) error
	Update(context.Context, *{{.BizName}}) error
{{- if .BtnStatus }}
	UpdateStatus(context.Context, string, string) error
{{- end }}
	Delete(context.Context, []string) error
}

type {{.BizName}}Usecase struct {
	repo {{.BizName}}Repo
	log  *log.Helper
}

func New{{.BizName}}Usecase(repo {{.BizName}}Repo, logger log.Logger) *{{.BizName}}Usecase {
	return &{{.BizName}}Usecase{repo: repo, log: log.NewHelper(logger)}
}

func (uc *{{.BizName}}Usecase) List(ctx context.Context, req *List{{.BizName}}Request) (*List{{.BizName}}Response, error) {
	return uc.repo.List(ctx, req)
}

{{- if .BtnDetail }}
func (uc *{{.BizName}}Usecase) Get(ctx context.Context, id string) (*{{.BizName}}, error) {
	return uc.repo.Get(ctx, id)
}
{{- end }}

func (uc *{{.BizName}}Usecase) Create(ctx context.Context, req *{{.BizName}}) (*{{.BizName}}, error) {
	item := req
	item.ID = uuid.GenerateXID()
	item.CreatedAt = time.Now()
	item.UpdatedAt = time.Now()
	err := uc.repo.Create(ctx, item)
	return item, err
}

func (uc *{{.BizName}}Usecase) Update(ctx context.Context, req *{{.BizName}}) (*{{.BizName}}, error) {
	item := req
	item.UpdatedAt = time.Now()
	err := uc.repo.Update(ctx, item)
	return item, err
}

{{- if .BtnStatus }}
func (uc *{{.BizName}}Usecase) UpdateStatus(ctx context.Context, id, status string) error {
	return uc.repo.UpdateStatus(ctx, id, status)
}
{{- end }}

func (uc *{{.BizName}}Usecase) Delete(ctx context.Context, ids []string) error {
	return uc.repo.Delete(ctx, ids)
}
`

const dataTpl = `package data

import (
	"context"
	"strconv"
	"time"

	"github.com/antsurge/weaver-admin/app/admin/service/internal/biz"
	"github.com/antsurge/weaver-admin/app/admin/service/internal/data/ent"
	"github.com/antsurge/weaver-admin/app/admin/service/internal/data/ent/{{.SnakeName}}"
	"github.com/antsurge/weaver-admin/pkg/enthelper"
	"github.com/go-kratos/kratos/v2/log"
)

var _ biz.{{.BizName}}Repo = (*{{.LcFirst}}Repo)(nil)

type {{.LcFirst}}Repo struct {
	data *Data
	log  *log.Helper
}

func New{{.BizName}}Repo(data *Data, logger log.Logger) biz.{{.BizName}}Repo {
	return &{{.LcFirst}}Repo{data: data, log: log.NewHelper(logger)}
}

func (r *{{.LcFirst}}Repo) List(ctx context.Context, params *biz.List{{.BizName}}Request) (*biz.List{{.BizName}}Response, error) {
	query := r.data.db.{{.BizName}}.Query().Order(ent.Desc({{.SnakeName}}.FieldCreatedAt)).
		Where({{.SnakeName}}.DeletedAtIsNil())

	res, err := enthelper.Pagination[*ent.{{.BizName}}, *ent.{{.BizName}}Query](ctx, query, params.PaginationParams)
	if err != nil {
		return nil, err
	}

	data := make([]*biz.{{.BizName}}, 0, res.Total)
	for _, v := range res.Data {
		data = append(data, &biz.{{.BizName}}{
			ID:        v.ID,
{{- range .Columns}}
			{{.GoName}}: v.{{.GoName}},
{{- end}}
			CreatedAt: v.CreatedAt,
			UpdatedAt: v.UpdatedAt,
			DeletedAt: v.DeletedAt,
		})
	}
	return &biz.List{{.BizName}}Response{Data: data, Total: res.Total}, nil
}

{{- if .BtnDetail }}
func (r *{{.LcFirst}}Repo) Get(ctx context.Context, id string) (*biz.{{.BizName}}, error) {
	v, err := r.data.db.{{.BizName}}.Get(ctx, id)
	if err != nil {
		return nil, err
	}
	return &biz.{{.BizName}}{
		ID:        v.ID,
{{- range .Columns}}
		{{.GoName}}: v.{{.GoName}},
{{- end}}
		CreatedAt: v.CreatedAt,
		UpdatedAt: v.UpdatedAt,
		DeletedAt: v.DeletedAt,
	}, nil
}
{{- end }}

func (r *{{.LcFirst}}Repo) Create(ctx context.Context, d *biz.{{.BizName}}) error {
	_, err := r.data.db.{{.BizName}}.Create().
		SetID(d.ID).
{{- range .Columns}}
		Set{{.GoName}}(d.{{.GoName}}).
{{- end}}
		SetCreatedAt(d.CreatedAt).
		SetUpdatedAt(d.UpdatedAt).
		Save(ctx)
	return err
}

func (r *{{.LcFirst}}Repo) Update(ctx context.Context, d *biz.{{.BizName}}) error {
	_, err := r.data.db.{{.BizName}}.UpdateOneID(d.ID).
{{- range .Columns}}
		Set{{.GoName}}(d.{{.GoName}}).
{{- end}}
		SetUpdatedAt(d.UpdatedAt).
		Save(ctx)
	return err
}

{{- if .BtnStatus }}
func (r *{{.LcFirst}}Repo) UpdateStatus(ctx context.Context, id, status string) error {
	s, err := strconv.ParseInt(status, 10, 64)
	if err != nil {
		return err
	}
	_, err = r.data.db.{{.BizName}}.UpdateOneID(id).
		SetStatus(s).
		SetUpdatedAt(time.Now()).
		Save(ctx)
	return err
}
{{- end }}

func (r *{{.LcFirst}}Repo) Delete(ctx context.Context, ids []string) error {
	now := time.Now()
	_, err := r.data.db.{{.BizName}}.Update().
		Where({{.SnakeName}}.IDIn(ids...)).
		SetDeletedAt(now).
		Save(ctx)
	return err
}
`

const apiTpl = `import { requestClient } from '#/api/request';

// {{.Comment}}

export interface {{.BizName}} {
  id: string;
{{- range .Columns}}
  {{.JSONName}}: {{.TsType}};
{{- end}}
  createdAt?: string;
  updatedAt?: string;
}

export interface {{.BizName}}PageParams {
  currentPage?: number;
  pageSize?: number;
{{- range .QueryColumns}}
  {{.JSONName}}?: {{.TsType}};
{{- end}}
}

export function list{{.BizName}}Api(params: {{.BizName}}PageParams) {
  return requestClient.get('/admin/v1/{{.SnakeName}}s', { params });
}

{{- if .BtnDetail }}

export function get{{.BizName}}Api(id: string) {
  return requestClient.get(` + "`/admin/v1/{{.SnakeName}}s/${id}`" + `);
}
{{- end }}

export function create{{.BizName}}Api(data: Partial<{{.BizName}}>) {
  return requestClient.post('/admin/v1/{{.SnakeName}}s', data);
}

export function update{{.BizName}}Api(id: string, data: Partial<{{.BizName}}>) {
  return requestClient.put(` + "`/admin/v1/{{.SnakeName}}s/${id}`" + `, data);
}

{{- if .BtnStatus }}

export function update{{.BizName}}StatusApi(id: string, status: string) {
  return requestClient.put(` + "`/admin/v1/{{.SnakeName}}s/${id}/status`" + `, { status });
}
{{- end }}

export function delete{{.BizName}}Api(ids: string[]) {
  return requestClient.delete('/admin/v1/{{.SnakeName}}s', { data: { ids } });
}

{{- if .BtnImport }}

export function import{{.BizName}}Api(data: FormData) {
  return requestClient.post('/admin/v1/{{.SnakeName}}s/import', data, {
    headers: { 'Content-Type': 'multipart/form-data' },
  });
}
{{- end }}

{{- if .BtnExport }}

export function export{{.BizName}}Api(params: {{.BizName}}PageParams) {
  return requestClient.post('/admin/v1/{{.SnakeName}}s/export', params, {
    responseType: 'blob',
    responseReturn: 'raw',
    showFailMessage: false,
  });
}
{{- end }}
`

const schemaTpl = `{
  "type": "object",
  "title": "{{.Comment}}",
  "properties": {
{{- range .Columns}}
    "{{.JSONName}}": {
      "title": "{{.Comment}}",
{{- if eq .ProtoType "int64"}}
      "type": "integer"
{{- else if eq .ProtoType "double"}}
      "type": "number"
{{- else if eq .ProtoType "bool"}}
      "type": "boolean"
{{- else}}
      "type": "string"
{{- end}}
    },
{{- end}}
    "createdAt": { "title": "创建时间", "type": "string" },
    "updatedAt": { "title": "更新时间", "type": "string" }
  }
}
`

const vueTpl = `<script lang="ts" setup>
import type { VxeTableGridOptions } from '#/adapter/vxe-table';
import type { {{.BizName}} } from '#/api/{{.ModuleName}}/{{.SnakeName}}';

import { reactive, ref } from 'vue';

import { Page } from '@vben/common-ui';
import { Plus } from '@vben/icons';
import { $t } from '#/locales';

import {
  Button,
  CheckboxGroup,
  DatePicker,
  Form,
  FormItem,
  Input,
  InputNumber,
  message,
  Modal,
  RadioGroup,
  Select,
  Switch,
  Textarea,
} from 'ant-design-vue';

import { useVbenVxeGrid } from '#/adapter/vxe-table';
import { DEFAULT_PAGE_SIZE, PAGE_SIZES } from '#/types/pagination';

import {
  create{{.BizName}}Api,
  delete{{.BizName}}Api,
  list{{.BizName}}Api,
  update{{.BizName}}Api,
} from '#/api/{{.ModuleName}}/{{.SnakeName}}';

defineOptions({ name: '{{.BizName}}Management' });

const [Grid, gridApi] = useVbenVxeGrid({
  formOptions: {
    schema: [
{{- range .QueryColumns}}
{{.VueQueryItem}}
{{- end}}
    ],
  },
  gridOptions: {
    columns: [
{{- range .ListColumns}}
{{.VueCol}}
{{- end}}
      {
        field: 'operation',
        align: 'center',
        title: $t('common.fields.operation'),
        width: 150,
        slots: { default: 'operation' },
      },
    ],
    height: 'auto',
    keepSource: true,
    pagerConfig: {
      enabled: true,
      pageSize: DEFAULT_PAGE_SIZE,
      pageSizes: PAGE_SIZES,
    },
    proxyConfig: {
      ajax: {
        query: async ({ page }, formValues) => {
          const res = await list{{.BizName}}Api({
            ...page,
            ...formValues,
          });
          return res;
        },
      },
    },
    rowConfig: {
      keyField: 'id',
    },
{{- if .BtnBatchDelete }}
    checkboxConfig: {
      checkStrictly: true,
      showHeader: true,
    },
{{- end }}
    toolbarConfig: {
      custom: true,
      export: false,
      refresh: true,
      zoom: true,
      search: true,
    },
  } as VxeTableGridOptions,
});

{{- if .BtnBatchDelete }}
const selectedRows = ref<{{.BizName}}[]>([]);
{{- end }}

const formRef = ref();
const modalVisible = ref(false);
const submitting = ref(false);
const isEdit = ref(false);
const form = reactive<Partial<{{.BizName}}>>({
{{- range .FormColumns}}
  {{.JSONName}}: undefined,
{{- end}}
});

const rules: Record<string, any> = {
{{- range .FormColumns}}
{{- if .IsRequired}}
  {{.JSONName}}: [
    { required: true, message: $t('{{.I18nKey}}') + $t('lowcode.codegen.genRequired') },
  ],
{{- end}}
{{- end}}
};

function onRefresh() {
  gridApi.query();
}

{{- if .BtnBatchDelete }}
function onCheckboxChange() {
  const checkboxRecords = gridApi.grid?.getCheckboxRecords?.() ?? [];
  selectedRows.value = checkboxRecords;
}

function onBatchDelete() {
  if (!selectedRows.value.length) return;
  const ids = selectedRows.value.map((r) => r.id);
  Modal.confirm({
    title: $t('common.confirm'),
    content: $t('common.actions.delete'),
    async onOk() {
      await delete{{.BizName}}Api(ids);
      message.success('删除成功');
      selectedRows.value = [];
      onRefresh();
    },
  });
}
{{- end }}

function onCreate() {
  isEdit.value = false;
  Object.keys(form).forEach((key) => {
    form[key as keyof typeof form] = undefined;
  });
  modalVisible.value = true;
}

function onEdit(row: {{.BizName}}) {
  isEdit.value = true;
  Object.assign(form, row);
  modalVisible.value = true;
}

function onDelete(row: {{.BizName}}) {
  Modal.confirm({
    title: $t('common.confirm'),
    content: $t('common.actions.delete'),
    async onOk() {
      await delete{{.BizName}}Api([row.id]);
      message.success('删除成功');
      onRefresh();
    },
  });
}

async function onSubmit() {
  try {
    await formRef.value.validate();
  } catch {
    return;
  }
  submitting.value = true;
  try {
    if (isEdit.value) {
      await update{{.BizName}}Api(form.id as string, form);
    } else {
      await create{{.BizName}}Api(form);
    }
    message.success(isEdit.value ? '更新成功' : '创建成功');
    modalVisible.value = false;
    onRefresh();
  } finally {
    submitting.value = false;
  }
}
</script>

<template>
  <Page auto-content-height>
    <Grid
{{- if .BtnBatchDelete }}
      @checkbox-change="onCheckboxChange"
      @checkbox-all="onCheckboxChange"
{{- end }}
    >
      <template #toolbar-tools>
        <div class="flex gap-2">
{{- if .BtnCreate }}
          <Button
            v-access:code="['{{.AuthPrefix}}:Create']"
            type="primary"
            @click="onCreate"
          >
            <Plus class="mr-1" />
            <span v-text="$t('common.actions.create')" />
          </Button>
{{- end }}
{{- if .BtnBatchDelete }}
          <Button
            v-access:code="['{{.AuthPrefix}}:BatchDelete']"
            danger
            :disabled="selectedRows.length === 0"
            @click="onBatchDelete"
          >
            <span v-text="$t('common.actions.delete')" />
          </Button>
{{- end }}
        </div>
      </template>
      <template #operation="{ row }">
{{- if .BtnEdit }}
        <Button type="link" v-text="$t('common.actions.edit')" @click="onEdit(row)" />
{{- end }}
{{- if .BtnDelete }}
        <Button type="link" danger v-text="$t('common.actions.delete')" @click="onDelete(row)" />
{{- end }}
      </template>
    </Grid>
    <Modal
      v-model:open="modalVisible"
      :title="isEdit ? $t('{{.I18nPrefix}}._edit') : $t('{{.I18nPrefix}}._create')"
      :confirm-loading="submitting"
      @ok="onSubmit"
    >
      <Form ref="formRef" :model="form" :rules="rules" :label-col="{ span: 5 }" :wrapper-col="{ span: 18 }">
{{- range .FormColumns}}
        <FormItem :label="$t('{{.I18nKey}}')" name="{{.JSONName}}">
          {{.VueFormControl}}
        </FormItem>
{{- end}}
      </Form>
    </Modal>
  </Page>
</template>
`

const vueFormTpl = `<script lang="ts" setup>
import type { {{.BizName}} } from '#/api/{{.ModuleName}}/{{.SnakeName}}';

import { reactive, ref } from 'vue';

import { Page } from '@vben/common-ui';

import {
  Button,
  Card,
  CheckboxGroup,
  DatePicker,
  Form,
  FormItem,
  Input,
  InputNumber,
  message,
  RadioGroup,
  Select,
  Switch,
  Textarea,
} from 'ant-design-vue';

import { create{{.BizName}}Api } from '#/api/{{.ModuleName}}/{{.SnakeName}}';
import { $t } from '#/locales';

defineOptions({ name: '{{.BizName}}Submit' });

const formRef = ref();
const submitting = ref(false);
const form = reactive<Partial<{{.BizName}}>>({
{{- range .FormColumns}}
  {{.JSONName}}: undefined,
{{- end}}
});

const rules: Record<string, any> = {
{{- range .FormColumns}}
{{- if .IsRequired}}
  {{.JSONName}}: [
    { required: true, message: $t('{{.I18nKey}}') + $t('lowcode.codegen.genRequired') },
  ],
{{- end}}
{{- end}}
};

async function onSubmit() {
  try {
    await formRef.value.validate();
  } catch {
    return;
  }
  submitting.value = true;
  try {
    await create{{.BizName}}Api(form);
    message.success('提交成功');
    formRef.value.resetFields();
  } finally {
    submitting.value = false;
  }
}
</script>

<template>
  <Page auto-content-height>
    <Card class="max-w-[720px]" :title="$t('{{.I18nPrefix}}._fill')">
      <Form ref="formRef" :model="form" :rules="rules" :label-col="{ span: 5 }" :wrapper-col="{ span: 14 }">
{{- range .FormColumns}}
        <FormItem :label="$t('{{.I18nKey}}')" name="{{.JSONName}}">
          {{.VueFormControl}}
        </FormItem>
{{- end}}
        <FormItem :wrapper-col="{ offset: 5, span: 14 }">
          <Button type="primary" :loading="submitting" @click="onSubmit">提交</Button>
        </FormItem>
      </Form>
    </Card>
  </Page>
</template>
`

// serviceTpl Kratos service 层：pb 请求 -> biz 用例（按按钮分支生成方法）
const serviceTpl = `package service

import (
	"context"

	adminV1 "github.com/antsurge/weaver-admin/api/gen/go/admin/service/v1"
	pb "github.com/antsurge/weaver-admin/api/gen/go/{{.ModuleName}}/service/v1"

	"github.com/antsurge/weaver-admin/app/admin/service/internal/biz"
	"github.com/antsurge/weaver-admin/pkg/enthelper"
	"github.com/go-kratos/kratos/v2/log"
)

// {{.BizName}}Service {{.Comment}}服务
type {{.BizName}}Service struct {
	adminV1.Unimplemented{{.BizName}}ServiceServer
	uc  *biz.{{.BizName}}Usecase
	log *log.Helper
}

func New{{.BizName}}Service(uc *biz.{{.BizName}}Usecase, logger log.Logger) *{{.BizName}}Service {
	return &{{.BizName}}Service{uc: uc, log: log.NewHelper(logger)}
}

{{- if .BtnList }}

func (s *{{.BizName}}Service) List{{.BizName}}(ctx context.Context, req *pb.List{{.BizName}}Request) (*pb.List{{.BizName}}Response, error) {
	res, err := s.uc.List(ctx, &biz.List{{.BizName}}Request{
		PaginationParams: enthelper.PaginationParams{CurrentPage: int(req.GetCurrentPage()), PageSize: int(req.GetPageSize())},
	})
	if err != nil {
		return nil, err
	}
	items := make([]*pb.{{.BizName}}, 0, res.Total)
	for _, v := range res.Data {
		items = append(items, &pb.{{.BizName}}{
			Id: v.ID,
{{- range .Columns}}
			{{.GoName}}: v.{{.GoName}},
{{- end}}
		})
	}
	return &pb.List{{.BizName}}Response{Items: items, Total: int64(res.Total)}, nil
}
{{- end }}

{{- if .BtnDetail }}

func (s *{{.BizName}}Service) Get{{.BizName}}(ctx context.Context, req *pb.Get{{.BizName}}Request) (*pb.Get{{.BizName}}Response, error) {
	item, err := s.uc.Get(ctx, req.GetId())
	if err != nil {
		return nil, err
	}
	return &pb.Get{{.BizName}}Response{Data: &pb.{{.BizName}}{
		Id: item.ID,
{{- range .Columns}}
		{{.GoName}}: item.{{.GoName}},
{{- end}}
	}}, nil
}
{{- end }}

{{- if .BtnCreate }}

func (s *{{.BizName}}Service) Create{{.BizName}}(ctx context.Context, req *pb.Create{{.BizName}}Request) (*pb.Create{{.BizName}}Response, error) {
	item := &biz.{{.BizName}}{
{{- range .FormColumns}}
		{{.GoName}}: req.Get{{.GoName}}(),
{{- end}}
	}
	created, err := s.uc.Create(ctx, item)
	if err != nil {
		return nil, err
	}
	return &pb.Create{{.BizName}}Response{Data: &pb.{{.BizName}}{Id: created.ID}}, nil
}
{{- end }}

{{- if .BtnEdit }}

func (s *{{.BizName}}Service) Update{{.BizName}}(ctx context.Context, req *pb.Update{{.BizName}}Request) (*pb.Update{{.BizName}}Response, error) {
	item := &biz.{{.BizName}}{
		ID: req.GetId(),
{{- range .FormColumns}}
		{{.GoName}}: req.Get{{.GoName}}(),
{{- end}}
	}
	updated, err := s.uc.Update(ctx, item)
	if err != nil {
		return nil, err
	}
	return &pb.Update{{.BizName}}Response{Data: &pb.{{.BizName}}{Id: updated.ID}}, nil
}
{{- end }}

{{- if or .BtnDelete .BtnBatchDelete }}

func (s *{{.BizName}}Service) Delete{{.BizName}}(ctx context.Context, req *pb.Delete{{.BizName}}Request) (*pb.Delete{{.BizName}}Response, error) {
	if err := s.uc.Delete(ctx, req.GetIds()); err != nil {
		return nil, err
	}
	return &pb.Delete{{.BizName}}Response{}, nil
}
{{- end }}

{{- if .BtnStatus }}

func (s *{{.BizName}}Service) Update{{.BizName}}Status(ctx context.Context, req *pb.Update{{.BizName}}StatusRequest) (*pb.Update{{.BizName}}StatusResponse, error) {
	if err := s.uc.UpdateStatus(ctx, req.GetId(), req.GetStatus()); err != nil {
		return nil, err
	}
	return &pb.Update{{.BizName}}StatusResponse{}, nil
}
{{- end }}
`

// handlerTpl 导入/导出自定义 handler（与 OrganizationHandler 模式一致）
const handlerTpl = `package handler

import (
	"errors"
	"io"
	nethttp "net/http"
	"strings"

	"github.com/antsurge/weaver-admin/app/admin/service/internal/biz"
	"github.com/go-kratos/kratos/v2/log"
	khttp "github.com/go-kratos/kratos/v2/transport/http"
	"github.com/xuri/excelize/v2"
)

// {{.BizName}}Handler {{.Comment}}导入导出处理器
type {{.BizName}}Handler struct {
	uc  *biz.{{.BizName}}Usecase
	log *log.Helper
}

func New{{.BizName}}Handler(uc *biz.{{.BizName}}Usecase, logger log.Logger) *{{.BizName}}Handler {
	return &{{.BizName}}Handler{uc: uc, log: log.NewHelper(logger)}
}

{{- if .BtnExport }}

// Export{{.BizName}} 导出{{.Comment}}（xlsx 二进制流）
func (h *{{.BizName}}Handler) Export{{.BizName}}(ctx khttp.Context) error {
	req := ctx.Request()
	if err := req.ParseForm(); err != nil {
		return err
	}
	pageSize := int64(0)
	list, err := h.uc.List(ctx, &biz.List{{.BizName}}Request{})
	_ = pageSize
	if err != nil {
		return err
	}

	f := excelize.NewFile()
	defer f.Close()
	sheet := "Sheet1"
	f.SetSheetName("Sheet1", sheet)
	var cell string
{{- range $j, $c := .ListColumns}}
	cell, _ = excelize.CoordinatesToCellName({{inc $j}}, 1)
	_ = f.SetCellValue(sheet, cell, "{{$c.Comment}}")
{{- end}}

	for i, v := range list.Data {
		row := i + 2
{{- range $j, $c := .ListColumns}}
		cell, _ = excelize.CoordinatesToCellName({{inc $j}}, row)
		_ = f.SetCellValue(sheet, cell, v.{{$c.GoName}})
{{- end}}
	}

	ctx.Response().Header().Set("Content-Type", "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet")
	ctx.Response().Header().Set("Content-Disposition", "attachment; filename="+"{{.SnakeName}}.xlsx")
	ctx.Response().WriteHeader(nethttp.StatusOK)
	_, err = f.WriteTo(ctx.Response())
	return err
}
{{- end }}

{{- if .BtnImport }}

// Import{{.BizName}} 导入{{.Comment}}（multipart/form-data 文件上传）
func (h *{{.BizName}}Handler) Import{{.BizName}}(ctx khttp.Context) error {
	req := ctx.Request()
	if err := req.ParseMultipartForm(32 << 20); err != nil {
		return errors.New("请上传文件")
	}
	file, _, err := req.FormFile("file")
	if err != nil {
		return errors.New("请上传文件")
	}
	defer file.Close()

	// 限制文件大小 10MB
	data, err := io.ReadAll(io.LimitReader(file, 10<<20))
	if err != nil {
		return errors.New("文件读取失败")
	}
	if len(data) == 0 {
		return errors.New("文件内容为空")
	}

	f, err := excelize.OpenReader(strings.NewReader(string(data)))
	if err != nil {
		return errors.New("文件解析失败，请上传正确的 Excel 文件")
	}
	defer f.Close()

	rows, err := f.GetRows("Sheet1")
	if err != nil || len(rows) < 2 {
		return errors.New("文件无有效数据")
	}

	count := 0
{{- if .HasFormString }}
	for _, row := range rows[1:] {
{{- else }}
	for idx := 1; idx < len(rows); idx++ {
		_ = idx
{{- end }}
		item := &biz.{{.BizName}}{}
{{- range $j, $c := .FormColumns}}
{{- if isString $c.GoType }}
		if len(row) > {{inc $j}} {
			item.{{$c.GoName}} = row[{{inc $j}}]
		}
{{- end}}
{{- end}}
		if _, err := h.uc.Create(ctx, item); err == nil {
			count++
		}
	}
	h.log.Infof("导入{{.Comment}}成功 %d 条", count)
	return nil
}
{{- end }}
`

// handlerPBTpl 导入导出路由注册（字面量路径，必须先于参数化路由注册）
const handlerPBTpl = `package pb

import (
	"context"

	"github.com/go-kratos/kratos/v2/transport/http"
	"github.com/go-kratos/kratos/v2/transport/http/binding"
)

var _ = new(context.Context)
var _ = binding.EncodeURL

const _ = http.SupportPackageIsVersion1

// {{.BizName}}HandlerHTTPServer {{.Comment}}导入导出 HTTP 处理器接口
type {{.BizName}}HandlerHTTPServer interface {
{{- if .BtnExport }}
	Export{{.BizName}}(http.Context) error
{{- end }}
{{- if .BtnImport }}
	Import{{.BizName}}(http.Context) error
{{- end }}
}

const (
{{- if .BtnExport }}
	Operation{{.BizName}}Export = "/{{.ModuleName}}.service.v1.{{.BizName}}/Export"
{{- end }}
{{- if .BtnImport }}
	Operation{{.BizName}}Import = "/{{.ModuleName}}.service.v1.{{.BizName}}/Import"
{{- end }}
)

// Register{{.BizName}}HandlerServer 注册{{.Comment}}导入导出路由。
// 注意：带字面量路径的路由（/import、/export）必须先于参数化路由（/{id}）注册。
func Register{{.BizName}}HandlerServer(s *http.Server, srv {{.BizName}}HandlerHTTPServer) {
	r := s.Route("/")
{{- if .BtnExport }}
	r.POST("admin/v1/{{.SnakeName}}s/export", _{{.BizName}}_Export0_HTTP_Handler(srv))
{{- end }}
{{- if .BtnImport }}
	r.POST("admin/v1/{{.SnakeName}}s/import", _{{.BizName}}_Import0_HTTP_Handler(srv))
{{- end }}
}

{{- if .BtnExport }}

func _{{.BizName}}_Export0_HTTP_Handler(srv {{.BizName}}HandlerHTTPServer) func(ctx http.Context) error {
	return func(ctx http.Context) error {
		h := ctx.Middleware(func(ctx1 context.Context, req interface{}) (interface{}, error) {
			mergeContext(ctx, ctx1)
			return nil, srv.Export{{.BizName}}(ctx)
		})
		http.SetOperation(ctx, Operation{{.BizName}}Export)
		_, err := h(ctx, nil)
		return err
	}
}
{{- end }}
{{- if .BtnImport }}

func _{{.BizName}}_Import0_HTTP_Handler(srv {{.BizName}}HandlerHTTPServer) func(ctx http.Context) error {
	return func(ctx http.Context) error {
		h := ctx.Middleware(func(ctx1 context.Context, req interface{}) (interface{}, error) {
			mergeContext(ctx, ctx1)
			return nil, srv.Import{{.BizName}}(ctx)
		})
		http.SetOperation(ctx, Operation{{.BizName}}Import)
		_, err := h(ctx, nil)
		return err
	}
}
{{- end }}
`

// menuSQLTpl 生成的菜单 SQL（目录/菜单/按钮 + 角色授权）
const menuSQLTpl = `-- {{.Comment}}菜单初始化 SQL
-- 说明：1) 幂等执行；2) 目录仅当 MenuModule={{.MenuModule}} 不存在时创建；3) 按钮按生成配置自动挂到菜单下。
-- 角色授权默认给 super_admin，如需其他角色请自行调整。

-- 1. 业务菜单（挂到 MenuModule={{.MenuModule}} 目录下；若目录不存在先创建）
INSERT INTO menu (id, parent_id, name, code, title, path, icon, type, component, weight, status, auth_code, created_at, updated_at)
VALUES (
  '{{.SnakeName}}_menu',
  {{.MenuModule}}_catalog_id,
  '{{.Comment}}',
  '{{.BizName}}',
  '{{.Comment}}',
  '/{{.MenuModule}}/{{.SnakeName}}',
  'lucide:box',
  'menu',
  '{{.MenuModule}}/{{.SnakeName}}/index',
  20,
  'enabled',
  '{{.AuthPrefix}}',
  NOW(6), NOW(6)
);

-- 2. 按钮（按生成配置）
{{- if .BtnList }}
INSERT INTO menu (id, parent_id, name, code, title, type, weight, status, auth_code, created_at, updated_at)
VALUES ('{{.SnakeName}}_btn_list', '{{.SnakeName}}_menu', '列表', 'List', '列表', 'action', 1, 'enabled', '{{.AuthPrefix}}:List', NOW(6), NOW(6));
{{- end }}
{{- if .BtnDetail }}
INSERT INTO menu (id, parent_id, name, code, title, type, weight, status, auth_code, created_at, updated_at)
VALUES ('{{.SnakeName}}_btn_detail', '{{.SnakeName}}_menu', '详情', 'Detail', '详情', 'action', 2, 'enabled', '{{.AuthPrefix}}:Info', NOW(6), NOW(6));
{{- end }}
{{- if .BtnCreate }}
INSERT INTO menu (id, parent_id, name, code, title, type, weight, status, auth_code, created_at, updated_at)
VALUES ('{{.SnakeName}}_btn_create', '{{.SnakeName}}_menu', '创建', 'Create', '创建', 'action', 3, 'enabled', '{{.AuthPrefix}}:Create', NOW(6), NOW(6));
{{- end }}
{{- if .BtnEdit }}
INSERT INTO menu (id, parent_id, name, code, title, type, weight, status, auth_code, created_at, updated_at)
VALUES ('{{.SnakeName}}_btn_edit', '{{.SnakeName}}_menu', '编辑', 'Edit', '编辑', 'action', 4, 'enabled', '{{.AuthPrefix}}:Edit', NOW(6), NOW(6));
{{- end }}
{{- if .BtnDelete }}
INSERT INTO menu (id, parent_id, name, code, title, type, weight, status, auth_code, created_at, updated_at)
VALUES ('{{.SnakeName}}_btn_delete', '{{.SnakeName}}_menu', '删除', 'Delete', '删除', 'action', 5, 'enabled', '{{.AuthPrefix}}:Delete', NOW(6), NOW(6));
{{- end }}
{{- if .BtnBatchDelete }}
INSERT INTO menu (id, parent_id, name, code, title, type, weight, status, auth_code, created_at, updated_at)
VALUES ('{{.SnakeName}}_btn_batch_delete', '{{.SnakeName}}_menu', '批量删除', 'BatchDelete', '批量删除', 'action', 6, 'enabled', '{{.AuthPrefix}}:BatchDelete', NOW(6), NOW(6));
{{- end }}
{{- if .BtnStatus }}
INSERT INTO menu (id, parent_id, name, code, title, type, weight, status, auth_code, created_at, updated_at)
VALUES ('{{.SnakeName}}_btn_status', '{{.SnakeName}}_menu', '状态', 'Status', '状态', 'action', 7, 'enabled', '{{.AuthPrefix}}:SwitchStatus', NOW(6), NOW(6));
{{- end }}
{{- if .BtnImport }}
INSERT INTO menu (id, parent_id, name, code, title, type, weight, status, auth_code, created_at, updated_at)
VALUES ('{{.SnakeName}}_btn_import', '{{.SnakeName}}_menu', '导入', 'Import', '导入', 'action', 8, 'enabled', '{{.AuthPrefix}}:Import', NOW(6), NOW(6));
{{- end }}
{{- if .BtnExport }}
INSERT INTO menu (id, parent_id, name, code, title, type, weight, status, auth_code, created_at, updated_at)
VALUES ('{{.SnakeName}}_btn_export', '{{.SnakeName}}_menu', '导出', 'Export', '导出', 'action', 9, 'enabled', '{{.AuthPrefix}}:Export', NOW(6), NOW(6));
{{- end }}

-- 3. 授权给超管角色（幂等）
INSERT IGNORE INTO role_menu (id, role_id, menu_id, created_at)
SELECT '{{.SnakeName}}_rm_super', r.id, '{{.SnakeName}}_menu', NOW(6)
FROM role r WHERE r.is_super_admin = 1
  AND NOT EXISTS (SELECT 1 FROM role_menu rm WHERE rm.role_id = r.id AND rm.menu_id = '{{.SnakeName}}_menu');
`

// menuJSONTpl 生成的菜单树配置（供一键落库接口 /admin/v1/codegen/apply-menu 使用）
const menuJSONTpl = `{
  "genTableId": "{{.ID}}",
  "menuModule": "{{.MenuModule}}",
  "authPrefix": "{{.AuthPrefix}}",
  "menus": [
    {
      "name": "{{.Comment}}",
      "type": "menu",
      "path": "/{{.MenuModule}}/{{.SnakeName}}",
      "component": "{{.MenuModule}}/{{.SnakeName}}/index",
      "authCode": "{{.AuthPrefix}}",
      "weight": 20,
      "actions": [
{{- if .BtnList }}
        { "name": "列表", "code": "List", "authCode": "{{.AuthPrefix}}:List" },
{{- end }}
{{- if .BtnDetail }}
        { "name": "详情", "code": "Detail", "authCode": "{{.AuthPrefix}}:Info" },
{{- end }}
{{- if .BtnCreate }}
        { "name": "创建", "code": "Create", "authCode": "{{.AuthPrefix}}:Create" },
{{- end }}
{{- if .BtnEdit }}
        { "name": "编辑", "code": "Edit", "authCode": "{{.AuthPrefix}}:Edit" },
{{- end }}
{{- if .BtnDelete }}
        { "name": "删除", "code": "Delete", "authCode": "{{.AuthPrefix}}:Delete" },
{{- end }}
{{- if .BtnBatchDelete }}
        { "name": "批量删除", "code": "BatchDelete", "authCode": "{{.AuthPrefix}}:BatchDelete" },
{{- end }}
{{- if .BtnStatus }}
        { "name": "状态", "code": "Status", "authCode": "{{.AuthPrefix}}:SwitchStatus" },
{{- end }}
{{- if .BtnImport }}
        { "name": "导入", "code": "Import", "authCode": "{{.AuthPrefix}}:Import" },
{{- end }}
{{- if .BtnExport }}
        { "name": "导出", "code": "Export", "authCode": "{{.AuthPrefix}}:Export" },
{{- end }}
      ]
    }
  ]
}
`
