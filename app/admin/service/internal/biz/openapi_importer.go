package biz

import (
	"fmt"
	"sort"
	"strings"

	"gopkg.in/yaml.v3"
)

// httpMethods 按顺序处理的 HTTP 方法
var httpMethods = []string{"get", "post", "put", "delete", "patch", "head", "options"}

// ParseOpenAPIFile 解析 openapi.yaml 文件内容，返回 ApiInterface 列表
func ParseOpenAPIFile(data []byte) ([]*ApiInterface, error) {
	doc := struct {
		Paths map[string]map[string]struct {
			Tags        []string `yaml:"tags"`
			Summary     string   `yaml:"summary"`
			OperationID string   `yaml:"operationId"`
		} `yaml:"paths"`
	}{}

	if err := yaml.Unmarshal(data, &doc); err != nil {
		return nil, fmt.Errorf("parse openapi yaml: %w", err)
	}

	if len(doc.Paths) == 0 {
		return nil, fmt.Errorf("no paths found in openapi.yaml")
	}

	var items []*ApiInterface

	for path, methods := range doc.Paths {
		for _, method := range httpMethods {
			op, ok := methods[method]
			if !ok {
				continue
			}
			svcName, tag := parseImportOperation(op.OperationID, op.Tags, method)
			code := fmt.Sprintf("%s|%s|%s", svcName, strings.ToUpper(method), path)
			items = append(items, &ApiInterface{
				Service: svcName,
				Tag:     tag,
				Method:  strings.ToUpper(method),
				Path:    path,
				Summary: op.Summary,
				Code:    code,
			})
		}
	}

	// 排序，保证导入结果稳定
	sort.SliceStable(items, func(i, j int) bool {
		if items[i].Service != items[j].Service {
			return items[i].Service < items[j].Service
		}
		if items[i].Path != items[j].Path {
			return items[i].Path < items[j].Path
		}
		return items[i].Method < items[j].Method
	})

	return items, nil
}

// parseImportOperation 从 operationId 提取 service 名（如 "Permission_ListMenu" -> "PermissionService"）
func parseImportOperation(operationID string, tags []string, method string) (string, string) {
	tag := ""
	if len(tags) > 0 {
		tag = tags[0]
	}
	if operationID == "" {
		return importFallbackName(method, tag), tag
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

func importFallbackName(method, tag string) string {
	if tag != "" {
		return tag + "Service"
	}
	return strings.ToUpper(method) + "Service"
}