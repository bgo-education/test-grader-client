package scan

import (
	"encoding/csv"
	"fmt"
	"os"

	"github.com/tealeg/xlsx"
)

const (
	XLSX = ".xlsx"
	CSV  = ".csv"
)

type Report struct {
	Header []string
	Data   [][]string
}

func Header(num int) []string {
	header := []string{
		"Ma so",
		"Ma de",
		"Anh",
	}

	for i := 1; i <= num; i++ {
		header = append(header, fmt.Sprintf("Cau %d", i))
	}
	return header
}

func NewReport(header []string) *Report {
	return &Report{
		Header: header,
	}
}

func (r *Report) Cols() int {
	return len(r.Header)
}

func (r *Report) Size() int {
	return len(r.Data)
}

func (r *Report) Add(data []string) {
	r.Data = append(r.Data, data)
}

func (r *Report) ToCSV(dst string) error {
	if r.Size() == 0 {
		fmt.Println("Data empty")
		return nil
	}

	f, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer f.Close()

	tsv := csv.NewWriter(f)
	defer tsv.Flush()

	tsv.Write(r.Header)
	tsv.WriteAll(r.Data)

	return nil
}

func (r *Report) ToXLSX(dst string) error {
	if r.Size() == 0 {
		fmt.Println("Data empty")
		return nil
	}

	file := xlsx.NewFile()
	sheet, err := file.AddSheet("result")
	if err != nil {
		return err
	}

	data := append([][]string{r.Header}, r.Data...)
	for _, r := range data {
		row := sheet.AddRow()
		for _, c := range r {
			cell := row.AddCell()
			cell.Value = c
		}
	}

	return file.Save(dst)
}
