package aot

type funcCompiler struct {
	printer
}

func newFuncCompiler() funcCompiler {
	return funcCompiler{newPrinter()}
}

func (c *funcCompiler) genParams(paramCount int) {
	for i := 0; i < paramCount; i++ {
		c.printf("l%d", i)
		c.printIf(i < paramCount-1, ", ", " uint64")
	}
}

func (c *funcCompiler) genResults(resultCount int) {
	if resultCount == 1 {
		c.print(" uint64")
	}
}
