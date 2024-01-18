package excel

import (
	"github.com/xuri/excelize/v2"
	"strconv"
	"time"
)

// ExcelDateToDate excel读取到的日期转换为time类型
func ExcelDateToDate(excelDate string) time.Time {
	excelTime := time.Date(1899, time.December, 30, 0, 0, 0, 0, time.UTC)
	var days, _ = strconv.Atoi(excelDate)
	return excelTime.Add(time.Second * time.Duration(days*86400))
}

// CreateXls 导出为excel文件 *.xlsx
// filename 导出的文件名
// titles 表头信息
// data 每行的数据信息
func CreateXls(filename string, titles []any, data [][]any) error {
	f := excelize.NewFile()
	defer f.Close()

	//流式写入器
	sw, err := f.NewStreamWriter("Sheet1")
	if err != nil {
		return err
	}

	//行ID
	rowID := 1

	//设置第一行
	if len(titles) > 0 {
		_ = sw.SetRow("A1", titles)
		rowID = 2
	}

	if len(data) > 0 {
		for _, v := range data {
			cell, err := excelize.CoordinatesToCellName(1, rowID)
			if err != nil {
				return err
			}

			if err := sw.SetRow(cell, v); err != nil {
				return err
			}

			rowID++
		}
	}

	if err = sw.Flush(); err != nil {
		return err
	}
	if err = f.SaveAs(filename); err != nil {
		return err
	}

	return err
}

// ReadXls 读取excel文件中的数据 仅读取第一个工作表
func ReadXls(filename string) ([][]string, error) {
	//读取excel数据
	f, err := excelize.OpenFile(filename, excelize.Options{})
	if err != nil {
		return nil, err
	}

	//按索引获取工作表名 0默认为Sheet1
	name := f.GetSheetName(0)

	data := make([][]string, 0)

	//按行获取全部单元格的值
	rows, _ := f.GetRows(name)
	for _, row := range rows {
		d := make([]string, 0)
		for _, colCell := range row {
			d = append(d, colCell)
		}
		data = append(data, d)
	}

	return data, err
}
