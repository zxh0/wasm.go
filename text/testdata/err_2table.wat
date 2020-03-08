(module
  (import "x" "t1" (table 1 2 funcref))
  (import "x" "t2" (table 1 2 funcref))
)
(;;
err.wat:3:4: error: only one table allowed
  (import "x" "t2" (table 1 2 funcref))
   ^^^^^^
;;)

;;------------------------------;;

(module
  (import "x" "t1" (table 1 2 funcref))
  (table (import "x" "t2") 1 8 funcref)
)
(;;
err.wat:3:4: error: only one table allowed
  (table (import "x" "t2") 1 8 funcref)
   ^^^^^
;;)

;;------------------------------;;

(module
  (import "x" "t1" (table 1 2 funcref))
  (table 1 3 funcref)
)
(;;
err.wat:3:4: error: only one table allowed
  (table 1 3 funcref)
   ^^^^^
;;)

;;------------------------------;;

(module
  (table (import "x" "t") 1 8 funcref)
  (table 1 3 funcref)
)
(;;
err.wat:3:4: error: only one table allowed
  (table 1 3 funcref)
   ^^^^^
;;)

;;------------------------------;;

(module
  (table 1 3 funcref)
  (table 1 3 funcref)
)
(;;
err.wat:3:4: error: only one table allowed
  (table 1 3 funcref)
   ^^^^^
;;)