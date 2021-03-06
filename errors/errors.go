package errors

import (
	"encoding/json"

	"go.uber.org/zap"
)

// Error 基础的错误接口
type Error interface {
	error
	GetCode() int64
	GetFields(fields ...zap.Field) []zap.Field
	GetFieldsWithCause(fields ...zap.Field) []zap.Field
	FormatRPCError() string
	Is(Error) bool
	ToError() error
}

// BaseError Error 接口的实现,可 json 序列化
type baseError struct {
	Code    int64  `json:"code"`
	Message string `json:"msg"`
	e       error
}

func (err *baseError) ToError() error {
	return err.e
}

func (err *baseError) Is(err2 Error) bool {
	return err.Code == err2.GetCode()
}

func (err *baseError) FormatRPCError() string {
	json, _ := json.Marshal(err)
	return string(json)
}

func (err *baseError) Error() string {
	return err.Message
}

func (err *baseError) GetCode() int64 {
	return err.Code
}

func (err *baseError) GetFields(fields ...zap.Field) []zap.Field {
	if len(fields) == 0 {
		return []zap.Field{zap.Int64("errCode", err.GetCode())}
	}
	return append([]zap.Field{zap.Int64("errCode", err.GetCode())}, fields...)
}
func (err *baseError) GetFieldsWithCause(fields ...zap.Field) []zap.Field {
	if len(fields) == 0 {
		return []zap.Field{zap.Int64("errCode", err.GetCode()), zap.String("error", err.Message)}
	}
	return append([]zap.Field{zap.Int64("errCode", err.GetCode()), zap.String("error", err.Message)}, fields...)
}

// ParseErrorFromJSON 从 Jons数据解析出 Error 对象
func ParseErrorFromJSON(data []byte) Error {
	err := &baseError{}
	e := json.Unmarshal(data, err)
	if e != nil {
		return nil
	}
	return err
}

func ErrorToJson(err Error) string {
	data, _ := json.Marshal(err)
	return string(data)
}
