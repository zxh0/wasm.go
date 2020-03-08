(module
  (type $ft1 (func (param i32)))
  (func (type $ft1) (param i64))
)
(;;
err.wat:3:10: error: type mismatch
  (func (type $ft1) (param i64))
         ^^^^
;;)

;;------------------------------;;

(module
  (type $ft1 (func (param i32)))
  (func $f1 (type $ft1))
  (table funcref (elem $f1))
  (func
    (i64.const 1)
    (call_indirect (type $ft1) (param i64))
  )
)
(;;
err.wat:7:21: error: type mismatch in call_indirect
    (call_indirect (type $ft1) (param i64))
                    ^^^^
;;)
