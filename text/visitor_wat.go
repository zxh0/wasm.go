package text

import (
	"math"

	"github.com/antlr/antlr4/runtime/Go/antlr"
	"github.com/zxh0/wasm.go/binary"
	"github.com/zxh0/wasm.go/text/parser"
)

var _ parser.WASTVisitor = (*watVisitor)(nil)

type watVisitor struct {
	baseVisitor
	errorReporter
	moduleBuilder *moduleBuilder
	codeBuilder   *codeBuilder
}

func newWatVisitor() parser.WASTVisitor {
	return &watVisitor{
		errorReporter: errorReporter{
			reportsValidationError: true,
		},
	}
}

func (v *watVisitor) VisitModule(ctx *parser.ModuleContext) interface{} {
	return ctx.WatModule().Accept(v).(*WatModule).Module
}

func (v *watVisitor) VisitWatModule(ctx *parser.WatModuleContext) interface{} {
	name := getText(ctx.NAME())
	v.moduleBuilder = newModuleBuilder()
	v.moduleBuilder.pass = 1
	nv := &watNamesVisitor{
		errorReporter: v.errorReporter,
		moduleBuilder: v.moduleBuilder,
	}
	for _, field := range ctx.AllModuleField() {
		field.Accept(v)
		field.Accept(nv)
	}
	v.moduleBuilder.pass = 2
	for _, field := range ctx.AllModuleField() {
		field.Accept(v)
	}
	return &WatModule{
		Line:   ctx.GetKw().GetLine(),
		Name:   name,
		Module: v.moduleBuilder.module,
	}
}
func (v *watVisitor) VisitModuleField(ctx *parser.ModuleFieldContext) interface{} {
	switch v.moduleBuilder.pass {
	case 1: // typeDef
		if ctx.TypeDef() != nil {
			ctx.TypeDef().Accept(v)
		}
	case 2: // other fields
		if ctx.TypeDef() == nil {
			ctx.GetChild(0).(antlr.ParseTree).Accept(v)
		}
	}
	return nil
}

func (v *watVisitor) VisitTypeDef(ctx *parser.TypeDefContext) interface{} {
	name := ctx.NAME()
	ft := ctx.FuncType().Accept(v).(binary.FuncType)
	err := v.moduleBuilder.addTypeDef(getText(name), ft)
	v.reportErr(err, name)
	return nil
}

func (v *watVisitor) VisitImport_(ctx *parser.Import_Context) interface{} {
	imp := binary.Import{
		Module: getStr(ctx.STRING(0)),
		Name:   getStr(ctx.STRING(1)),
		Desc:   ctx.ImportDesc().Accept(v).(binary.ImportDesc),
	}
	v.moduleBuilder.addImport(imp)
	return nil
}
func (v *watVisitor) VisitImportDesc(ctx *parser.ImportDescContext) interface{} {
	switch ctx.GetKind().GetText() {
	case "func":
		return binary.ImportDesc{
			Tag:      binary.ImportTagFunc,
			FuncType: uint32(ctx.TypeUse().Accept(v).(int)),
		}
	case "table":
		return binary.ImportDesc{
			Tag:   binary.ImportTagTable,
			Table: ctx.TableType().Accept(v).(binary.TableType),
		}
	case "memory":
		return binary.ImportDesc{
			Tag: binary.ImportTagMem,
			Mem: ctx.MemoryType().Accept(v).(binary.MemType),
		}
	case "global":
		return binary.ImportDesc{
			Tag:    binary.ImportTagGlobal,
			Global: ctx.GlobalType().Accept(v).(binary.GlobalType),
		}
	default:
		panic("unreachable")
	}
}

func (v *watVisitor) VisitFunc_(ctx *parser.Func_Context) interface{} {
	v.codeBuilder = newCodeBuilder()
	ftIdx := ctx.TypeUse().Accept(v).(int)

	idx := 0
	if ctx.EmbeddedIm() != nil {
		imp := ctx.EmbeddedIm().Accept(v).(binary.Import)
		imp.Desc = binary.ImportDesc{
			Tag:      binary.ImportTagFunc,
			FuncType: uint32(ftIdx),
		}
		idx = v.moduleBuilder.addImport(imp)
	} else {
		for _, local := range ctx.AllFuncLocal() {
			local.Accept(v)
		}
		expr := getExpr(ctx.Expr(), v)
		locals := v.codeBuilder.locals
		idx = v.moduleBuilder.addFunc(ftIdx, locals, expr)
	}

	if ctx.EmbeddedEx() != nil {
		names := ctx.EmbeddedEx().Accept(v).([]string)
		for _, name := range names {
			v.moduleBuilder.addExport(name, binary.ExportTagFunc, idx)
		}
	}

	v.codeBuilder = nil
	return nil
}
func (v *watVisitor) VisitFuncLocal(ctx *parser.FuncLocalContext) interface{} {
	if name := ctx.NAME(); name != nil {
		vt := ctx.ValType(0).Accept(v).(binary.ValType)
		err := v.codeBuilder.addLocal(name.GetText(), vt)
		v.reportErr(err, name)
	} else {
		for _, vt := range ctx.AllValType() {
			_ = v.codeBuilder.addLocal("", vt.Accept(v).(binary.ValType))
		}
	}
	return nil
}

