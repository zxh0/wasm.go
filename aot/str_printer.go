package aot

import (
	"fmt"
	"strings"
)

type printer struct {
	sb *strings.Builder
}

func (p *printer) print(s string) {
	p.sb.WriteString(s)
}

func (p *printer) println(s string) {
	p.sb.WriteString(s)
	p.sb.WriteByte('\n')
}

func (p *printer) printf(format string, a ...interface{}) {
	p.sb.WriteString(fmt.Sprintf(format, a...))
}
