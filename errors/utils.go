package errors

import (
	"encoding/json"
	"fmt"
	// "strings"

	"go.uber.org/zap"
)

type NetError interface {
	Error() string
	Timeout() bool
	Temporary() bool
}

func NewNetError(err NetError) Error {
	return WrappedError(ErrorSystemNet, err)
}

func NamedScope(name string) zap.Field {
	return zap.String("scope", name)
}

func ParseError(jsonStr string) Error {
	err := &baseError{}
	e := json.Unmarshal([]byte(jsonStr), err)
	if e != nil {
		err.Code = ErrorSystemRPC
		err.Message = fmt.Sprintf("无法解析错误消息[%s],%#v", jsonStr, e)
	}
	return err
}

func ConverUnknowError(err interface{}) Error {
	if err == nil {
		return nil
	}
	if e, ok := err.(error); ok {
		return ConverError(e)
	}
	return SystemError("未知异常")
}

func ConverError(err error) Error {
	if err == nil {
		return nil
	}
	if IsBaseError(err) {
		return err.(Error)
	}
	// if strings.HasPrefix(err.Error(), "context ") || strings.HasPrefix(err.Error(), "rpc error:") {
	// 	return SystemError(err.Error())
	// }
	return WrappedSystemError(err)
}