func (v *watVisitor) VisitTable(ctx *parser.TableContext) interface{} {
	if ctx.EmbeddedIm() != nil {
		imp := ctx.EmbeddedIm().Accept(v).(binary.Import)
		imp.Desc = binary.ImportDesc{
			Tag:   binary.ImportTagTable,
			Table: ctx.TableType().Accept(v).(binary.TableType),
		}
		v.moduleBuilder.addImport(imp)
	} else if ctx.TableType() != nil {
		tt := ctx.TableType().Accept(v).(binary.TableType)
		err := v.moduleBuilder.addTable(tt)
		v.reportErr(err, ctx.GetChild(1))
	} else if ctx.ElemType() != nil {
		funcIndices := ctx.FuncVars().Accept(v).([]binary.FuncIdx)
		err := v.moduleBuilder.addTableWithElems(funcIndices)
		v.reportErr(err, ctx.GetChild(1))
	}
	if ctx.EmbeddedEx() != nil {
		names := ctx.EmbeddedEx().Accept(v).([]string)
		for _, name := range names {
			v.moduleBuilder.addExport(name, binary.ExportTagTable, 0)
		}
	}

	return nil
}

func (v *watVisitor) VisitMemory(ctx *parser.MemoryContext) interface{} {
	if ctx.EmbeddedIm() != nil {
		imp := ctx.EmbeddedIm().Accept(v).(binary.Import)
		imp.Desc = binary.ImportDesc{
			Tag: binary.ImportTagMem,
			Mem: ctx.MemoryType().Accept(v).(binary.MemType),
		}
		v.moduleBuilder.addImport(imp)
	} else if ctx.MemoryType() != nil {
		mt := ctx.MemoryType().Accept(v).(binary.MemType)
		err := v.moduleBuilder.addMemory(mt)
		v.reportErr(err, ctx.GetChild(1))
	} else {
		offset := []binary.Instruction{newI32Const0()}
		initData := getAllStr(ctx.AllSTRING())
		min := uint32(math.Ceil(float64(len(initData)) / binary.PageSize))
		mt := binary.Limits{Min: min} // TODO
		err := v.moduleBuilder.addMemory(mt)
		v.reportErr(err, ctx.GetChild(1))
		_ = v.moduleBuilder.addData("", offset, initData)
	}

	if ctx.EmbeddedEx() != nil {
		names := ctx.EmbeddedEx().Accept(v).([]string)
		for _, name := range names {
			v.moduleBuilder.addExport(name, binary.ExportTagMem, 0) // TODO
		}
	}

	return nil
}

func (v *watVisitor) VisitGlobal(ctx *parser.GlobalContext) interface{} {
	idx := 0
	if ctx.EmbeddedIm() != nil {
		imp := ctx.EmbeddedIm().Accept(v).(binary.Import)
		imp.Desc = binary.ImportDesc{
			Tag:    binary.ImportTagGlobal,
			Global: ctx.GlobalType().Accept(v).(binary.GlobalType),
		}
		idx = v.moduleBuilder.addImport(imp)
	} else {
		gt := ctx.GlobalType().Accept(v).(binary.GlobalType)
		expr := getExpr(ctx.Expr(), v)
		idx = v.moduleBuilder.addGlobal(gt, expr)
	}
	if ctx.EmbeddedEx() != nil {
		names := ctx.EmbeddedEx().Accept(v).([]string)
		for _, name := range names {
			v.moduleBuilder.addExport(name, binary.ExportTagGlobal, idx)
		}
	}

	return nil
}

