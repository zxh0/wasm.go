package binary

import (
	"encoding/binary"
	"errors"
	"fmt"
	"io/ioutil"
	"math"
	"unicode/utf8"
)

type WasmReader struct {
	data []byte
}

func DecodeFile(filename string) (Module, error) {
	data, err := ioutil.ReadFile(filename)
	if err != nil {
		return Module{}, err
	}
	return Decode(data)
}

// TODO: return *Module ?
func Decode(data []byte) (module Module, err error) {
	defer func() {
		if r := recover(); r != nil {
			switch x := r.(type) {
			case error:
				err = x
			default:
				err = errors.New("unknown error")
			}
		}
	}()
	reader := &WasmReader{data: data}
	reader.readModule(&module)
	return
}

func (reader *WasmReader) remaining() int {
	return len(reader.data)
}

// fixed length value
func (reader *WasmReader) readByte() byte {
	if len(reader.data) < 1 {
		panic(errUnexpectedEnd)
	}
	b := reader.data[0]
	reader.data = reader.data[1:]
	return b
}
func (reader *WasmReader) readU32() uint32 {
	if len(reader.data) < 4 {
		panic(errUnexpectedEnd)
	}
	n := binary.LittleEndian.Uint32(reader.data)
	reader.data = reader.data[4:]
	return n
}
func (reader *WasmReader) readF32() float32 {
	if len(reader.data) < 4 {
		panic(errUnexpectedEnd)
	}
	n := binary.LittleEndian.Uint32(reader.data)
	reader.data = reader.data[4:]
	return math.Float32frombits(n)
}
func (reader *WasmReader) readF64() float64 {
	if len(reader.data) < 8 {
		panic(errUnexpectedEnd)
	}
	n := binary.LittleEndian.Uint64(reader.data)
	reader.data = reader.data[8:]
	return math.Float64frombits(n)
}

// variable length value
func (reader *WasmReader) readVarU32() uint32 {
	n, w := decodeVarUint(reader.data, 32)
	reader.data = reader.data[w:]
	return uint32(n)
}
func (reader *WasmReader) readVarS32() int32 {
	n, w := decodeVarInt(reader.data, 32)
	reader.data = reader.data[w:]
	return int32(n)
}
func (reader *WasmReader) readVarS64() int64 {
	n, w := decodeVarInt(reader.data, 64)
	reader.data = reader.data[w:]
	return n
}

// bytes & name
func (reader *WasmReader) readBytes() []byte {
	n := reader.readVarU32()
	if len(reader.data) < int(n) {
		panic(errUnexpectedEnd)
	}
	bytes := reader.data[:n]
	reader.data = reader.data[n:]
	return bytes
}
func (reader *WasmReader) readName() string {
	data := reader.readBytes()
	if !utf8.Valid(data) {
		panic(errors.New("invalid UTF-8 encoding"))
	}
	return string(data)
}

// module
func (reader *WasmReader) readModule(module *Module) {
	if reader.remaining() < 4 {
		panic(errors.New("unexpected end of magic header"))
	}
	module.Magic = reader.readU32()
	if module.Magic != MagicNumber {
		panic(errors.New("magic header not detected"))
	}

	if reader.remaining() < 4 {
		panic(errors.New("unexpected end of binary version"))
	}
	module.Version = reader.readU32()
	if module.Version != Version {
		panic(fmt.Errorf("unknown binary version: %d", module.Version))
	}

	reader.readSections(module)
	if len(module.FuncSec) != len(module.CodeSec) {
		panic(errors.New("function and code section have inconsistent lengths"))
	}
	if reader.remaining() > 0 {
		panic(errors.New("junk after last section"))
	}
}

