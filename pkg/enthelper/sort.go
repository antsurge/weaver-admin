package enthelper

import commonV1 "github.com/antsurge/weaver-admin/api/gen/go/common/v1"

type Sort struct {
	Field string `json:"field" form:"field" query:"field"`
	Order string `json:"order" form:"order" query:"order"`
}

type OrderFunc[Q any] func(Q, string) Q

func ApplySorts[Q any](q Q, sorts []Sort, fieldMap map[string]OrderFunc[Q]) Q {
	for _, s := range sorts {
		if fn, ok := fieldMap[s.Field]; ok {
			q = fn(q, s.Field)
		}
	}

	return q
}

func ConvertSorts(src []*commonV1.Sort) []Sort {
	if len(src) == 0 {
		return nil
	}

	dst := make([]Sort, 0, len(src))

	for _, s := range src {
		if s == nil {
			continue
		}

		dst = append(dst, Sort{
			Field: s.Field,
			Order: s.Order,
		})
	}

	return dst
}
