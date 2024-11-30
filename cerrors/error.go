package cerrors

import (
	"fmt"
)

type (
	Code int64

	Error struct {
		Code        Code
		UserMessage string
		Origin      error
	}
)

func (e *Error) Error() string {
	var originErrMessage string
	if e.Origin != nil {
		originErrMessage = e.Origin.Error()
	}
	return fmt.Sprintf("internal code: %d; origin message: %s; user message: %s", e.Code, originErrMessage, e.UserMessage)
}
