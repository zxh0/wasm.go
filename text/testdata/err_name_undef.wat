(module
  (import "x" "f1" (func (type $ft1)))
)
(;;
err.wat:2:32: error: undefined function type variable "$ft1"
  (import "x" "f1" (func (type $ft1)))
                               ^^^^
;;)

;;------------------------------;;

(module
  (func (type $ft))
)
(;;
err.wat:2:15: error: undefined function type variable "$ft"
  (func (type $ft))
              ^^^
;;)

;;------------------------------;;

(module
  (func (global.get $x))
)
(;;
err.wat:2:21: error: undefined global variable "$x"
  (func (global.get $x))
                    ^^
;;)

;;------------------------------;;

(module
  (func (local.get $x))
)
(;;
err.wat:2:20: error: undefined local variable "$x"
  (func (local.get $x))
                   ^^
;;)

;;------------------------------;;

(module
  (func
    (block $1 (br $2))
  )
)
(;;
err.wat:3:19: error: undefined label variable "$2"
    (block $1 (br $2))
                  ^^
;;)

;;------------------------------;;

(module
  (type $ft1 (func))
  (func
    (call_indirect (type $ft2))
  )
)
(;;
err.wat:4:26: error: undefined function type variable "$ft2"
    (call_indirect (type $ft2))
                         ^^^^
;;)

;;------------------------------;;

(module
  (func (call $f))
)
(;;
err.wat:2:15: error: undefined function variable "$f"
  (func (call $f))
              ^^
;;)

;;------------------------------;;

(module
  (export "x" (func $f))
)
(;;
err.wat:2:21: error: undefined function variable "$f"
  (export "x" (func $f))
                    ^^
;;)

;;------------------------------;;

(module
  (export "x" (table $t))
)
(;;
err.wat:2:22: error: undefined table variable "$t"
  (export "x" (table $t))
                     ^^
;;)

;;------------------------------;;

(module
  (export "x" (memory $m))
)
(;;
err.wat:2:23: error: undefined memory variable "$m"
  (export "x" (memory $m))
                      ^^
;;)

;;------------------------------;;

(module
  (export "x" (global $b))
)
(;;
err.wat:2:23: error: undefined global variable "$b"
  (export "x" (global $b))
                      ^^
;;)

;;------------------------------;;

(module
  (start $f)
)
(;;
err.wat:2:10: error: undefined function variable "$f"
  (start $f)
         ^^
;;)

;;------------------------------;;

(module
  (table $t 1 3 funcref)
  (elem $t (offset (i32.const 0)) $f)
)
(;;
err.wat:3:35: error: undefined function variable "$f"
  (elem $t (offset (i32.const 0)) $f)
                                  ^^
;;)

;;------------------------------;;

(module
  (func $f)
  (elem $t (offset (i32.const 0)) $f)
)
(;;
err.wat:3:9: error: undefined table variable "$t"
  (elem $t (offset (i32.const 0)) $f)
        ^^
;;)

;;------------------------------;;

(module
  (data $m (offset (i32.const 0)) "foo")
)
(;;
err.wat:2:9: error: undefined memory variable "$m"
  (data $m (offset (i32.const 0)) "foo")
        ^^
;;)
