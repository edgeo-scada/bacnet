package main

import (
	"fmt"
	"io"
	"os"
)

// OutputFormat represents output format types
type OutputFormat string

const (
	FormatTable OutputFormat = "table"
	FormatJSON  OutputFormat = "json"
	FormatCSV   OutputFormat = "csv"
	FormatRaw   OutputFormat = "raw"
)

// Formatter handles output formatting
type Formatter struct {
	format OutputFormat
	writer io.Writer
}

// NewFormatter creates a new formatter
func NewFormatter(format string) *Formatter {
	return &Formatter{
		format: OutputFormat(format),
		writer: os.Stdout,
	}
}

// SetWriter sets the output writer
func (f *Formatter) SetWriter(w io.Writer) {
	f.writer = w
}

// Printf formats and prints output
func (f *Formatter) Printf(format string, args ...interface{}) {
	fmt.Fprintf(f.writer, format, args...)
}

// Println prints a line
func (f *Formatter) Println(args ...interface{}) {
	fmt.Fprintln(f.writer, args...)
}

// PrintTable prints data in table format
func (f *Formatter) PrintTable(headers []string, rows [][]string) {
	// Calculate column widths
	widths := make([]int, len(headers))
	for i, h := range headers {
		widths[i] = len(h)
	}
	for _, row := range rows {
		for i, cell := range row {
			if i < len(widths) && len(cell) > widths[i] {
				widths[i] = len(cell)
			}
		}
	}

	// Print headers
	for i, h := range headers {
		fmt.Fprintf(f.writer, "%-*s ", widths[i], h)
	}
	fmt.Fprintln(f.writer)

	// Print separator
	for i := range headers {
		for j := 0; j < widths[i]; j++ {
			fmt.Fprint(f.writer, "-")
		}
		fmt.Fprint(f.writer, " ")
	}
	fmt.Fprintln(f.writer)

	// Print rows
	for _, row := range rows {
		for i, cell := range row {
			if i < len(widths) {
				fmt.Fprintf(f.writer, "%-*s ", widths[i], cell)
			}
		}
		fmt.Fprintln(f.writer)
	}
}

// PrintKeyValue prints key-value pairs
func (f *Formatter) PrintKeyValue(pairs map[string]interface{}, order []string) {
	maxKeyLen := 0
	for _, key := range order {
		if len(key) > maxKeyLen {
			maxKeyLen = len(key)
		}
	}

	for _, key := range order {
		if val, ok := pairs[key]; ok {
			fmt.Fprintf(f.writer, "%-*s: %v\n", maxKeyLen, key, val)
		}
	}
}
