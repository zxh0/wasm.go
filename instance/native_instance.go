package instance

var _ Module = (*NativeInstance)(nil)

type NativeInstance struct {
	exported map[string]interface{}
}

func NewNativeInstance() *NativeInstance {
	return &NativeInstance{
		exported: map[string]interface{}{},
	}
}

func (n *NativeInstance) RegisterFunc(nameAndSig string, f GoFunc) {
	name, sig := parseNameAndSig(nameAndSig)
	n.exported[name] = nativeFunction{t: sig, f: f}
}

func (n *NativeInstance) Register(name string, x interface{}) {
	n.exported[name] = x
}

func (n *NativeInstance) GetMember(name string) interface{} {
	return n.exported[name]
}

func (n *NativeInstance) InvokeFunc(name string, args ...WasmVal) ([]WasmVal, error) {
	return n.exported[name].(Function).Call(args...) // TODO
}

func (n *NativeInstance) GetGlobalVal(name string) (WasmVal, error) {
	return n.exported[name].(Global).Get(), nil // TODO
}
func (n *NativeInstance) SetGlobalVal(name string, val WasmVal) error {
	n.exported[name].(Global).Set(val) // TODO
	return nil
}
