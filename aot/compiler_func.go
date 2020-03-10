package aot

type funcCompiler struct {
	printer
}

func newFuncCompiler() funcCompiler {
	return funcCompiler{newPrinter()}
}

func (c *funcCompiler) genParams(paramCount int) {
	for i := 0; i < paramCount; i++ {
		c.printf("p%d", i)
		if i < paramCount-1 {
			c.print(", ")
		} else {
			c.print(" uint64")
		}
	}
}

func (c *funcCompiler) genResults(resultCount int) {
	if resultCount == 1 {
		c.print(" uint64")
	}
}
