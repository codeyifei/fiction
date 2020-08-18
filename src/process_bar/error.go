package process_bar

import (
	"fmt"
)

type Error struct {
	message string
}

func NewError(message string) (err error) {
	err = &Error{message: message}
	return
}

func (e *Error) Error() (message string) {
	message = fmt.Sprint(e.message)
	return
}


