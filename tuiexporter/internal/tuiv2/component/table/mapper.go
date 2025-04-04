package table

import "github.com/charmbracelet/bubbles/table"

type GetTextByDataFunc[T any] func(data T) string

type CellMapper[T any] struct {
	Header       string
	GetTextRowFn GetTextByDataFunc[T]
}

type CellMappers[T any] []*CellMapper[T]

func getColumns[T any](mappers CellMappers[T], maxlen []int) []table.Column {
	columns := make([]table.Column, 0, len(mappers))

	for i, mapper := range mappers {
		if len(mapper.Header) > maxlen[i] {
			maxlen[i] = len(mapper.Header)
		}
		columns = append(columns, table.Column{
			Title: mapper.Header,
			Width: maxlen[i],
		})
	}

	return columns
}
