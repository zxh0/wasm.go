(module
  (type (func))
  (type (func))
  (type (func))
  (import "x" "f1" (func (type 100)))
)
(;;
err.wat:5:32: error: function type variable out of range: 100 (max 2)
  (import "x" "f1" (func (type 100)))
                               ^^^
;;)

;;------------------------------;;

(module
  (type (func))
  (type (func))
  (type (func))
  (func (type 100))
)
(;;
err.wat:5:15: error: function type variable out of range: 100 (max 2)
  (func (type 100))
              ^^^
;;)

;;------------------------------;;

(module
  (global i32 (i32.const 0))
  (global i32 (i32.const 0))
  (global i32 (i32.const 0))
  (func (global.get 100) (drop))
)
(;;
err.wat:5:21: error: global variable out of range: 100 (max 2)
  (func (global.get 100) (drop))
                    ^^^
;;)

;;------------------------------;;

(module
  (func
    (local i32 i32 i32)
    (local.get 100)
    (drop)
  )
)
(;;
err.wat:4:16: error: local variable out of range: 100 (max 2)
    (local.get 100)
               ^^^
;;)

;;------------------------------;;

(module
  (func
    (block $1 (br 100))
  )
)
(;;
err.wat:3:19: error: invalid depth: 100 (max 1)
    (block $1 (br 100))
                  ^^^
;;)

;;------------------------------;;

(module
  (type (func))
  (type (func))
  (type (func))
  (table 1 500 funcref)
  (func
    (call_indirect (type 100) (i32.const 1))
  )
)
(;;
err.wat:7:26: error: function type variable out of range: 100 (max 2)
    (call_indirect (type 100) (i32.const 1))
                         ^^^
;;)

;;------------------------------;;

(module
  (func) (func)
  (func (call 100))
)
(;;
err.wat:3:15: error: function variable out of range: 100 (max 2)
  (func (call 100))
              ^^^
;;)

;;------------------------------;;

(module
  (func) (func) (func)
  (export "x" (func 100))
)
(;;
err.wat:3:21: error: function variable out of range: 100 (max 2)
  (export "x" (func 100))
                    ^^^
;;)

;;------------------------------;;

(module
  (table 1 3 funcref)
  (export "x" (table 100))
)
(;;
err.wat:3:22: error: table variable out of range: 100 (max 0)
  (export "x" (table 100))
                     ^^^
;;)

;;------------------------------;;

(module
  (memory 1 16)
  (export "x" (memory 100))
)
(;;
err.wat:3:23: error: memory variable out of range: 100 (max 0)
  (export "x" (memory 100))
                      ^^^
;;)

;;------------------------------;;

(module
  (global i32 (i32.const 0))
  (global i32 (i32.const 0))
  (global i32 (i32.const 0))
  (export "x" (global 100))
)
(;;
err.wat:5:23: error: global variable out of range: 100 (max 2)
  (export "x" (global 100))
                      ^^^
;;)

;;------------------------------;;

(module
  (func) (func) (func)
  (start 100)
)
(;;
err.wat:3:10: error: function variable out of range: 100 (max 2)
  (start 100)
         ^^^
;;)

;;------------------------------;;

(module
  (func) (func) (func)
  (table $t 1 3 funcref)
  (elem $t (offset (i32.const 0)) 100)
)
(;;
err.wat:4:35: error: function variable out of range: 100 (max 2)
  (elem $t (offset (i32.const 0)) 100)
                                  ^^^
;;)

;;------------------------------;;

(module
  (func $f)
  (table $t 1 3 funcref)
  (elem 100 (offset (i32.const 0)) $f)
)
(;;
err.wat:4:9: error: table variable out of range: 100 (max 0)
  (elem 100 (offset (i32.const 0)) $f)
        ^^^
;;)

;;------------------------------;;

(module
  (memory 1 16)
  (data 100 (offset (i32.const 0)) "foo")
)
(;;
err.wat:3:9: error: memory variable out of range: 100 (max 0)
  (data 100 (offset (i32.const 0)) "foo")
        ^^^
;;)
