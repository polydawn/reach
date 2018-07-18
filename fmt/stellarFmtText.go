package stellarfmt

import (
	"io"
)

type PrinterLogText struct{ Stderr io.Writer }

func (p PrinterLogText) PrintLog(msg string) {
	p.Stderr.Write([]byte(msg))
}
