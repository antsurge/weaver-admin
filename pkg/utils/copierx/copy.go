package copierx

import (
	"time"

	"github.com/jinzhu/copier"
	"google.golang.org/protobuf/types/known/timestamppb"
)

var defaultConverters = []copier.TypeConverter{
	{
		SrcType: time.Time{},
		DstType: &timestamppb.Timestamp{},
		Fn: func(src interface{}) (interface{}, error) {
			t := src.(time.Time)
			return timestamppb.New(t), nil
		},
	},
}

func Copy(dst interface{}, src interface{}) error {
	return copier.CopyWithOption(dst, src, copier.Option{
		Converters: defaultConverters,
	})
}
