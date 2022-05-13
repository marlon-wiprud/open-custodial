package _err

import (
	"encoding/json"
	"errors"
	"fmt"
)

const Hello = "hello"

type Err struct {
	error          // develop facing
	Message string // user facing
}

func NewError(err error, message string) Err {
	return Err{err, message}
}

func (e Err) Error() string {
	return e.error.Error()
}

func (e Err) Unwrap() error {
	return e.error
}

type ErrorRender struct {
	Details string `json:"details"`
	Message string `json:"message"`
}

func (e Err) MarshalJSON() ([]byte, error) {
	return json.Marshal(ErrorRender{
		Details: e.Error(),
		Message: e.Message,
	})
}

type DuplicateLabel struct{ Err }

func NewDuplicateLabelErr(label string) DuplicateLabel {
	message := fmt.Sprintf("duplicate label %s", label)
	return DuplicateLabel{Err{error: errors.New(message), Message: message}}
}

type BadForm struct{ Err }

func NewBadFormErr(e error) BadForm {
	return BadForm{Err{error: e, Message: "invalid request form"}}
}
