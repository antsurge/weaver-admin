package handler

import (
	"archive/zip"
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"net/url"
	"strings"
	"time"

	"github.com/antsurge/weaver-admin/app/admin/service/internal/biz"
	"github.com/go-kratos/kratos/v2/log"
	"github.com/go-kratos/kratos/v2/transport/http"
)

type CodegenHandler struct {
	codegenUc *biz.CodegenUsecase
	menuUc    *biz.MenuUsecase
	log       *log.Helper
}

func NewCodegenHandler(codegenUc *biz.CodegenUsecase, menuUc *biz.MenuUsecase, logger log.Logger) *CodegenHandler {
	return &CodegenHandler{
		codegenUc: codegenUc,
		menuUc:    menuUc,
		log:       log.NewHelper(logger),
	}
}

// DownloadCode 根据生成配置 ID 生成代码并打包 zip 下载
func (h *CodegenHandler) DownloadCode(ctx http.Context) error {
	var req struct {
		ID string `json:"id"`
	}
	body := ctx.Request().Body
	if body == nil {
		return errors.New("请求体不能为空")
	}
	defer body.Close()
	if err := json.NewDecoder(body).Decode(&req); err != nil {
		return err
	}
	if req.ID == "" {
		return errors.New("id 不能为空")
	}

	files, err := h.codegenUc.GenerateCode(ctx, &biz.GenerateCodeInput{ID: req.ID})
	if err != nil {
		return err
	}
	if len(files) == 0 {
		return errors.New("未生成任何代码文件")
	}

	// 打包 zip
	var buf bytes.Buffer
	zw := zip.NewWriter(&buf)
	for _, f := range files {
		w, err := zw.Create(f.FileName)
		if err != nil {
			return err
		}
		if _, err := w.Write([]byte(f.Content)); err != nil {
			return err
		}
	}
	if err := zw.Close(); err != nil {
		return err
	}

	fileName := fmt.Sprintf("code_%s.zip", time.Now().Format("20060102150405"))
	ctx.Response().Header().Set("Content-Type", "application/zip")
	ctx.Response().Header().Set("Content-Disposition", "attachment; filename*=UTF-8''"+url.QueryEscape(fileName))
	ctx.Response().Header().Set("Content-Length", fmt.Sprintf("%d", buf.Len()))

	_, err = ctx.Response().Write(buf.Bytes())
	return err
}

// ApplyMenu 将生成配置一键落库为菜单树（目录 -> 菜单 -> 按钮）。
// 复用菜单创建的领域方法（MenuUsecase.CreateMenu），保证与菜单管理一致。
func (h *CodegenHandler) ApplyMenu(ctx http.Context) error {
	var req struct {
		ID string `json:"id"`
	}
	body := ctx.Request().Body
	if body == nil {
		return errors.New("请求体不能为空")
	}
	defer body.Close()
	if err := json.NewDecoder(body).Decode(&req); err != nil {
		return err
	}
	if req.ID == "" {
		return errors.New("id 不能为空")
	}

	// 读取生成配置
	gen, err := h.codegenUc.GetGenTable(ctx, req.ID)
	if err != nil {
		return err
	}
	if gen == nil {
		return errors.New("生成配置不存在")
	}

	// 菜单所属模块（父目录归属）
	menuModule := gen.MenuModule
	if menuModule == "" {
		menuModule = gen.ModuleName
	}
	// 权限码前缀（模块:业务，如 Order:OrderCenter），模块名与业务名统一走 ToAuthCode 规范化为首字母大写驼峰
	authPrefix := biz.ToAuthCode(menuModule) + ":" + biz.ToAuthCode(gen.BizName)
	// 按钮列表（未配置默认全选）
	buttons := gen.Buttons
	if len(buttons) == 0 {
		buttons = []string{
			biz.GenBtnList, biz.GenBtnDetail, biz.GenBtnCreate, biz.GenBtnEdit,
			biz.GenBtnDelete, biz.GenBtnBatchDelete, biz.GenBtnStatus, biz.GenBtnImport, biz.GenBtnExport,
		}
	}

	// 1. 查找或创建目录（模块级 catalog 菜单，parentID 为空）
	catalog, err := h.findOrCreateCatalog(ctx, menuModule)
	if err != nil {
		return err
	}

	// 2. 创建业务菜单（type=menu），幂等：同 authCode 已存在则复用
	menu, err := h.findOrCreateMenu(ctx, &biz.Menu{
		ParentID:  catalog.ID,
		Name:      gen.TableComment,
		Code:      gen.BizName,
		Title:     gen.TableComment,
		Type:      "menu",
		Path:      "/" + menuModule + "/" + ToLowerSnake(gen.BizName),
		Icon:      "lucide:box",
		Component: menuModule + "/" + ToLowerSnake(gen.BizName) + "/index",
		Weight:    20,
		Status:    "enabled",
		AuthCode:  authPrefix,
	})
	if err != nil {
		return err
	}

	// 3. 创建按钮（type=action），按 buttons 配置分支生成
	actions := buildMenuActions(authPrefix, buttons)
	for _, act := range actions {
		act.ParentID = menu.ID
		if _, err := h.findOrCreateMenu(ctx, act); err != nil {
			return err
		}
	}

	_, err = ctx.Response().Write([]byte(`{"message":"菜单生成成功"}`))
	return err
}

