// Package openapi_scanner 提供从 openapi.yaml 解析接口元数据的能力。
// 启动时调用一次 Scan() 缓存所有接口，后续通过 Metadata() 获取。
package openapi_scanner

import (
	"os"
	"path/filepath"
	"sort"
	"strings"
	"sync"

	"gopkg.in/yaml.v3"
)

// HTTP method 常量
var httpMethods = []string{"get", "post", "put", "delete", "patch", "head", "options"}

// Endpoint 单个接口
type Endpoint struct {
	Method  string `json:"method"`
	Path    string `json:"path"`
	Summary string `json:"summary"`
}

// ServiceGroup 按 service 分组的接口集合
type ServiceGroup struct {
	Service   string     `json:"service"`   // 服务名，如 "PermissionService"
	Tag       string     `json:"tag"`       // OpenAPI tag
	Endpoints []Endpoint `json:"endpoints"`
}

// Service 全局扫描器（单例）
type Service struct {
	mu       sync.RWMutex
	groups   []ServiceGroup
	pathByOP map[string]string // operationId -> "service|method|path"
}

// New 创建扫描器实例（内部使用）
func New() *Service {
	return &Service{
		pathByOP: make(map[string]string),
	}
}

// Scan 解析 openapi.yaml。
// openapiPath 一般指向项目内的 api/gen/openapi/openapi.yaml
func (s *Service) Scan(openapiPath string) error {
	data, err := os.ReadFile(openapiPath)
	if err != nil {
		// 尝试相对项目根目录解析
		data, err = os.ReadFile(filepath.Join(".", openapiPath))
		if err != nil {
			return err
		}
	}

	doc := struct {
		Paths map[string]map[string]struct {
			Tags        []string `yaml:"tags"`
			Summary     string   `yaml:"summary"`
			OperationID string   `yaml:"operationId"`
		} `yaml:"paths"`
	}{}

	if err := yaml.Unmarshal(data, &doc); err != nil {
		return err
	}

	// 按 service 聚合
	groups := make(map[string]*ServiceGroup)

	for path, methods := range doc.Paths {
		// 按 HTTP method 顺序处理
		for _, method := range httpMethods {
			op, ok := methods[method]
			if !ok {
				continue
			}
			svcName, tag := parseOperation(op.OperationID, op.Tags, method)
			group, ok := groups[svcName]
			if !ok {
				group = &ServiceGroup{Service: svcName, Tag: tag}
				groups[svcName] = group
			}
			group.Endpoints = append(group.Endpoints, Endpoint{
				Method:  strings.ToUpper(method),
				Path:    path,
				Summary: op.Summary,
			})
		}
	}

	// 转 slice + 排序
	out := make([]ServiceGroup, 0, len(groups))
	for _, g := range groups {
		sort.SliceStable(g.Endpoints, func(i, j int) bool {
			if g.Endpoints[i].Path != g.Endpoints[j].Path {
				return g.Endpoints[i].Path < g.Endpoints[j].Path
			}
			return g.Endpoints[i].Method < g.Endpoints[j].Method
		})
		out = append(out, *g)
	}
	sort.SliceStable(out, func(i, j int) bool {
		return out[i].Service < out[j].Service
	})

	s.mu.Lock()
	s.groups = out
	s.mu.Unlock()
	return nil
}

// Metadata 返回所有接口分组
func (s *Service) Metadata() []ServiceGroup {
	s.mu.RLock()
	defer s.mu.RUnlock()
	out := make([]ServiceGroup, len(s.groups))
	copy(out, s.groups)
	return out
}

// parseOperation 从 operationId 提取 service 名（如 "Permission_ListMenu" -> "PermissionService"）
func parseOperation(operationID string, tags []string, method string) (string, string) {
	tag := ""
	if len(tags) > 0 {
		tag = tags[0]
	}
	if operationID == "" {
		return fallbackName(method, tag), tag
	}
	// 形如 "Identity_ListAdmin" / "Authentication_Login" / "Permission_MenuTree"
	idx := strings.Index(operationID, "_")
	if idx <= 0 {
		return operationID + "Service", tag
	}
	prefix := operationID[:idx]
	// 已含 Service 后缀不再加
	if strings.HasSuffix(prefix, "Service") {
		return prefix, tag
	}
	return prefix + "Service", tag
}

func fallbackName(method, tag string) string {
	if tag != "" {
		return tag + "Service"
	}
	return strings.ToUpper(method) + "Service"
}