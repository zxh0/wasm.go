package text

import (
	"fmt"
	"strings"

	"github.com/antlr/antlr4/runtime/Go/antlr"
)

var _ antlr.ErrorListener = (*ErrorListener)(nil)

/* ValidationError */

type ValidationError struct {
	msg   string
	token antlr.Token
}

func newVerificationError(format string, a ...interface{}) *ValidationError {
	return &ValidationError{
		msg: fmt.Sprintf(format, a...),
	}
}

func (e *ValidationError) Error() string {
	return e.msg
}

func (e *ValidationError) FillDetail(input antlr.CharStream) {
	e.msg = getErrDetail(e.msg, e.token, input)
}

/* SemanticError */

type SemanticError struct {
	msg   string
	token antlr.Token
}

func newSemanticError(format string, a ...interface{}) *SemanticError {
	return &SemanticError{
		msg: fmt.Sprintf(format, a...),
	}
}

func (e *SemanticError) Error() string {
	return e.msg
}

func (e *SemanticError) FillDetail(input antlr.CharStream) {
	e.msg = getErrDetail(e.msg, e.token, input)
}

/* SyntaxError */

type SyntaxError struct {
	msg    string
	line   int
	column int
	token  antlr.Token
}

func (err SyntaxError) Error() string {
	return err.msg
}

/* SyntaxErrors */

type SyntaxErrors []SyntaxError

func (errs SyntaxErrors) FillDetail(input antlr.CharStream) {
	for i := range errs {
		errs[i].msg = getErrDetail(errs[i].msg, errs[i].token, input)
	}
}

func (errs SyntaxErrors) Error() string {
	s := ""
	for i, err := range errs {
		s += err.Error()
		if i < len(errs)-1 {
			s += "\n"
		}
	}
	return s
}

/* helpers */

func getErrDetail(msg string, token antlr.Token, input antlr.CharStream) string {
	msg = fmt.Sprintf("%s:%d:%d: error: %s",
		input.GetSourceName(), token.GetLine(), token.GetColumn()+1, msg)
	allText := fmt.Sprintf("%s", input) // TODO
	errLine := strings.Split(allText, "\n")[token.GetLine()-1]
	underline := getUnderline(token.GetColumn(), token)
	return msg + "\n" + errLine + "\n" + underline
}

func getUnderline(column int, token antlr.Token) string {
	if token != nil {
		return strings.Repeat(" ", token.GetColumn()) +
			strings.Repeat("^", len(token.GetText()))
	}

	return strings.Repeat(" ", column) + "^"
}
