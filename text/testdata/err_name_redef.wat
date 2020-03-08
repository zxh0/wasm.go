(module
  (import "x" "f" (func $f))
  (import "x" "f" (func $f))
)
(;;
err.wat:3:25: error: redefinition of function "$f"
  (import "x" "f" (func $f))
                        ^^
;;)

;;------------------------------;;

(module
  (import "x" "f" (global $g i32))
  (import "x" "f" (global $g i32))
)
(;;
err.wat:3:27: error: redefinition of global "$g"
  (import "x" "f" (global $g i32))
                          ^^
;;)

;;------------------------------;;

(module
  (import "x" "f" (global $g i32))
  (global $g i32)
)
(;;
err.wat:3:11: error: redefinition of global "$g"
  (global $g i32)
          ^^
;;)

;;------------------------------;;

(module
  (import "x" "f" (func $f))
  (func $f)
)
(;;
err.wat:3:9: error: redefinition of function "$f"
  (func $f)
        ^^
;;)

;;------------------------------;;

(module
  (type $t (func (param i32)))
  (type $t (func (param i32)))
)
(;;
err.wat:3:9: error: redefinition of function type "$t"
  (type $t (func (param i32)))
        ^^
;;)

;;------------------------------;;

(module
  (func $f)
  (func $f)
)
(;;
err.wat:3:9: error: redefinition of function "$f"
  (func $f)
        ^^
;;)

;;------------------------------;;

(module
  (table $t 1 3 funcref)
  (table $t 1 3 funcref)
)
(;;
err.wat:3:10: error: redefinition of table "$t"
  (table $t 1 3 funcref)
         ^^
;;)

;;------------------------------;;

(module
  (memory $m 1 8)
  (memory $m 1 8)
)
(;;
err.wat:3:11: error: redefinition of memory "$m"
  (memory $m 1 8)
          ^^
;;)

;;------------------------------;;

(module
  (global $g i32)
  (global $g i32)
)
(;;
err.wat:3:11: error: redefinition of global "$g"
  (global $g i32)
          ^^
;;)

;;------------------------------;;

(module
  (func (param $a i32) (param $a i32))
)
(;;
err.wat:2:31: error: redefinition of parameter "$a"
  (func (param $a i32) (param $a i32))
                              ^^
;;)

;;------------------------------;;

(module
  (func (param $a i32) (local $a i32))
)
(;;
err.wat:2:31: error: redefinition of local "$a"
  (func (param $a i32) (local $a i32))
                              ^^
;;)

;;------------------------------;;

(module
  (func (param $a i32)
    (local $b i32)
    (local $b i32)
  )
)
(;;
err.wat:4:12: error: redefinition of local "$b"
    (local $b i32)
           ^^
;;)
