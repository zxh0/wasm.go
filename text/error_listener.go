package text

import "github.com/antlr/antlr4/runtime/Go/antlr"

type ErrorListener struct {
	antlr.DefaultErrorListener
	errors  SyntaxErrors
	allText []string
}

func (listener *ErrorListener) GetErrors(input antlr.CharStream) error {
	if len(listener.errors) > 0 {
		errs := listener.errors
		errs.FillDetail(input)
		return errs
	} else {
		return nil
	}
}

func (listener *ErrorListener) SyntaxError(recognizer antlr.Recognizer,
	offendingSymbol interface{}, line, column int, msg string,
	e antlr.RecognitionException) {

	err := SyntaxError{
		msg:    msg,
		line:   line,
		column: column,
		token:  offendingSymbol.(antlr.Token),
	}
	listener.errors = append(listener.errors, err)
}
