package text

type scriptBuilder struct {
	script *Script
}

func newScriptBuilder() *scriptBuilder {
	return &scriptBuilder{script: &Script{}}
}

func (b *scriptBuilder) addCmd(cmd interface{}) {
	b.script.Cmds = append(b.script.Cmds, cmd)
}
