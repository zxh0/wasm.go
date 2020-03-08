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