// sections
func (reader *WasmReader) readSections(module *Module) {
	prevSecID := byte(0)
	for reader.remaining() > 0 {
		secID := reader.readByte()
		if secID == SecCustomID {
			module.CustomSecs = append(module.CustomSecs,
				reader.readCustomSec(module))
			continue
		}

		if secID > SecDataID {
			panic(fmt.Errorf("invalid section id: %d", secID))
		}
		if secID <= prevSecID {
			panic(fmt.Errorf("junk after last section, id: %d", secID))
		}
		prevSecID = secID

		n := reader.readVarU32()
		remainingBeforeRead := reader.remaining()
		reader.readNonCustomSec(secID, module)
		if reader.remaining()+int(n) != remainingBeforeRead {
			panic(fmt.Errorf("section size mismatch, id: %d", secID))
		}
	}
}
func (reader *WasmReader) readCustomSec(module *Module) CustomSec {
	secReader := &WasmReader{data: reader.readBytes()}
	name := secReader.readName()
	// TODO
	return CustomSec{Name: name}
}
func (reader *WasmReader) readNonCustomSec(secID byte, module *Module) {
	switch secID {
	case SecTypeID:
		module.TypeSec = reader.readTypeSec()
	case SecImportID:
		module.ImportSec = reader.readImportSec()
	case SecFuncID:
		module.FuncSec = reader.readIndices()
	case SecTableID:
		module.TableSec = reader.readTableSec()
	case SecMemID:
		module.MemSec = reader.readMemSec()
	case SecGlobalID:
		module.GlobalSec = reader.readGlobalSec()
	case SecExportID:
		module.ExportSec = reader.readExportSec()
	case SecStartID:
		module.StartSec = reader.readStartSec()
	case SecElemID:
		module.ElemSec = reader.readElemSec()
	case SecCodeID:
		module.CodeSec = reader.readCodeSec()
	case SecDataID:
		module.DataSec = reader.readDataSec()
	}
	return
}

// type sec
func (reader *WasmReader) readTypeSec() []FuncType {
	vec := make([]FuncType, reader.readVarU32())
	for i := range vec {
		vec[i] = reader.readFuncType()
	}
	return vec
}

// import sec
func (reader *WasmReader) readImportSec() []Import {
	vec := make([]Import, reader.readVarU32())
	for i := range vec {
		vec[i] = reader.readImport()
	}
	return vec
}
func (reader *WasmReader) readImport() Import {
	return Import{
		Module: reader.readName(),
		Name:   reader.readName(),
		Desc:   reader.readImportDesc(),
	}
}
func (reader *WasmReader) readImportDesc() ImportDesc {
	desc := ImportDesc{Tag: reader.readByte()}
	switch desc.Tag {
	case ImportTagFunc:
		desc.FuncType = reader.readVarU32()
	case ImportTagTable:
		desc.Table = reader.readTableType()
	case ImportTagMem:
		desc.Mem = reader.readLimits()
	case ImportTagGlobal:
		desc.Global = reader.readGlobalType()
	default:
		panic(fmt.Errorf("invalid import desc tag: %d", desc.Tag))
	}
	return desc
}

// table sec
func (reader *WasmReader) readTableSec() []TableType {
	vec := make([]TableType, reader.readVarU32())
	for i := range vec {
		vec[i] = reader.readTableType()
	}
	return vec
}

// mem sec
func (reader *WasmReader) readMemSec() []MemType {
	vec := make([]MemType, reader.readVarU32())
	for i := range vec {
		vec[i] = reader.readLimits()
	}
	return vec
}

// global sec
func (reader *WasmReader) readGlobalSec() []Global {
	vec := make([]Global, reader.readVarU32())
	for i := range vec {
		vec[i] = Global{
			Type: reader.readGlobalType(),
			Expr: reader.readExpr(),
		}
	}
	return vec
}

// export sec
func (reader *WasmReader) readExportSec() []Export {
	vec := make([]Export, reader.readVarU32())
	for i := range vec {
		vec[i] = reader.readExport()
	}
	return vec
}
func (reader *WasmReader) readExport() Export {
	return Export{
		Name: reader.readName(),
		Desc: reader.readExportDesc(),
	}
}
func (reader *WasmReader) readExportDesc() ExportDesc {
	desc := ExportDesc{
		Tag: reader.readByte(),
		Idx: reader.readVarU32(),
	}
	switch desc.Tag {
	case ExportTagFunc: // func_idx
	case ExportTagTable: // table_idx
	case ExportTagMem: // mem_idx
	case ExportTagGlobal: // global_idx
	default:
		panic(fmt.Errorf("invalid export desc tag: %d", desc.Tag))
	}
	return desc
}

