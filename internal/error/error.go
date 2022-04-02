package errors

import (
	"fmt"
	"runtime"
)

type Error interface {
	Error() string
	Warp(err error) error
	WarpStr(str string) error
}

type protocolError struct {
	Message string
	Caller  []string
}

func (p *protocolError) Error() string {
	return p.Message
}

func (p *protocolError) Warp(err error) error {
	return fmt.Errorf("%s:%w", err, p)
}

func (p *protocolError) WarpStr(str string) error {
	return fmt.Errorf("%s:%w", str, p)
}

func NewProtocolError(message string) *protocolError {
	return &protocolError{
		Message: "protocol error",
		Caller:  []string{call(1)},
	}
}

func call(skip int) string {
	pc, file, line, ok := runtime.Caller(skip)
	pcName := runtime.FuncForPC(pc).Name() //获取函数名
	return fmt.Sprintf("%v   %s   %d   %t   %s", pc, file, line, ok, pcName)
}
