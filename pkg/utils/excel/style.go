package excel

import (
	"math"

	"github.com/xuri/excelize/v2"
)

// AlignmentType 对齐方式
type AlignmentType string

const (
	AlignCenter AlignmentType = "center"
	AlignLeft   AlignmentType = "left"
	AlignRight  AlignmentType = "right"
)

// AutoAdjustColumnWidth 根据内容自动调整列宽
// f: excelize 文件
// sheet: sheet 名称
// startRow: 开始行（通常表头行）
// endRow: 结束行
func AutoAdjustColumnWidth(f *excelize.File, sheet string, startRow, endRow int) error {
	if f == nil {
		return nil
	}

	// 获取 sheet 所有列数据
	cols, err := f.GetCols(sheet)
	if err != nil {
		return err
	}

	for colIdx, col := range cols {
		maxLen := 0.0
		for rowIdx, cellVal := range col {
			rowNumber := rowIdx + 1
			if rowNumber < startRow || (endRow > 0 && rowNumber > endRow) {
				continue
			}
			// 简单计算宽度：中文算 2，英文算 1
			width := float64(len(cellVal))
			for _, c := range cellVal {
				if c > 127 {
					width += 1.0 // 中文额外增加
				}
			}
			if width > maxLen {
				maxLen = width
			}
		}

		if maxLen > 0 {
			colLetter := string(rune('A' + colIdx))
			// 加一点额外空间
			if err := f.SetColWidth(sheet, colLetter, colLetter, math.Ceil(maxLen)+2); err != nil {
				return err
			}
		}
	}

	return nil
}

// NewHeaderStyle 创建表头样式
// f: excelize 文件
// align: 对齐方式，默认居中
// bold: 是否加粗，默认 true
func NewHeaderStyle(f *excelize.File, align AlignmentType, bold bool) (int, error) {
	if f == nil {
		return 0, nil
	}

	if align == "" {
		align = AlignCenter
	}

	styleID, err := f.NewStyle(&excelize.Style{
		Font: &excelize.Font{
			Bold: bold,
		},
		Alignment: &excelize.Alignment{
			Horizontal: string(align),
			Vertical:   "center",
		},
	})
	if err != nil {
		return 0, err
	}

	return styleID, nil
}

// NewContentStyle 创建内容样式
// f: excelize 文件
// align: 对齐方式，默认居中
func NewContentStyle(f *excelize.File, align AlignmentType) (int, error) {
	if f == nil {
		return 0, nil
	}

	if align == "" {
		align = AlignCenter
	}

	styleID, err := f.NewStyle(&excelize.Style{
		Alignment: &excelize.Alignment{
			Horizontal: string(align),
			Vertical:   "center",
		},
	})
	if err != nil {
		return 0, err
	}

	return styleID, nil
}
