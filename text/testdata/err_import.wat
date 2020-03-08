(module
  (func)
  (import "x" "y" (func))
)
(;;
err.wat:3:4: error: imports must occur before all non-import definitions
  (import "x" "y" (func))
   ^^^^^^
;;)

;;------------------------------;;

(module
  (func)
  (import "x" "y" (table 1 2 funcref))
)
(;;
err.wat:3:4: error: imports must occur before all non-import definitions
  (import "x" "y" (table 1 2 funcref))
   ^^^^^^
;;)

;;------------------------------;;

(module
  (func)
  (import "x" "y" (memory 1 8))
)
(;;
err.wat:3:4: error: imports must occur before all non-import definitions
  (import "x" "y" (memory 1 8))
   ^^^^^^
;;)

;;------------------------------;;

(module
  (func)
  (import "x" "y" (global i32))
)
(;;
err.wat:3:4: error: imports must occur before all non-import definitions
  (import "x" "y" (global i32))
   ^^^^^^
;;)

;;------------------------------;;

(module
  (table 1 3 funcref)
  (func (import "x" "y"))
)
(;;
err.wat:3:10: error: imports must occur before all non-import definitions
  (func (import "x" "y"))
         ^^^^^^
;;)

;;------------------------------;;

(module
  (func)
  (table (import "x" "y") 1 8 funcref)
)
(;;
err.wat:3:11: error: imports must occur before all non-import definitions
  (table (import "x" "y") 1 8 funcref)
          ^^^^^^
;;)

;;------------------------------;;

(module
  (global $g i32)
  (memory (import "x" "y") 1 8)
)
(;;
err.wat:3:12: error: imports must occur before all non-import definitions
  (memory (import "x" "y") 1 8)
           ^^^^^^
;;)

;;------------------------------;;

(module
  (memory $m 1 8)
  (global (import "x" "y") f32)
)
(;;
err.wat:3:12: error: imports must occur before all non-import definitions
  (global (import "x" "y") f32)
           ^^^^^^
;;)