package enthelper

type QueryOption struct {
	IncludeDeleted bool   // 是否包含已删除
	OnlyDeleted    bool   // 只查已删除（回收站）
	Sorts          []Sort // 排序
}