// start sec
func (reader *WasmReader) readStartSec() *uint32 {
	idx := reader.readVarU32()
	return &idx
}

// elem sec
func (reader *WasmReader) readElemSec() []Elem {
	vec := make([]Elem, reader.readVarU32())
	for i := range vec {
		vec[i] = reader.readElem()
	}
	return vec
}
func (reader *WasmReader) readElem() Elem {
	return Elem{
		Table:  reader.readVarU32(),
		Offset: reader.readExpr(),
		Init:   reader.readIndices(),
	}
}

// code sec
func (reader *WasmReader) readCodeSec() []Code {
	vec := make([]Code, reader.readVarU32())
	for i := range vec {
		vec[i] = reader.readCode(i)
	}
	return vec
}
func (reader *WasmReader) readCode(idx int) Code {
	n := reader.readVarU32()
	remainingBeforeRead := reader.remaining()
	code := Code{
		Locals: reader.readLocalsVec(),
		Expr:   reader.readExpr(),
	}
	if reader.remaining()+int(n) != remainingBeforeRead {
		panic(fmt.Errorf("invalid code[%d]", idx))
	}
	if code.GetLocalCount() >= math.MaxUint32 {
		panic(fmt.Errorf("too many locals: %d",
			code.GetLocalCount()))
	}
	return code
}
func (reader *WasmReader) readLocalsVec() []Locals {
	vec := make([]Locals, reader.readVarU32())
	for i := range vec {
		vec[i] = reader.readLocals()
	}
	return vec
}
func (reader *WasmReader) readLocals() Locals {
	return Locals{
		N:    reader.readVarU32(),
		Type: reader.readValType(),
	}
}

// data sec
func (reader *WasmReader) readDataSec() []Data {
	vec := make([]Data, reader.readVarU32())
	for i := range vec {
		vec[i] = reader.readData()
	}
	return vec
}
func (reader *WasmReader) readData() (data Data) {
	return Data{
		Mem:    reader.readVarU32(),
		Offset: reader.readExpr(),
		Init:   reader.readBytes(),
	}
}

// value types
func (reader *WasmReader) readValTypes() []ValType {
	vec := make([]ValType, reader.readVarU32())
	for i := range vec {
		vec[i] = reader.readValType()
	}
	return vec
}
func (reader *WasmReader) readValType() ValType {
	vt := reader.readByte()
	checkValType(vt)
	return vt
}
func checkValType(vt byte) {
	switch vt {
	case ValTypeI32:
	case ValTypeI64:
	case ValTypeF32:
	case ValTypeF64:
	default:
		panic(fmt.Errorf("invalid value type: %d", vt))
	}
}

// entity types
func (reader *WasmReader) readBlockType() []ValType {
	if b := reader.readByte(); b == NoVal {
		return nil
	} else {
		checkValType(b)
		return []ValType{b}
	}
}
func (reader *WasmReader) readFuncType() FuncType {
	ft := FuncType{
		Tag:         reader.readByte(),
		ParamTypes:  reader.readValTypes(),
		ResultTypes: reader.readValTypes(),
	}
	if ft.Tag != 0x60 {
		panic(fmt.Errorf("invalid functype tag: %d", ft.Tag))
	}
	return ft
}
func (reader *WasmReader) readTableType() TableType {
	tt := TableType{
		ElemType: reader.readByte(),
		Limits:   reader.readLimits(),
	}
	if tt.ElemType != FuncRef {
		panic(fmt.Errorf("invalid elemtype: %d", tt.ElemType))
	}
	return tt
}
func (reader *WasmReader) readGlobalType() GlobalType {
	gt := GlobalType{
		ValType: reader.readValType(),
		Mut:     reader.readByte(),
	}
	switch gt.Mut {
	case MutConst:
	case MutVar:
	default:
		panic(fmt.Errorf("invalid mutability: %d", gt.Mut))
	}
	return gt
}
func (reader *WasmReader) readLimits() Limits {
	limits := Limits{
		Tag: reader.readByte(),
		Min: reader.readVarU32(),
	}
	if limits.Tag == 1 {
		limits.Max = reader.readVarU32()
	}
	return limits
}

