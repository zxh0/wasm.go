grammar WAST;
import WAT;

// https://github.com/WebAssembly/spec/tree/master/interpreter#scripts

script     : cmd* EOF ;

cmd        : wastModule
           | '(' 'register' STRING NAME? ')'
           | action_
           | assertion
           | meta
           ;

wastModule : watModule
           | '(' 'module' NAME? kind='binary' STRING* ')'
           | '(' 'module' NAME? kind='quote'  STRING* ')'
           ;

action_    : '(' kind='invoke' NAME? STRING expr ')'
           | '(' kind='get'    NAME? STRING ')'
           ;

assertion  : '(' kind='assert_return'     action_ expected* ')'
           | '(' kind='assert_trap'       action_    STRING ')'
           | '(' kind='assert_exhaustion' action_    STRING ')'
           | '(' kind='assert_malformed'  wastModule STRING ')'
           | '(' kind='assert_invalid'    wastModule STRING ')'
           | '(' kind='assert_unlinkable' wastModule STRING ')'
           | '(' kind='assert_trap'       wastModule STRING ')'
           ;
expected   : '(' constInstr ')'
           | '(' op=CST_OPS nan='nan:canonical' ')'
           | '(' op=CST_OPS nan='nan:arithmetic' ')'
           ;

meta       : '(' 'script' NAME? script ')'
           | '(' 'input'  NAME? STRING ')'
           | '(' 'output' NAME? STRING? ')'
           ;
