package errors

type CodedError struct {
	Code    int
	Message string
}

func (e CodedError) Error() string {
	return e.Message
}

func New(code int, message string) CodedError {
	return CodedError{Code: code, Message: message}
}
