package text

type symbolTable struct {
	kind      string
	idxByName map[string]int
	imported  int
	defined   int
}

func newSymbolTable(kind string) *symbolTable {
	return &symbolTable{
		kind:      kind,
		idxByName: map[string]int{},
	}
}

func (st *symbolTable) getIdx(_var string) (int, error) {
	idx := -1
	if _var[0] == '$' {
		var found bool
		if idx, found = st.idxByName[_var]; !found {
			return -1, newSemanticError(`undefined %s variable "%s"`,
				st.kind, _var)
		}
	} else {
		idx = int(parseU32(_var))
	}

	if st.kind != "label" {
		if max := st.imported + st.defined; idx >= max {
			return idx, newVerificationError("%s variable out of range: %d (max %d)",
				st.kind, idx, max-1)
		}
	}
	return idx, nil
}

func (st *symbolTable) importName(name string) error {
	idx := st.imported
	if err := st.addName(name, idx); err != nil {
		return err
	}
	st.imported++
	return nil
}
func (st *symbolTable) defineName(name string) error {
	idx := st.imported + st.defined
	if err := st.addName(name, idx); err != nil {
		return err
	}
	st.defined++
	return nil
}
func (st *symbolTable) addName(name string, idx int) error {
	if name != "" {
		if _, found := st.idxByName[name]; !found {
			st.idxByName[name] = idx
		} else {
			return newSemanticError(`redefinition of %s "%s"`,
				st.kind, name)
		}
	}
	return nil
}

func (st *symbolTable) defineLabel(name string, idx int) {
	st.idxByName[name] = idx
}
