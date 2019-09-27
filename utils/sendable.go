package utils

import (
	"github.com/sendgrid/rest"
	"github.com/sendgrid/sendgrid-go/helpers/mail"
	"github.com/stretchr/testify/mock"
)

var ExecSendEmail = execSend

type Sendable interface {
	Send(message *mail.SGMailV3) (*rest.Response, error)
}

func execSend(s Sendable, message *mail.SGMailV3) (*rest.Response, error) {
	return s.Send(message)
}

type TestEmail struct {
	mock.Mock
}

func (e *TestEmail) Send(message *mail.SGMailV3) (*rest.Response, error) {
	return nil, nil
}

func MockEmail() {
	mailMock := new(TestEmail)
	ExecSendEmail = func(s Sendable, message *mail.SGMailV3) (*rest.Response, error) {
		return mailMock.Send(message)
	}
}
