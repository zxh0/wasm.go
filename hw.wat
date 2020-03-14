(module
  (import "env" "print_char" (func $print_char (param i32)))
  (memory (data "Hello, World!\n"))
  (func $main (export "main")
    (local $a i32)

    (loop
      (call $print_char (i32.load8_u (local.get $a)))
      (local.set $a (i32.add (local.get $a) (i32.const 1)))
      (br_if 0 (i32.lt_u (local.get $a) (i32.const 14)))
    )
  )
)
