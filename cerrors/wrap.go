package cerrors

func NewErrorWithUserMessage(code Code, err error, userMessage string) *Error {
	return &Error{
		Code:        code,
		UserMessage: userMessage,
		Origin:      err,
	}
}
