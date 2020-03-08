grammar WAT;

module     : watModule EOF ;
watModule  : '(' 'module' NAME? moduleField* ')' ;

moduleField: typeDef
           | import_
           | func_
           | table
           | memory
           | global
           | export
           | start
           | elem
           | data
           ;

// Module Fields
typeDef    : '(' 'type' NAME? '(' 'func' funcType ')' ')'
           ;
import_    : '(' 'import' STRING STRING importDesc ')' ;
importDesc : '(' kind='func'   NAME? typeUse ')'
           | '(' kind='table'  NAME? tableType ')'
           | '(' kind='memory' NAME? memoryType ')'
           | '(' kind='global' NAME? globalType ')'
           ;
func_      : '(' 'func' NAME? embeddedEx typeUse funcLocal* expr ')'
           | '(' 'func' NAME? embeddedEx embeddedIm typeUse ')' ;
funcLocal  : '(' 'local' valType* ')'
           | '(' 'local' NAME valType ')'
           ;
table      : '(' 'table' NAME? embeddedEx tableType ')'
           | '(' 'table' NAME? embeddedEx embeddedIm tableType ')'
           | '(' 'table' NAME? embeddedEx elemType '(' 'elem' funcVars ')' ')'
           ;
memory     : '(' 'memory' NAME? embeddedEx memoryType ')'
           | '(' 'memory' NAME? embeddedEx embeddedIm memoryType ')'
           | '(' 'memory' NAME? embeddedEx '(' 'data' STRING* ')' ')'
           ;
global     : '(' 'global' NAME? embeddedEx globalType expr ')'
           | '(' 'global' NAME? embeddedEx embeddedIm globalType ')'
           ;
export     : '(' 'export' STRING exportDesc ')' ;
exportDesc : '(' kind='func'   variable ')'
           | '(' kind='table'  variable ')'
           | '(' kind='memory' variable ')'
           | '(' kind='global' variable ')'
           ;
start      : '(' 'start' variable ')'
           ;
elem       : '(' 'elem' variable? '(' 'offset' expr ')' funcVars ')'
           | '(' 'elem' variable?              expr     funcVars ')'
           ;
data       : '(' 'data' variable? '(' 'offset' expr ')' STRING* ')'
           | '(' 'data' variable?              expr     STRING* ')'
           ;

embeddedIm : '(' 'import' STRING STRING ')' ;
embeddedEx : ('(' 'export' STRING ')')* ;
typeUse    : ('(' 'type' variable ')')? funcType ;
funcVars   : variable* ;

// Types
valType    : VAL_TYPE ;
blockType  : result? ;
globalType : valType | '(' 'mut' valType ')' ;
memoryType : limits ;
tableType  : limits elemType ;
elemType   : 'funcref' ;
limits     : nat nat? ;

funcType   : param* result* ;
param      : '(' 'param' valType* ')'
           | '(' 'param' NAME valType ')' ;
result     : '(' 'result' valType* ')' ;

// Instructions
expr       : instr*
           ;
instr      : plainInstr
           | blockInstr
           | foldedInstr
           ;
foldedInstr: '(' plainInstr foldedInstr* ')'
           | '(' op='block' label=NAME? blockType expr ')'
           | '(' op='loop'  label=NAME? blockType expr ')'
           | '(' op='if'    label=NAME? blockType foldedInstr*
                 '(' 'then' expr ')' ('(' 'else' expr ')')? ')'
           ;
blockInstr : op='block' label=NAME? blockType expr 'end' l2=NAME?
           | op='loop'  label=NAME? blockType expr 'end' l2=NAME?
           | op='if'    label=NAME? blockType expr
                           ('else' l1=NAME? expr)? 'end' l2=NAME?
           ;
plainInstr : op='unreachable'
           | op='nop'
           | op='br' variable
           | op='br_if' variable
           | op='br_table' variable+
           | op='return'
           | op='call' variable
           | op='call_indirect' typeUse
           | op='drop'
           | op='select'
           | op=VAR_OPS variable
           | op=MEM_OPS memArg
           | op='memory.size'
           | op='memory.grow'
           | op=NUM_OPS
           | constInstr
           ;
constInstr : op=CST_OPS value
           ;

