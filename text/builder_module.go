package text

import (
	"github.com/zxh0/wasm.go/binary"
)

type moduleBuilder struct {
	round    int
	module   *binary.Module
	ftyBySig map[string]int // sig  -> ftyIdx
	ftyNames *symbolTable   // name -> ftyIdx
	funNames *symbolTable   // name -> funIdx
	tabNames *symbolTable   // name -> tabIdx
	memNames *symbolTable   // name -> memIdx
	glbNames *symbolTable   // name -> glbIdx
}

func newModuleBuilder() *moduleBuilder {
	return &moduleBuilder{
		module:   &binary.Module{},
		ftyBySig: map[string]int{},
		ftyNames: newSymbolTable("function type"),
		funNames: newSymbolTable("function"),
		tabNames: newSymbolTable("table"),
		memNames: newSymbolTable("memory"),
		glbNames: newSymbolTable("global"),
	}
}

func (b *moduleBuilder) getFuncTypeIdx(_var string) (int, error) {
	return b.ftyNames.getIdx(_var)
}
func (b *moduleBuilder) getFuncIdx(_var string) (int, error) {
	return b.funNames.getIdx(_var)
}
func (b *moduleBuilder) getTableIdx(_var string) (int, error) {
	return b.tabNames.getIdx(_var)
}
func (b *moduleBuilder) getMemIdx(_var string) (int, error) {
	return b.memNames.getIdx(_var)
}
func (b *moduleBuilder) getGlobalIdx(_var string) (int, error) {
	return b.glbNames.getIdx(_var)
}

func (b *moduleBuilder) ensureNoStart() error {
	if b.module.StartSec != nil {
		return newVerificationError("multiple start sections")
	}
	return nil
}
func (b *moduleBuilder) ensureNoNonImports() error {
	if b.funNames.defined > 0 ||
		b.tabNames.defined > 0 ||
		b.memNames.defined > 0 ||
		b.glbNames.defined > 0 {
		return newSemanticError("imports must occur before all non-import definitions")
	}
	return nil
}
func (b *moduleBuilder) checkCount(kind string) error {
	if kind == "table" && b.tabNames.imported > 0 {
		return newVerificationError("only one table allowed")
	}
	if kind == "memory" && b.memNames.imported > 0 {
		return newVerificationError("only one memory block allowed")
	}
	return nil
}

func (b *moduleBuilder) importName(kind, name string) error {
	if err := b.ensureNoNonImports(); err != nil {
		return err
	}
	switch kind {
	case "func":
		return b.funNames.importName(name)
	case "table":
		return b.tabNames.importName(name)
	case "memory":
		return b.memNames.importName(name)
	case "global":
		return b.glbNames.importName(name)
	default:
		panic("unreachable")
	}
}
func (b *moduleBuilder) defineName(kind, name string) error {
	switch kind {
	case "func":
		return b.funNames.defineName(name)
	case "table":
		return b.tabNames.defineName(name)
	case "memory":
		return b.memNames.defineName(name)
	case "global":
		return b.glbNames.defineName(name)
	default:
		panic("unreachable")
	}
}

func (b *moduleBuilder) addTypeDef(name string, ft binary.FuncType) error {
	if err := b.ftyNames.defineName(name); err != nil {
		return err
	}

	b.module.TypeSec = append(b.module.TypeSec, ft)
	sig := ft.GetSignature()
	if _, found := b.ftyBySig[sig]; !found {
		b.ftyBySig[sig] = len(b.module.TypeSec) - 1
	}
	return nil
}
func (b *moduleBuilder) addTypeUse(ft binary.FuncType) int {
	sig := ft.GetSignature()
	if idx, found := b.ftyBySig[sig]; found {
		return idx
	}

	b.module.TypeSec = append(b.module.TypeSec, ft)
	idx := len(b.module.TypeSec) - 1
	b.ftyBySig[sig] = idx
	_ = b.ftyNames.defineName("")
	return idx
}

func (b *moduleBuilder) addImport(imp binary.Import) int {
	b.module.ImportSec = append(b.module.ImportSec, imp)
	return b.calcImportedCount(imp.Desc.Tag) - 1
}
func (b *moduleBuilder) calcImportedCount(tag byte) int {
	n := 0
	for _, imp := range b.module.ImportSec {
		if imp.Desc.Tag == tag {
			n++
		}
	}
	return n
}

func (b *moduleBuilder) addFunc(ftIdx int,
	locals []binary.Locals, expr []binary.Instruction) int {

	b.module.FuncSec = append(b.module.FuncSec, uint32(ftIdx))
	b.module.CodeSec = append(b.module.CodeSec, binary.Code{
		Locals: locals,
		Expr:   expr,
	})

	return b.funNames.imported + len(b.module.FuncSec) - 1
}

func (b *moduleBuilder) addTable(tt binary.TableType) error {
	b.module.TableSec = append(b.module.TableSec, tt)
	if b.tabNames.imported+len(b.module.TableSec) > 1 {
		return newVerificationError("only one table allowed")
	}
	return nil
}
func (b *moduleBuilder) addTableWithElems(funcIndices []binary.FuncIdx) error {
	err := b.addTable(binary.TableType{
		ElemType: binary.FuncRef,
		Limits: binary.Limits{
			Min: uint32(len(funcIndices)),
		},
	})
	b.module.ElemSec = append(b.module.ElemSec, binary.Elem{
		Table:  0,
		Offset: []binary.Instruction{newI32Const0()},
		Init:   funcIndices,
	})
	return err
}

func (b *moduleBuilder) addMemory(mt binary.MemType) error {
	b.module.MemSec = append(b.module.MemSec, mt)
	if b.memNames.imported+len(b.module.MemSec) > 1 {
		return newVerificationError("only one memory block allowed")
	}
	return nil
}

func (b *moduleBuilder) addGlobal(gt binary.GlobalType,
	expr []binary.Instruction) int {

	b.module.GlobalSec = append(b.module.GlobalSec,
		binary.Global{Type: gt, Expr: expr})
	return b.glbNames.imported + len(b.module.GlobalSec) - 1
}

func (b *moduleBuilder) addExport(name string, kind byte, idx int) {
	b.module.ExportSec = append(b.module.ExportSec,
		binary.Export{
			Name: name,
			Desc: binary.ExportDesc{
				Tag: kind,
				Idx: uint32(idx),
			},
		})
}

func (b *moduleBuilder) addStart(_var string) error {
	fIdx, err := b.getFuncIdx(_var)
	idx := uint32(fIdx)
	b.module.StartSec = &idx
	return err
}

func (b *moduleBuilder) addElem(_var string,
	offset []binary.Instruction, initData []binary.FuncIdx) error {

	if _var != "" {
		if _, err := b.getTableIdx(_var); err != nil {
			return err
		}
	}
	b.module.ElemSec = append(b.module.ElemSec, binary.Elem{
		Table:  0,
		Offset: offset,
		Init:   initData,
	})
	return nil
}

func (b *moduleBuilder) addData(_var string,
	offset []binary.Instruction, initData string) error {

	if _var != "" {
		if _, err := b.getMemIdx(_var); err != nil {
			return err
		}
	}
	b.module.DataSec = append(b.module.DataSec, binary.Data{
		Mem:    0,
		Offset: offset,
		Init:   escape(initData),
	})
	return nil
}
