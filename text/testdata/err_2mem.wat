(module
  (import "x" "m1" (memory 1 8))
  (import "x" "m2" (memory 1 8))
)
(;;
err.wat:3:4: error: only one memory block allowed
  (import "x" "m2" (memory 1 8))
   ^^^^^^
;;)

;;------------------------------;;

(module
  (import "x" "m1" (memory 1 8))
  (memory (import "x" "m2") 1 8)
)
(;;
err.wat:3:4: error: only one memory block allowed
  (memory (import "x" "m2") 1 8)
   ^^^^^^
;;)

;;------------------------------;;

(module
  (import "x" "m1" (memory 1 8))
  (memory 1 8)
)
(;;
err.wat:3:4: error: only one memory block allowed
  (memory 1 8)
   ^^^^^^
;;)

;;------------------------------;;

(module
  (memory (import "x" "m") 1 8)
  (memory 1 8)
)
(;;
err.wat:3:4: error: only one memory block allowed
  (memory 1 8)
   ^^^^^^
;;)

;;------------------------------;;

(module
  (memory 1 8)
  (memory 1 8)
)
(;;
err.wat:3:4: error: only one memory block allowed
  (memory 1 8)
   ^^^^^^
;;)