func (v *watVisitor) VisitExport(ctx *parser.ExportContext) interface{} {
	var idx int
	var err error

	name := getStr(ctx.STRING())
	kindAndVar := ctx.ExportDesc().Accept(v).([]string)
	switch kindAndVar[0] {
	case "func":
		idx, err = v.moduleBuilder.getFuncIdx(kindAndVar[1])
		v.moduleBuilder.addExport(name, binary.ImportTagFunc, idx)
	case "table":
		idx, err = v.moduleBuilder.getTableIdx(kindAndVar[1])
		v.moduleBuilder.addExport(name, binary.ImportTagTable, idx)
	case "memory":
		idx, err = v.moduleBuilder.getMemIdx(kindAndVar[1])
		v.moduleBuilder.addExport(name, binary.ImportTagMem, idx)
	case "global":
		idx, err = v.moduleBuilder.getGlobalIdx(kindAndVar[1])
		v.moduleBuilder.addExport(name, binary.ImportTagGlobal, idx)
	default:
		panic("unreachable")
	}

	if err != nil {
		v.reportErr(err, ctx.ExportDesc().GetChild(2).GetChild(0))
	}
	return nil
}
func (v *watVisitor) VisitExportDesc(ctx *parser.ExportDescContext) interface{} {
	return []string{
		ctx.GetKind().GetText(),
		ctx.Variable().GetText(),
	}
}

func (v *watVisitor) VisitStart(ctx *parser.StartContext) interface{} {
	err := v.moduleBuilder.ensureNoStart()
	v.reportErr(err, ctx.GetChild(1))
	_var := ctx.Variable()
	err = v.moduleBuilder.addStart(_var.GetText())
	v.reportErr(err, _var)
	return nil
}

func (v *watVisitor) VisitElem(ctx *parser.ElemContext) interface{} {
	_var := ctx.Variable()
	offset := ctx.Expr().Accept(v).([]binary.Instruction)
	initData := ctx.FuncVars().Accept(v).([]binary.FuncIdx)
	err := v.moduleBuilder.addElem(getText(_var), offset, initData)
	v.reportErr(err, _var)
	return nil
}

func (v *watVisitor) VisitData(ctx *parser.DataContext) interface{} {
	_var := ctx.Variable()
	offset := ctx.Expr().Accept(v).([]binary.Instruction)
	initData := getAllStr(ctx.AllSTRING())
	err := v.moduleBuilder.addData(getText(_var), offset, initData)
	v.reportErr(err, _var)
	return nil
}

func (v *watVisitor) VisitEmbeddedIm(ctx *parser.EmbeddedImContext) interface{} {
	return binary.Import{
		Module: getStr(ctx.STRING(0)),
		Name:   getStr(ctx.STRING(1)),
	}
}
func (v *watVisitor) VisitEmbeddedEx(ctx *parser.EmbeddedExContext) interface{} {
	names := make([]string, 0)
	for _, name := range ctx.AllSTRING() {
		names = append(names, getStr(name))
	}
	return names
}
func (v *watVisitor) VisitTypeUse(ctx *parser.TypeUseContext) interface{} {
	ft := ctx.FuncType().Accept(v).(binary.FuncType)
	if _var := ctx.Variable(); _var != nil {
		idx, err := v.moduleBuilder.getFuncTypeIdx(_var.GetText())
		if err != nil {
			v.reportErr(err, _var)
			return idx
		}

		ftUse := v.moduleBuilder.module.TypeSec[idx]
		if len(ft.ParamTypes) == 0 && len(ft.ResultTypes) == 0 {
			if v.codeBuilder != nil { // TODO
				for range ftUse.ParamTypes {
					_ = v.codeBuilder.addParam("")
				}
			}
		} else {
			if ft.GetSignature() != ftUse.GetSignature() {
				msg := "type mismatch"
				if _, ok := ctx.GetParent().(parser.IPlainInstrContext); ok {
					msg += " in call_indirect"
				}
				err := newVerificationError(msg)
				v.reportErr(err, ctx.GetChild(1))
			}
		}
		return idx
	}
	return v.moduleBuilder.addTypeUse(ft)
}
func (v *watVisitor) VisitFuncVars(ctx *parser.FuncVarsContext) interface{} {
	funcIndices := make([]binary.FuncIdx, 0)
	for _, _var := range ctx.AllVariable() {
		idx, err := v.moduleBuilder.getFuncIdx(_var.GetText())
		v.reportErr(err, _var)
		funcIndices = append(funcIndices, uint32(idx))
	}
	return funcIndices
}

