package exporter

import (
	"os"
	"reflect"
	"strings"
)

type CsvExporter struct {
	CsvFilePath string
	ExportHeads []string
	SplitSign string
	csvFile   *os.File
}

// Init 初始化 CsvExporter
func (csvExporter *CsvExporter) Init() error {
	if csvExporter.CsvFilePath == "" && csvExporter.csvFile == nil {
		panic("CsvFilePath is not assigned")
	}
	if csvExporter.SplitSign == "" {
		csvExporter.SplitSign = ","
	}
	fp, err := os.Create(csvExporter.CsvFilePath)
	if err != nil {
		return err
	}
	if len(csvExporter.ExportHeads) > 0 {
		_, err = fp.WriteString(strings.Join(csvExporter.ExportHeads, csvExporter.SplitSign) + "\n")
		if err != nil {
			return err
		}
	}
	csvExporter.csvFile = fp
	return nil
}

// ExportData 输出一行内容
func (csvExporter *CsvExporter) ExportData(dataList ...interface{}) error {
	var resultString string
	for _, item := range dataList {
		valueElem := reflect.ValueOf(item)
		itemElem := valueElem.Type()
		var lineString []string
		for i := 0; i < itemElem.NumField(); i++ {
			val := valueElem.FieldByName(itemElem.Field(i).Name).String()
			lineString = append(lineString, val)
		}
		resultString += strings.Join(lineString, csvExporter.SplitSign) + "\n"
	}
	_, err := csvExporter.csvFile.WriteString(resultString)
	if err != nil { return err }
	return nil
}

// Close 关闭 CsvExporter，做收尾工作
func (csvExporter *CsvExporter) Close() error {
	if csvExporter.csvFile != nil {
		err := csvExporter.csvFile.Close()
		if err != nil { return err }
	}
	return nil
}
