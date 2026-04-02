package excel

import "github.com/xuri/excelize/v2"

// SheetTemplateOptions 可选参数
type SheetTemplateOptions struct {
	ColCount int // 合并列数，默认10
}

func ApplySheetTemaplte(f *excelize.File, sheetName, title, exporter string, opts *SheetTemplateOptions) error {
	// 设置合并列数
	colCount := 10
	if opts != nil && opts.ColCount > 0 {
		colCount = opts.ColCount
	}

	endCol := string(rune('A' + colCount - 1))

	// ----------------------
	// 第一行：标题，合并居中
	// ----------------------
	if err := f.MergeCell(sheetName, "A1", endCol+"1"); err != nil {
		return err
	}
	f.SetCellValue(sheetName, "A1", title)

	titleStyle, _ := f.NewStyle(&excelize.Style{
		Alignment: &excelize.Alignment{
			Horizontal: "center",
			Vertical:   "center",
		},
		Font: &excelize.Font{
			Bold: true,
			Size: 12,
		},
	})
	f.SetCellStyle(sheetName, "A1", endCol+"1", titleStyle)

	// ----------------------
	// 第二行：导出人，合并右对齐
	// ----------------------
	if err := f.MergeCell(sheetName, "A2", endCol+"2"); err != nil {
		return err
	}
	f.SetCellValue(sheetName, "A2", "导出人："+exporter)

	exporterStyle, _ := f.NewStyle(&excelize.Style{
		Alignment: &excelize.Alignment{
			Horizontal: "right",
			Vertical:   "center",
		},
		Font: &excelize.Font{
			Italic: true,
		},
	})
	f.SetCellStyle(sheetName, "A2", endCol+"2", exporterStyle)

	return nil
}