func (v *watVisitor) VisitValType(ctx *parser.ValTypeContext) interface{} {
	switch ctx.GetText() {
	case "i32":
		return binary.ValTypeI32
	case "i64":
		return binary.ValTypeI64
	case "f32":
		return binary.ValTypeF32
	case "f64":
		return binary.ValTypeF64
	default:
		panic("unreachable")
	}
}
func (v *watVisitor) VisitBlockType(ctx *parser.BlockTypeContext) interface{} {
	if ctx.Result() != nil {
		return ctx.Result().Accept(v).(binary.BlockType)
	}
	return binary.BlockType{}
}
func (v *watVisitor) VisitGlobalType(ctx *parser.GlobalTypeContext) interface{} {
	vt := ctx.ValType().Accept(v).(binary.ValType)
	mut := binary.MutConst
	if ctx.GetChildCount() > 1 {
		mut = binary.MutVar
	}
	return binary.GlobalType{
		ValType: vt,
		Mut:     mut,
	}
}
func (v *watVisitor) VisitMemoryType(ctx *parser.MemoryTypeContext) interface{} {
	return ctx.Limits().Accept(v)
}
func (v *watVisitor) VisitTableType(ctx *parser.TableTypeContext) interface{} {
	return binary.TableType{
		ElemType: binary.FuncRef,
		Limits:   ctx.Limits().Accept(v).(binary.Limits),
	}
}
func (v *watVisitor) VisitLimits(ctx *parser.LimitsContext) interface{} {
	mt := binary.Limits{}
	mt.Min = parseU32(ctx.Nat(0).GetText())
	if max := ctx.Nat(1); max != nil {
		mt.Tag = 1
		mt.Max = parseU32(max.GetText())
	}
	return mt
}
func (v *watVisitor) VisitFuncType(ctx *parser.FuncTypeContext) interface{} {
	ft := binary.FuncType{}
	for _, param := range ctx.AllParam() {
		ft.ParamTypes = append(ft.ParamTypes,
			param.Accept(v).([]binary.ValType)...)
	}
	for _, result := range ctx.AllResult() {
		ft.ResultTypes = append(ft.ResultTypes,
			result.Accept(v).([]binary.ValType)...)
	}
	return ft
}
func (v *watVisitor) VisitParam(ctx *parser.ParamContext) interface{} {
	params := make([]binary.ValType, 0, 1)

	if name := ctx.NAME(); name != nil {
		vt := ctx.ValType(0).Accept(v).(binary.ValType)
		params = append(params, vt)
		if v.codeBuilder != nil {
			err := v.codeBuilder.addParam(name.GetText())
			v.reportErr(err, name)
		}
	} else {
		for _, vt := range ctx.AllValType() {
			params = append(params, vt.Accept(v).(binary.ValType))
			if v.codeBuilder != nil {
				_ = v.codeBuilder.addParam("")
			}
		}
	}

	return params
}
func (v *watVisitor) VisitResult(ctx *parser.ResultContext) interface{} {
	vts := ctx.AllValType()
	results := make([]binary.ValType, len(vts))
	for i, vt := range ctx.AllValType() {
		results[i] = vt.Accept(v).(binary.ValType)
	}
	return results
}

func (v *watVisitor) VisitExpr(ctx *parser.ExprContext) interface{} {
	expr := make([]binary.Instruction, 0)
	for _, instr := range ctx.AllInstr() {
		instrs := instr.Accept(v).([]binary.Instruction)
		expr = append(expr, instrs...)
	}
	return expr
}

func (v *watVisitor) VisitInstr(ctx *parser.InstrContext) interface{} {
	if ctx.PlainInstr() != nil {
		instr := ctx.PlainInstr().Accept(v).(binary.Instruction)
		return []binary.Instruction{instr}
	}
	if ctx.BlockInstr() != nil {
		instr := ctx.BlockInstr().Accept(v).(binary.Instruction)
		return []binary.Instruction{instr}
	}
	return ctx.FoldedInstr().Accept(v)
}

func (v *watVisitor) VisitFoldedInstr(ctx *parser.FoldedInstrContext) interface{} {
	instrs := make([]binary.Instruction, 0)
	for _, foldedInstr := range ctx.AllFoldedInstr() {
		instrs = append(instrs, foldedInstr.Accept(v).([]binary.Instruction)...)
	}

	var instr binary.Instruction
	if op := ctx.GetOp(); op != nil {
		v.codeBuilder.enterBlock()
		defer v.codeBuilder.exitBlock()

		if label := ctx.GetLabel(); label != nil {
			v.codeBuilder.defineLabel(label.GetText())
		}

		op := ctx.GetOp().GetText()
		rt := ctx.BlockType().Accept(v).(binary.BlockType)
		expr1 := ctx.Expr(0).Accept(v).([]binary.Instruction)
		expr2 := getExpr(ctx.Expr(1), v)
		instr = newBlockInstr(op, rt, expr1, expr2)
	} else {
		instr = ctx.PlainInstr().Accept(v).(binary.Instruction)
	}

	instrs = append(instrs, instr)
	return instrs
}

