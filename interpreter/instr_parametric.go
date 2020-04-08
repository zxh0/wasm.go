package interpreter

func drop(vm *vm, _ interface{}) {
	vm.popU64()
}

func _select(vm *vm, _ interface{}) {
	v3 := vm.popU64()
	v2 := vm.popU64()
	v1 := vm.popU64()

	if v3 != 0 {
		vm.pushU64(v1)
	} else {
		vm.pushU64(v2)
	}
}
