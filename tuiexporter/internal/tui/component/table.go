package component

import "github.com/rivo/tview"

type getTextByDataFunc[T any] func(data *T) string

type cellMapper[T any] struct {
	header       string
	getTextRowFn getTextByDataFunc[T]
}

type cellMappers[T any] map[int]*cellMapper[T]

func getCellFromData[T any](mappers cellMappers[T], data *T, column int) *tview.TableCell {
	text := "N/A"

	if cell, ok := mappers[column]; ok {
		text = cell.getTextRowFn(data)
	}

	if text == "" {
		text = "N/A"
	}

	text = tview.Escape(text)

	return tview.NewTableCell(text)
}