func (v *watVisitor) VisitBlockInstr(ctx *parser.BlockInstrContext) interface{} {
	v.codeBuilder.enterBlock()
	defer v.codeBuilder.exitBlock()

	if label := ctx.GetLabel(); label != nil {
		v.codeBuilder.defineLabel(label.GetText())
	}

	op := ctx.GetOp().GetText()
	rt := ctx.BlockType().Accept(v).(binary.BlockType)
	expr1 := ctx.Expr(0).Accept(v).([]binary.Instruction)
	expr2 := getExpr(ctx.Expr(1), v)
	return newBlockInstr(op, rt, expr1, expr2)
}

func (v *watVisitor) VisitPlainInstr(ctx *parser.PlainInstrContext) interface{} {
	if ctx.ConstInstr() != nil {
		return ctx.ConstInstr().Accept(v).(binary.Instruction)
	}

	op := ctx.GetOp().GetText()
	instr := newInstruction(op)
	opcode := instr.Opcode

	switch opcode {
	case binary.Br, binary.BrIf:
		_var := ctx.Variable(0)
		idx, err := v.codeBuilder.getBrLabelIdx(_var.GetText())
		instr.Args = uint32(idx)
		v.reportErr(err, ctx.Variable(0))
	case binary.BrTable:
		labels := make([]uint32, 0, 1)
		for _, _var := range ctx.AllVariable() {
			idx, err := v.codeBuilder.getBrLabelIdx(_var.GetText())
			labels = append(labels, uint32(idx))
			v.reportErr(err, _var)
		}
		instr.Args = binary.BrTableArgs{
			Labels:  labels[:len(labels)-1],
			Default: labels[len(labels)-1],
		}
	case binary.Call:
		_var := ctx.Variable(0)
		idx, err := v.moduleBuilder.getFuncIdx(_var.GetText())
		instr.Args = uint32(idx)
		v.reportErr(err, _var)
	case binary.CallIndirect:
		ftIdx := ctx.TypeUse().Accept(v).(int)
		instr.Args = uint32(ftIdx)
		// TODO
	}

	if opcode >= binary.LocalGet && opcode <= binary.LocalTee {
		_var := ctx.Variable(0)
		if v.codeBuilder != nil {
			idx, err := v.codeBuilder.getLocalIdx(_var.GetText())
			instr.Args = uint32(idx)
			v.reportErr(err, _var)
		} else {
			instr.Args = parseU32(_var.GetText())
		}
	} else if opcode >= binary.GlobalGet && opcode <= binary.GlobalSet {
		_var := ctx.Variable(0)
		idx, err := v.moduleBuilder.getGlobalIdx(_var.GetText())
		instr.Args = uint32(idx)
		v.reportErr(err, _var)
	} else if opcode >= binary.I32Load && opcode <= binary.I64Store32 {
		instr.Args = ctx.MemArg().Accept(v).(binary.MemArg)
	}

	return instr
}

func (v *watVisitor) VisitConstInstr(ctx *parser.ConstInstrContext) interface{} {
	instr := newInstruction(ctx.GetOp().GetText())
	val := ctx.Value().GetText()
	switch instr.Opcode {
	case binary.I32Const:
		instr.Args = parseI32(val)
	case binary.I64Const:
		instr.Args = parseI64(val)
	case binary.F32Const:
		instr.Args = parseF32(val)
	case binary.F64Const:
		instr.Args = parseF64(val)
	default:
		panic("unreachable")
	}
	return instr
}

func (v *watVisitor) VisitMemArg(ctx *parser.MemArgContext) interface{} {
	memArg := binary.MemArg{}
	if offset := ctx.GetOffset(); offset != nil {
		memArg.Offset = parseU32(offset.GetText())
	}
	if align := ctx.GetAlign(); align != nil {
		alignVal := parseU32(align.GetText())
		switch alignVal {
		case 1:
			memArg.Align = 0
		case 2:
			memArg.Align = 1
		case 4:
			memArg.Align = 2
		case 8:
			memArg.Align = 3
		case 16:
			memArg.Align = 4
		case 32:
			memArg.Align = 5
		case 64:
			memArg.Align = 6
		default:
			panic("invalid align") // TODO
		}
	}
	return memArg
}
