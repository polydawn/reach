package stellarfmt

import (
	"bytes"
	"io"
)

var (
	defaultColor = []byte("\033[1;33m")
	colorReset   = []byte("\033[0m")
)

type PrinterLogAnsi struct{ Stderr io.Writer }

func (p PrinterLogAnsi) PrintLog(msg string) {
	var buf bytes.Buffer
	buf.Write(defaultColor)
	buf.WriteString(msg)
	buf.Write(colorReset)
	p.Stderr.Write(buf.Bytes())
}
