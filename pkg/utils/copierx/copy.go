package copierx

import (
	"errors"
	"time"

	"github.com/jinzhu/copier"
	"google.golang.org/protobuf/types/known/structpb"
	"google.golang.org/protobuf/types/known/timestamppb"
)

var defaultConverters = []copier.TypeConverter{
	{
		// 支持 map[string]any -> *structpb.Struct（如代码生成器 GenTable.FieldsJSON）。
		// 缺少该转换器时 jinzhu/copier 无法转换类型，导致列表/详情返回的 fieldsJson 为空，
		// 前端复制 CRUD 记录时字段带不过来。
		SrcType: map[string]any{},
		DstType: &structpb.Struct{},
		Fn: func(src interface{}) (interface{}, error) {
			m, ok := src.(map[string]any)
			if !ok || len(m) == 0 {
				return nil, nil
			}
			st, err := structpb.NewStruct(m)
			if err != nil {
				return nil, errors.New("copy map to structpb.Struct: " + err.Error())
			}
			return st, nil
		},
	},
	{
		SrcType: time.Time{},
		DstType: &timestamppb.Timestamp{},
		Fn: func(src interface{}) (interface{}, error) {
			t := src.(time.Time)
			return timestamppb.New(t), nil
		},
	},
	{
		// 支持 *time.Time -> *timestamppb.Timestamp（含 nil 指针）
		SrcType: &time.Time{},
		DstType: &timestamppb.Timestamp{},
		Fn: func(src interface{}) (interface{}, error) {
			t := src.(*time.Time)
			if t == nil {
				return nil, nil
			}
			return timestamppb.New(*t), nil
		},
	},
}

func Copy(dst interface{}, src interface{}) error {
	return copier.CopyWithOption(dst, src, copier.Option{
		Converters: defaultConverters,
	})
}