// DeleteGenerated 彻底删除生成配置：删除写盘的代码文件、删除真实数据表、
// 删除已落库的菜单树（目录 -> 菜单 -> 按钮）、物理删除配置记录。
func (h *CodegenHandler) DeleteGenerated(ctx http.Context) error {
	var req struct {
		IDs []string `json:"ids"`
	}
	body := ctx.Request().Body
	if body == nil {
		return errors.New("请求体不能为空")
	}
	defer body.Close()
	if err := json.NewDecoder(body).Decode(&req); err != nil {
		return err
	}
	if len(req.IDs) == 0 {
		return errors.New("请至少选择一条记录")
	}

	result, err := h.codegenUc.DeleteGenerated(ctx, req.IDs)
	if err != nil {
		return err
	}

	// 删除已落库的菜单树（目录 -> 菜单 -> 按钮）
	for _, g := range result.DeletedConfigs {
		if !g.MenuEnabled {
			continue
		}
		menuModule := g.MenuModule
		if menuModule == "" {
			menuModule = g.ModuleName
		}
		authPrefix := biz.ToAuthCode(menuModule) + ":" + biz.ToAuthCode(g.BizName)
		if err := h.removeMenuTree(ctx, authPrefix); err != nil {
			h.log.Errorf("codegen remove menu(%s) failed: %v", authPrefix, err)
		}
	}

	resp, _ := json.Marshal(map[string]any{
		"message":       "删除成功",
		"removedFiles":  result.RemovedFiles,
		"droppedTables": result.DroppedTables,
	})
	_, err = ctx.Response().Write(resp)
	return err
}

// removeMenuTree 根据权限码前缀删除菜单树（含全部子孙按钮节点）。
// 菜单归属的模块目录若已无子菜单则一并删除。
func (h *CodegenHandler) removeMenuTree(ctx http.Context, authPrefix string) error {
	tree, err := h.menuUc.MenuTree(ctx, &biz.ListMenuRequest{})
	if err != nil {
		return err
	}

	// 找 authCode 匹配的业务菜单节点，收集其全部子孙 id
	// visited 防环：菜单数据若存在循环 parent 引用，直接递归会栈溢出
	visited := make(map[string]bool)
	var collectIDs func(nodes []*biz.Menu) ([]string, *biz.Menu)
	collectIDs = func(nodes []*biz.Menu) ([]string, *biz.Menu) {
		for _, n := range nodes {
			if n.ID != "" && visited[n.ID] {
				continue
			}
			visited[n.ID] = true
			if n.AuthCode == authPrefix {
				ids := []string{n.ID}
				ids = append(ids, collectChildrenIDs(n.Children)...)
				return ids, n
			}
			if ids, found := collectIDs(n.Children); found != nil {
				return ids, found
			}
		}
		return nil, nil
	}

	ids, node := collectIDs(tree)
	if node == nil || len(ids) == 0 {
		return nil
	}
	if err := h.menuUc.DeleteMenu(ctx, ids); err != nil {
		return err
	}

	// 若父目录（模块 catalog）已无其他子菜单，则删除目录
	if node.ParentID != "" {
		parentID := node.ParentID
		if err := h.removeCatalogIfEmpty(ctx, parentID); err != nil {
			h.log.Warnf("codegen clean catalog(%s) failed: %v", parentID, err)
		}
	}
	return nil
}

// collectChildrenIDs 递归收集子孙节点 id（visited 防环，防止循环引用导致栈溢出）
func collectChildrenIDs(children []*biz.Menu) []string {
	visited := make(map[string]bool)
	var collect func(nodes []*biz.Menu) []string
	collect = func(nodes []*biz.Menu) []string {
		var ids []string
		for _, c := range nodes {
			if c.ID != "" && visited[c.ID] {
				continue
			}
			visited[c.ID] = true
			ids = append(ids, c.ID)
			ids = append(ids, collect(c.Children)...)
		}
		return ids
	}
	return collect(children)
}

// removeCatalogIfEmpty 若目录节点已无子菜单则删除之
func (h *CodegenHandler) removeCatalogIfEmpty(ctx http.Context, catalogID string) error {
	tree, err := h.menuUc.MenuTree(ctx, &biz.ListMenuRequest{})
	if err != nil {
		return err
	}
	// visited 防环：菜单数据若存在循环 parent 引用，直接递归会栈溢出
	visited := make(map[string]bool)
	var find func(nodes []*biz.Menu) *biz.Menu
	find = func(nodes []*biz.Menu) *biz.Menu {
		for _, n := range nodes {
			if n.ID != "" && visited[n.ID] {
				continue
			}
			visited[n.ID] = true
			if n.ID == catalogID {
				return n
			}
			if c := find(n.Children); c != nil {
				return c
			}
		}
		return nil
	}
	catalog := find(tree)
	if catalog == nil {
		return nil
	}
	if len(catalog.Children) > 0 {
		return nil
	}
	return h.menuUc.DeleteMenu(ctx, []string{catalog.ID})
}

