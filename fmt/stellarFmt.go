package stellarfmt

import (
	"bytes"
	"text/template"

	"go.polydawn.net/go-timeless-api/repeatr/fmt"
)

type Printer struct {
	PrintLog func(string)
	Repeatr  repeatrfmt.Printer
}

func Tmpl(tmpl string, obj interface{}) string {
	var buf bytes.Buffer
	if err := template.Must(
		template.New("").
			Funcs(template.FuncMap{
				"inc": func(i int) int {
					return i + 1
				},
			}).
			Parse(tmpl),
	).Execute(&buf, obj); err != nil {
		panic(err)
	}
	return buf.String()
}
