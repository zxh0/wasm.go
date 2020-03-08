package interpreter

func drop(vm *vm, _ interface{}) {
	vm.popU64()
}

func _select(vm *vm, _ interface{}) {
	c := vm.popU64()
	b := vm.popU64()
	a := vm.popU64()

	if c != 0 {
		vm.pushU64(a)
	} else {
		vm.pushU64(b)
	}
}