// findOrCreateCatalog 查找模块目录（type=catalog，authCode 与模块同名）；不存在则创建
func (h *CodegenHandler) findOrCreateCatalog(ctx http.Context, menuModule string) (*biz.Menu, error) {
	catalogCode := biz.ToAuthCode(menuModule)
	tree, err := h.menuUc.MenuTree(ctx, &biz.ListMenuRequest{Type: "catalog"})
	if err != nil {
		return nil, err
	}
	for _, root := range tree {
		if root.AuthCode == catalogCode || root.Code == catalogCode {
			return root, nil
		}
	}
	// 创建目录
	catalog, err := h.menuUc.CreateMenu(ctx, &biz.Menu{
		Name:     catalogCode,
		Code:     catalogCode,
		Title:    catalogCode,
		Type:     "catalog",
		Path:     "/" + menuModule,
		Icon:     "lucide:layout-grid",
		Weight:   20,
		Status:   "enabled",
		AuthCode: catalogCode,
	})
	if err != nil {
		return nil, err
	}
	return catalog, nil
}

// findOrCreateMenu 按 authCode 幂等查找；不存在则创建
func (h *CodegenHandler) findOrCreateMenu(ctx http.Context, m *biz.Menu) (*biz.Menu, error) {
	tree, err := h.menuUc.MenuTree(ctx, &biz.ListMenuRequest{})
	if err != nil {
		return nil, err
	}
	// visited 防环：菜单数据若存在循环 parent 引用，直接递归会栈溢出
	visited := make(map[string]bool)
	var find func(nodes []*biz.Menu) *biz.Menu
	find = func(nodes []*biz.Menu) *biz.Menu {
		for _, n := range nodes {
			if n.ID != "" && visited[n.ID] {
				continue
			}
			visited[n.ID] = true
			if n.AuthCode == m.AuthCode {
				return n
			}
			if child := find(n.Children); child != nil {
				return child
			}
		}
		return nil
	}
	if found := find(tree); found != nil {
		return found, nil
	}
	return h.menuUc.CreateMenu(ctx, m)
}

// buildMenuActions 根据按钮配置构建按钮菜单列表
func buildMenuActions(authPrefix string, buttons []string) []*biz.Menu {
	btnMap := map[string]*biz.Menu{
		biz.GenBtnList:        {Name: "列表", Code: "List", Type: "action", Status: "enabled", AuthCode: authPrefix + ":List", Weight: 1},
		biz.GenBtnDetail:      {Name: "详情", Code: "Detail", Type: "action", Status: "enabled", AuthCode: authPrefix + ":Info", Weight: 2},
		biz.GenBtnCreate:      {Name: "创建", Code: "Create", Type: "action", Status: "enabled", AuthCode: authPrefix + ":Create", Weight: 3},
		biz.GenBtnEdit:        {Name: "编辑", Code: "Edit", Type: "action", Status: "enabled", AuthCode: authPrefix + ":Edit", Weight: 4},
		biz.GenBtnDelete:      {Name: "删除", Code: "Delete", Type: "action", Status: "enabled", AuthCode: authPrefix + ":Delete", Weight: 5},
		biz.GenBtnBatchDelete: {Name: "批量删除", Code: "BatchDelete", Type: "action", Status: "enabled", AuthCode: authPrefix + ":BatchDelete", Weight: 6},
		biz.GenBtnStatus:      {Name: "状态", Code: "Status", Type: "action", Status: "enabled", AuthCode: authPrefix + ":SwitchStatus", Weight: 7},
		biz.GenBtnImport:      {Name: "导入", Code: "Import", Type: "action", Status: "enabled", AuthCode: authPrefix + ":Import", Weight: 8},
		biz.GenBtnExport:      {Name: "导出", Code: "Export", Type: "action", Status: "enabled", AuthCode: authPrefix + ":Export", Weight: 9},
	}
	actions := make([]*biz.Menu, 0, len(buttons))
	for _, b := range buttons {
		if m, ok := btnMap[b]; ok {
			m.Title = m.Name
			actions = append(actions, m)
		}
	}
	return actions
}

// ToMenuCamel 模块名转首字母大写的驼峰（catalog/权限码前缀，如 lowcode -> Lowcode）
func ToMenuCamel(s string) string {
	if s == "" {
		return s
	}
	s = strings.Trim(strings.TrimSpace(s), "_")
	parts := strings.FieldsFunc(s, func(r rune) bool { return r == '_' || r == '-' })
	var sb strings.Builder
	for _, p := range parts {
		if p == "" {
			continue
		}
		sb.WriteString(strings.ToUpper(p[:1]))
		sb.WriteString(p[1:])
	}
	return sb.String()
}

// ToLowerSnake 业务名转小写下划线（如 JobLog -> job_log）
func ToLowerSnake(s string) string {
	var sb strings.Builder
	runes := []rune(s)
	for i, r := range runes {
		if i > 0 && r >= 'A' && r <= 'Z' {
			sb.WriteByte('_')
		}
		sb.WriteRune(r)
	}
	return strings.ToLower(sb.String())
}