// indices
func (reader *WasmReader) readIndices() []uint32 {
	vec := make([]uint32, reader.readVarU32())
	for i := range vec {
		vec[i] = reader.readVarU32()
	}
	return vec
}

// expr & instruction

func (reader *WasmReader) readExpr() Expr {
	instrs, end := reader.readInstructions()
	if end != _End {
		panic(fmt.Errorf("invalid expr end: %d", end))
	}
	return instrs
}

func (reader *WasmReader) readInstructions() (instrs []Instruction, end byte) {
	for {
		instr := reader.readInstruction()
		if instr.Opcode == _Else || instr.Opcode == _End {
			end = instr.Opcode
			return
		}
		instrs = append(instrs, instr)
	}
}

func (reader *WasmReader) readInstruction() (instr Instruction) {
	instr.Opcode = reader.readByte()
	if opnames[instr.Opcode] == "" {
		panic(fmt.Errorf("undefined opcode: 0x%02x", instr.Opcode))
	}
	instr.Args = reader.readArgs(instr.Opcode)
	return
}

func (reader *WasmReader) readArgs(opcode byte) interface{} {
	switch opcode {
	case Block, Loop:
		return reader.readBlockArgs()
	case If:
		return reader.readIfArgs()
	case Br, BrIf:
		return reader.readVarU32() // label_idx
	case BrTable:
		return reader.readBrTableArgs()
	case Call:
		return reader.readVarU32() // func_idx
	case CallIndirect:
		return reader.readCallIndirectArgs()
	case LocalGet, LocalSet, LocalTee:
		return reader.readVarU32() // local_idx
	case GlobalGet, GlobalSet:
		return reader.readVarU32() // global_idx
	case MemorySize, MemoryGrow:
		return reader.readZero()
	case I32Const:
		return reader.readVarS32()
	case I64Const:
		return reader.readVarS64()
	case F32Const:
		return reader.readF32()
	case F64Const:
		return reader.readF64()
	default:
		if opcode >= I32Load && opcode <= I64Store32 {
			return reader.readMemArg()
		}
		return nil
	}
}

func (reader *WasmReader) readBlockArgs() (args BlockArgs) {
	var end byte
	args.RT = reader.readBlockType()
	args.Instrs, end = reader.readInstructions()
	if end != _End {
		panic(fmt.Errorf("invalid block end: %d", end))
	}
	return
}

func (reader *WasmReader) readIfArgs() (args IfArgs) {
	var end byte
	args.RT = reader.readBlockType()
	args.Instrs1, end = reader.readInstructions()
	if end == _Else {
		args.Instrs2, end = reader.readInstructions()
		if end != _End {
			panic(fmt.Errorf("invalid block end: %d", end))
		}
	}
	return
}

func (reader *WasmReader) readBrTableArgs() BrTableArgs {
	return BrTableArgs{
		Labels:  reader.readIndices(),
		Default: reader.readVarU32(),
	}
}

func (reader *WasmReader) readCallIndirectArgs() uint32 {
	typeIdx := reader.readVarU32()
	reader.readZero()
	return typeIdx
}

func (reader *WasmReader) readMemArg() MemArg {
	return MemArg{
		Align:  reader.readVarU32(),
		Offset: reader.readVarU32(),
	}
}

func (reader *WasmReader) readZero() byte {
	b := reader.readByte()
	if b != 0 {
		panic(fmt.Errorf("zero flag expected, got %d", b))
	}
	return 0
}
