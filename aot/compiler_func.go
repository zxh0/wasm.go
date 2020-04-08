package aot

type funcCompiler struct {
	printer
}

func newFuncCompiler() funcCompiler {
	return funcCompiler{newPrinter()}
}

func (c *funcCompiler) genParams(paramCount int) {
	for i := 0; i < paramCount; i++ {
		c.printf("s%d", i)
		c.printIf(i < paramCount-1, ", ", " uint64")
	}
}

func (c *funcCompiler) genResults(resultCount int) {
	c.printIf(resultCount == 1, " uint64", "")
}
