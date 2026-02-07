package ui

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestTable_ContainsHeaders(t *testing.T) {
	headers := []string{"Name", "Version", "Status"}
	rows := [][]string{
		{"core", "1.0.0", "ok"},
	}

	result := Table(headers, rows)

	for _, h := range headers {
		assert.Contains(t, result, h, "Table should contain header: %s", h)
	}
}

func TestTable_ContainsRowData(t *testing.T) {
	headers := []string{"Package", "Version"}
	rows := [][]string{
		{"core", "1.2.3"},
		{"api", "2.0.0"},
	}

	result := Table(headers, rows)

	for _, row := range rows {
		for _, cell := range row {
			assert.Contains(t, result, cell, "Table should contain cell: %s", cell)
		}
	}
}

func TestTable_EmptyRows(t *testing.T) {
	headers := []string{"Package", "Version"}
	rows := [][]string{}

	result := Table(headers, rows)

	// Should still render headers even with no rows
	for _, h := range headers {
		assert.Contains(t, result, h, "Table with no rows should still contain header: %s", h)
	}
}

func TestTable_SingleRow(t *testing.T) {
	headers := []string{"Package", "Old", "New"}
	rows := [][]string{
		{"core", "1.0.0", "1.1.0"},
	}

	result := Table(headers, rows)

	assert.Contains(t, result, "core")
	assert.Contains(t, result, "1.0.0")
	assert.Contains(t, result, "1.1.0")
}
