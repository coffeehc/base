package errors

import errors1 "errors"

func BuildError(errorCode int64, message string) Error {
	return &baseError{
		Code:    errorCode,
		Message: message,
		e:       errors1.New(message),
	}
}

func SystemError(message string) Error {
	return BuildError(ErrorSystem, message)
}

func MessageError(message string) Error {
	return BuildError(ErrorMessage, message)
}

func WrappedError(errorCode int64, err error) Error {
	return &baseError{
		Code:    errorCode,
		Message: err.Error(),
		e:       err,
	}
}

func WrappedSystemError(err error) Error {
	if err == nil {
		return nil
	}
	return WrappedError(ErrorSystem, err)
}

func WrappedMessageError(err error) Error {
	if err == nil {
		return nil
	}
	return WrappedError(ErrorMessage, err)
}
