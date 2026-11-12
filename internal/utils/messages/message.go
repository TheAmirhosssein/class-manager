package messages

import (
	"fmt"

	"github.com/TheAmirhosssein/class-manager/pkg/errors"
)

type Message struct {
	Code        int
	Description string
	Error       error
}

func (m Message) HttpCode() int {
	if m.Code >= 1000 {
		return m.Code / 1000
	}
	return m.Code
}

func NewMessage(code int, description string) Message {
	return Message{Code: code, Description: description}
}

func NewMessageByError(code int, description string) Message {
	return Message{
		Code:        code,
		Description: description,
		Error:       errors.New(code, description),
	}
}

var (
	BadRequest  = NewMessageByError(400_000, "درخواست نامعتبر است")
	ServerError = NewMessageByError(500_000, "خطای داخلی سرور")
)

func SharedServerError() Message {
	return ServerError
}

func OK() Message {
	return NewMessage(200_000, "")
}

func (m Message) String() string {
	if m.Description != "" {
		return m.Description
	}
	if m.Error != nil {
		return m.Error.Error()
	}
	return fmt.Sprintf("code %d", m.Code)
}