memArg     : ('offset' '=' offset=nat)? ('align' '=' align=nat)? ;

// Value & Variable
nat        : NAT ;
value      : NAT | INT | FLOAT ;
variable   : NAT | NAME ;

// Lexer

VAL_TYPE: ValType ;

NAME    : '$' NameChar+ ;
STRING  : '"' (StrChar|StrEsc)* '"' ;
FLOAT   : Sign? Num '.' Num? ([eE] Sign? Num)?
        | Sign? Num [eE] Sign? Num
        | Sign? '0x' HexNum '.' HexNum? ([pP] Sign? Num)?
        | Sign? '0x' HexNum [pP] Sign? Num
        | Sign? 'inf'
        | Sign? 'nan'
        | Sign? 'nan:0x' HexNum
        ;
NAT     : Num | '0x' HexNum;
INT     : Sign? NAT ;

// Opcodes

VAR_OPS : 'local.get'
        | 'local.set'
        | 'local.tee'
        | 'global.get'
        | 'global.set'
        ;
MEM_OPS : ValType '.load'
        | IntType '.load8' OpSign
        | IntType '.load16' OpSign
        | 'i64'   '.load32' OpSign
        | ValType '.store'
        | IntType '.store8'
        | IntType '.store16'
        | 'i64'   '.store32'
        ;
CST_OPS : ValType '.const' ;
NUM_OPS : IntType '.' IntArith
        | IntType '.' IntRel
        | FloatType '.' FloatArith
        | FloatType '.' FloatRel
        | IntType '.trunc_' FloatType OpSign
        | FloatType '.convert_' IntType OpSign
        | 'i32.wrap_i64'
        | 'i64.extend_i32_s'
        | 'i64.extend_i32_u'
        | 'f32.demote_f64'
        | 'f64.promote_f32'
        | 'i32.reinterpret_f32'
        | 'i64.reinterpret_f64'
        | 'f32.reinterpret_i32'
        | 'f64.reinterpret_i64'
        ;

// Fragments

fragment Sign       : '+' | '-' ;
fragment Digit      : [0-9] ;
fragment HexDigit   : [0-9a-fA-F] ;
fragment Num        : Digit ('_'? Digit)* ;
fragment HexNum     : HexDigit ('_'? HexDigit)* ;
fragment NameChar   : [a-zA-Z0-9_.+\-*/\\^~=<>!?@#$%&|:'`] ;
fragment StrChar    : ~["\\\u0000-\u001f\u007f] ;
fragment StrEsc     : '\\t' | '\\n' | '\\r' | '\\"' | '\\\\'
                    | '\\' HexDigit HexDigit // hexadecimal
                    | '\\u{' HexDigit+ '}'   // unicode
                    ;

fragment OpSign     : '_s' | '_u' ;
fragment IntType    : 'i32' | 'i64' ;
fragment FloatType  : 'f32' | 'f64' ;
fragment ValType    : IntType | FloatType ;

fragment IntArith   : 'clz' | 'ctz' | 'popcnt'
                    | 'add' | 'sub' | 'mul' | 'div_s' | 'div_u' | 'rem_s' | 'rem_u'
                    | 'and' | 'or' | 'xor' | 'shl' | 'shr_s' | 'shr_u' | 'rotl' | 'rotr'
                    ;
fragment FloatArith : 'abs' | 'neg' | 'ceil' | 'floor' | 'trunc' | 'nearest' | 'sqrt'
                    | 'add' | 'sub'  | 'mul' | 'div' | 'min' | 'max'
                    | 'copysign'
                    ;
fragment IntRel     : 'eqz' | 'eq' | 'ne'
                    | 'lt_s' | 'lt_u'
                    | 'gt_s' | 'gt_u'
                    | 'le_s' | 'le_u'
                    | 'ge_s' | 'ge_u'
                    ;
fragment FloatRel   : 'eq' | 'ne' | 'lt' | 'gt' | 'le' | 'ge' ;

// Whitespace and Comments

WS            : [ \t\r\n]+ -> skip ;
LINE_COMMENT  : ';;' ~[\n]* -> skip ;
BLOCK_COMMENT : '(;' (BLOCK_COMMENT|.)*? ';)' -> skip ;
