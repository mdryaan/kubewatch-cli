package output

import (
	"os"

	"github.com/olekukonko/tablewriter"
)

type TablePrinter struct {
	table *tablewriter.Table
}

func NewTable(headers []string) *TablePrinter {
	t := tablewriter.NewWriter(os.Stdout)
	t.SetHeader(headers)
	t.SetBorder(false)
	t.SetColumnSeparator("  ")
	t.SetHeaderLine(false)
	t.SetHeaderAlignment(tablewriter.ALIGN_LEFT)
	t.SetAlignment(tablewriter.ALIGN_LEFT)
	t.SetTablePadding("  ")
	t.SetNoWhiteSpace(true)
	t.SetAutoWrapText(false)
	t.SetAutoFormatHeaders(false)
	t.SetHeaderColor(
		tablewriter.Colors{tablewriter.Bold, tablewriter.FgCyanColor},
		tablewriter.Colors{tablewriter.Bold, tablewriter.FgCyanColor},
		tablewriter.Colors{tablewriter.Bold, tablewriter.FgCyanColor},
		tablewriter.Colors{tablewriter.Bold, tablewriter.FgCyanColor},
		tablewriter.Colors{tablewriter.Bold, tablewriter.FgCyanColor},
		tablewriter.Colors{tablewriter.Bold, tablewriter.FgCyanColor},
		tablewriter.Colors{tablewriter.Bold, tablewriter.FgCyanColor},
		tablewriter.Colors{tablewriter.Bold, tablewriter.FgCyanColor},
	)
	return &TablePrinter{table: t}
}

func (tp *TablePrinter) AddRow(row []string) {
	tp.table.Append(row)
}

func (tp *TablePrinter) Render() {
	tp.table.Render()
}

func (tp *TablePrinter) SetColumnAlignment(alignments []int) {
	tp.table.SetColumnAlignment(alignments)
}
